# Общие настройки и функции для scripts/*.sh. Подключается через `source`.
# Совместимо с bash 3.2 (системный bash macOS).

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

LOKI_URL=${LOKI_URL:-http://localhost:3100}
PROM_URL=${PROM_URL:-http://localhost:9090}
ALERTMANAGER_URL=${ALERTMANAGER_URL:-http://localhost:9093}
GRAFANA_URL=${GRAFANA_URL:-http://localhost:3000}

# Цвета только для терминала: в файлах и пайпах — чистый текст.
if [ -t 1 ]; then
  RED=$'\e[31m' GREEN=$'\e[32m' YELLOW=$'\e[33m' DIM=$'\e[2m' BOLD=$'\e[1m' RESET=$'\e[0m'
else
  RED='' GREEN='' YELLOW='' DIM='' BOLD='' RESET=''
fi

need() {
  command -v "$1" >/dev/null 2>&1 || { echo "Нужна утилита: $1" >&2; exit 2; }
}

# SQL в контейнерах из backend/docker-compose.yml. Пользователь и база совпадают с именем.
#   psql_db identity -c "SELECT 1"
psql_db() {
  local db=$1; shift
  docker exec -i "${db}_db" psql -U "$db" -d "$db" -v ON_ERROR_STOP=1 "$@"
}

# Печатает JSON-строку лога F5 в читаемом виде:
#   2026-09-29T10:15:02Z business WARN request completed status=404 route=/api/v1/tasks/{id} ...
# Не-JSON строки (например, логи Postgres) выводятся как есть.
format_log() {
  local service=$1 line=$2
  printf '%s\n' "$line" | jq -r --arg s "$service" '
    "\(.time // "-") \($s) \(.level // "-") \(.msg // "")  " +
    ([to_entries[] | select(.key | IN("time", "level", "msg", "service") | not)
      | "\(.key)=\(.value)"] | join(" "))' 2>/dev/null || printf '%s %s\n' "$service" "$line"
}
