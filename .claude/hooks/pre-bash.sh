#!/usr/bin/env bash
# Pre-bash hook: block destructive commands before execution.
# Receives tool input as JSON on stdin.
# Exit 0 = allow, exit 2 = block with message.
set -euo pipefail

INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.command // empty')

if [ -z "$COMMAND" ]; then
  exit 0
fi

# --- 1. Block destructive file operations ---
if echo "$COMMAND" | grep -qE 'rm\s+(-[a-zA-Z]*f[a-zA-Z]*\s+|.*-rf\s+|.*--force)(/|~|\.\.|/etc|/usr|/var|/home|\$HOME)'; then
  echo "BLOCKED: destructive rm targeting sensitive path. Review the path and use a more targeted command."
  exit 2
fi

# --- 2. Block force push to main/master ---
if echo "$COMMAND" | grep -qE 'git\s+push\s+.*--force.*\s+(main|master)' || \
   echo "$COMMAND" | grep -qE 'git\s+push\s+-f\s+.*\s+(main|master)'; then
  echo "BLOCKED: force push to main/master. This rewrites shared history."
  exit 2
fi

# --- 3. Block dangerous git operations ---
if echo "$COMMAND" | grep -qE 'git\s+reset\s+--hard'; then
  echo "BLOCKED: git reset --hard discards uncommitted work. Use git stash or git reset --soft."
  exit 2
fi

if echo "$COMMAND" | grep -qE 'git\s+clean\s+-[a-zA-Z]*f'; then
  echo "BLOCKED: git clean -f deletes untracked files permanently. Review what would be deleted with git clean -n first."
  exit 2
fi

# --- 4. Block database destruction ---
if echo "$COMMAND" | grep -qiE '(DROP\s+DATABASE|DROP\s+SCHEMA)'; then
  echo "BLOCKED: DROP DATABASE/SCHEMA via shell. This is irreversible. Use a migration tool."
  exit 2
fi

# --- 5. Block credential exposure ---
if echo "$COMMAND" | grep -qE '(cat|echo|printf|less|head|tail)\s+.*(\.env|credentials|\.pem|\.key|id_rsa)'; then
  echo "BLOCKED: command may expose secrets. Don't cat credential files in the terminal."
  exit 2
fi

# --- 6. Block sudo operations ---
if echo "$COMMAND" | grep -qE '^\s*sudo\s+'; then
  echo "BLOCKED: sudo escalation not allowed in this environment."
  exit 2
fi

# --- 7. Block kill -9 on system processes ---
if echo "$COMMAND" | grep -qE 'kill\s+-9\s+(1|$$)\b'; then
  echo "BLOCKED: kill -9 on PID 1 or parent process. This would crash the system."
  exit 2
fi

# --- 8. Block curl/wget piped to shell ---
if echo "$COMMAND" | grep -qE '(curl|wget)\s+.*\|\s*(bash|sh|zsh)'; then
  echo "BLOCKED: piping remote content to shell. Download first, review, then execute."
  exit 2
fi

exit 0
