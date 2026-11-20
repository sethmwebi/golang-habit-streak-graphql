#!/usr/bin/env bash
set -euo pipefail

# migrate.sh - wrapper around golang-migrate for the habits DB

MIGRATIONS_PATH="${MIGRATIONS_PATH:-migrations}"
DB_PATH="${DB_PATH:-./data/habits.db}"
DATABASE_URL="sqlite3://${DB_PATH}"

CMD="${1:-up}" # default action is "up"
shift || true

if ! command -v migrate &>/dev/null; then
	echo "Error: 'migrate' CLI not found. Install with:"
	echo "  go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
	exit 1
fi

mkdir -p "$(dirname "$DB_PATH")"

echo "Running: migrate -path $MIGRATIONS_PATH -database $DATABASE_URL $CMD $*"
migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" "$CMD" "$@"
