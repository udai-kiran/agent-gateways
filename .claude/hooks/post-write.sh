#!/usr/bin/env bash
# Post-write hook: auto-fix and validate after file writes.
# Receives tool input as JSON on stdin.
set -euo pipefail

INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.file_path // empty')

if [ -z "$FILE_PATH" ] || [ ! -f "$FILE_PATH" ]; then
  exit 0
fi

# --- Go files: format and vet ---
if echo "$FILE_PATH" | grep -q '\.go$'; then
  # Auto-format. No debate about style.
  if command -v gofmt >/dev/null 2>&1; then
    gofmt -w "$FILE_PATH" 2>/dev/null || true
  fi

  # Run vet on the package, not just the file.
  # Catches real bugs (printf format mismatches, unreachable code, etc.)
  PACKAGE_DIR=$(dirname "$FILE_PATH")
  if command -v go >/dev/null 2>&1; then
    VET_OUTPUT=$(cd "$PACKAGE_DIR" && go vet ./... 2>&1) || true
    if [ -n "$VET_OUTPUT" ]; then
      echo "go vet found issues in $(basename "$PACKAGE_DIR"):"
      echo "$VET_OUTPUT"
    fi
  fi

  # Holzmann rule 4: flag functions over 50 lines.
  if command -v awk >/dev/null 2>&1; then
    awk '/^func /{name=$0; start=NR} start && /^}$/{len=NR-start+1; if(len>50) printf "WARNING: %s is %d lines (%s:%d). Consider splitting.\n", name, len, FILENAME, start}' "$FILE_PATH" 2>/dev/null || true
  fi
fi

# --- Python files: format and check ---
if echo "$FILE_PATH" | grep -q '\.py$'; then
  # Auto-format. Ruff is fast; fall back to black.
  if command -v ruff >/dev/null 2>&1; then
    ruff format "$FILE_PATH" 2>/dev/null || true
    # Also auto-fix safe lint rules (unused imports, etc.)
    ruff check --fix --select I,F401 "$FILE_PATH" 2>/dev/null || true
  elif command -v black >/dev/null 2>&1; then
    black -q "$FILE_PATH" 2>/dev/null || true
  fi

  # Syntax check — catches indentation errors and typos before anything else runs.
  SYNTAX_OUTPUT=$(python3 -c "import py_compile; py_compile.compile('$FILE_PATH', doraise=True)" 2>&1) || true
  if echo "$SYNTAX_OUTPUT" | grep -q "SyntaxError"; then
    echo "Syntax error in $FILE_PATH:"
    echo "$SYNTAX_OUTPUT"
  fi

  # Holzmann rule 4: flag functions over 40 lines.
  python3 -c "
import ast, sys
with open('$FILE_PATH') as f:
    try:
        tree = ast.parse(f.read())
    except SyntaxError:
        sys.exit(0)
for node in ast.walk(tree):
    if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
        length = node.end_lineno - node.lineno + 1
        if length > 40:
            print(f'WARNING: {node.name}() is {length} lines ($FILE_PATH:{node.lineno}). Consider splitting.')
" 2>/dev/null || true
fi

# --- TypeScript/React files: type check and lint ---
if echo "$FILE_PATH" | grep -qE '\.(tsx?|jsx?)$'; then
  PROJECT_DIR=$(dirname "$FILE_PATH")
  # Walk up to find the nearest tsconfig.json or package.json
  while [ "$PROJECT_DIR" != "/" ] && [ ! -f "$PROJECT_DIR/package.json" ]; do
    PROJECT_DIR=$(dirname "$PROJECT_DIR")
  done

  if [ -f "$PROJECT_DIR/package.json" ]; then
    # Type check if TypeScript
    if echo "$FILE_PATH" | grep -qE '\.tsx?$' && [ -f "$PROJECT_DIR/tsconfig.json" ]; then
      TSC_OUTPUT=$(cd "$PROJECT_DIR" && npx tsc --noEmit --pretty false 2>&1 | grep "$(basename "$FILE_PATH")" | head -5) || true
      if [ -n "$TSC_OUTPUT" ]; then
        echo "TypeScript errors in $(basename "$FILE_PATH"):"
        echo "$TSC_OUTPUT"
      fi
    fi

    # Lint the specific file
    if [ -f "$PROJECT_DIR/.eslintrc.json" ] || [ -f "$PROJECT_DIR/.eslintrc.js" ] || [ -f "$PROJECT_DIR/eslint.config.js" ] || [ -f "$PROJECT_DIR/eslint.config.mjs" ]; then
      LINT_OUTPUT=$(cd "$PROJECT_DIR" && npx eslint --no-warn --format compact "$FILE_PATH" 2>&1) || true
      if [ -n "$LINT_OUTPUT" ] && ! echo "$LINT_OUTPUT" | grep -q "0 problems"; then
        echo "ESLint issues in $(basename "$FILE_PATH"):"
        echo "$LINT_OUTPUT" | head -10
      fi
    fi
  fi

  # Holzmann rule 4: flag React components with JSX return over 50 lines.
  if echo "$FILE_PATH" | grep -qE '\.(tsx|jsx)$'; then
    node -e "
      const fs = require('fs');
      const src = fs.readFileSync('$FILE_PATH', 'utf8');
      const lines = src.split('\n');
      let inReturn = false, depth = 0, start = 0, funcName = '';
      const funcRe = /(?:function|const|let)\s+(\w+)/;
      for (let i = 0; i < lines.length; i++) {
        const l = lines[i];
        if (/^\s*return\s*\(/.test(l)) { inReturn = true; start = i; depth = 0; }
        if (inReturn) {
          depth += (l.match(/\(/g)||[]).length - (l.match(/\)/g)||[]).length;
          if (depth <= 0 && i > start) {
            const len = i - start + 1;
            if (len > 50) console.log('WARNING: JSX return is ' + len + ' lines ($FILE_PATH:' + (start+1) + '). Extract sub-components.');
            inReturn = false;
          }
        }
      }
    " 2>/dev/null || true
  fi
fi

# --- SQL files: basic safety checks ---
if echo "$FILE_PATH" | grep -qiE '\.sql$'; then
  # Warn on SELECT * in migration files
  if echo "$FILE_PATH" | grep -qiE '(migration|migrate)' && grep -qi 'SELECT \*' "$FILE_PATH"; then
    echo "WARNING: SELECT * in a migration ($FILE_PATH). Pin column names — schema may change."
  fi
fi

# --- Agent definition files: validate frontmatter ---
if echo "$FILE_PATH" | grep -q '\.claude/agents/.*\.md$'; then
  # Check required frontmatter fields
  MISSING=""
  for FIELD in name description tools; do
    if ! head -20 "$FILE_PATH" | grep -q "^${FIELD}:"; then
      MISSING="$MISSING $FIELD"
    fi
  done
  if [ -n "$MISSING" ]; then
    echo "Agent file missing required frontmatter:$MISSING"
  fi
fi

exit 0
