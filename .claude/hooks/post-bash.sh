#!/usr/bin/env bash
# Post-bash hook: surface warnings from command output.
# Receives tool input as JSON on stdin (includes command and stdout/stderr).
set -euo pipefail

INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.command // empty')
STDOUT=$(echo "$INPUT" | jq -r '.stdout // empty')
STDERR=$(echo "$INPUT" | jq -r '.stderr // empty')
OUTPUT="$STDOUT$STDERR"

if [ -z "$COMMAND" ]; then
  exit 0
fi

# --- 1. Surface deprecation warnings ---
if echo "$OUTPUT" | grep -qiE '(deprecat(ed|ion)|will be removed|no longer supported)'; then
  DEPS=$(echo "$OUTPUT" | grep -iE '(deprecat(ed|ion)|will be removed|no longer supported)' | head -5)
  echo "DEPRECATION WARNING from command output:"
  echo "$DEPS"
fi

# --- 2. Surface security audit findings ---
if echo "$COMMAND" | grep -qE '(npm audit|pip.audit|yarn audit|pnpm audit|govulncheck)'; then
  if echo "$OUTPUT" | grep -qiE '(vulnerabilit|critical|high)'; then
    echo "SECURITY: audit found vulnerabilities. Review output above and address critical/high issues before deploying."
  fi
fi

# --- 3. Warn on npm/pip install adding many deps ---
if echo "$COMMAND" | grep -qE '(npm install|pip install|yarn add|pnpm add)'; then
  # Check for large dependency additions
  if echo "$OUTPUT" | grep -qE 'added [0-9]{3,} packages'; then
    ADDED=$(echo "$OUTPUT" | grep -oE 'added [0-9]+ packages' | head -1)
    echo "WARNING: $ADDED. That's a lot of dependencies. Consider whether a lighter alternative exists."
  fi
fi

# --- 4. Surface build warnings that indicate real issues ---
if echo "$COMMAND" | grep -qE '(go build|go install|go test|make|cargo build|npm run build)'; then
  # Flag unused variable/import warnings that escaped linting
  WARNINGS=$(echo "$OUTPUT" | grep -iE '(unused|unreachable|shadow|implicit|undefined)' | head -5)
  if [ -n "$WARNINGS" ]; then
    echo "BUILD WARNINGS:"
    echo "$WARNINGS"
  fi
fi

# --- 5. Warn on test failures with useful context ---
if echo "$COMMAND" | grep -qE '(go test|pytest|jest|vitest|npm test|yarn test)'; then
  if echo "$OUTPUT" | grep -qiE '(FAIL|FAILED|ERROR|failures=)'; then
    FAIL_COUNT=$(echo "$OUTPUT" | grep -ciE '(FAIL|FAILED)' || echo "?")
    echo "TESTS: $FAIL_COUNT failure indicators detected. Review output above."
  fi
fi

exit 0
