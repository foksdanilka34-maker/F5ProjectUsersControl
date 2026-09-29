#!/usr/bin/env bash
# Все логи одного запроса: gateway и сервис, по порядку времени.
#
#   ./scripts/trace.sh <request_id> [период]      период по умолчанию 1h (формат Loki: 30m, 2h, 24h)
#
# Ищет в Loki. Если Loki не поднят (make obs-up), ищет прямо в logs/*.log.
set -euo pipefail
source "$(dirname "$0")/lib.sh"
need curl
need jq

id=${1:?Использование: trace.sh <request_id> [период, например 1h]}
since=${2:-1h}

from_loki() {
  curl -sfG "$LOKI_URL/loki/api/v1/query_range" \
    --data-urlencode "query={service_name=~\".+\"} |= \"$id\"" \
    --data-urlencode "since=$since" \
    --data-urlencode "limit=1000" \
    --data-urlencode "direction=forward" |
    jq -r '.data.result[] | .stream.service_name as $s | .values[] | [.[0], $s, .[1]] | @tsv'
}

# Та же форма вывода (время, сервис, строка) из локальных файлов логов.
from_files() {
  local f
  for f in "$ROOT_DIR"/logs/*.log; do
    [ -e "$f" ] || continue
    grep -F -- "$id" "$f" | while IFS= read -r line; do
      printf '%s\t%s\t%s\n' "$(printf '%s' "$line" | jq -r '.time // ""' 2>/dev/null)" "$(basename "$f" .log)" "$line"
    done
  done
}

if curl -sf "$LOKI_URL/ready" >/dev/null 2>&1; then
  source_name="Loki, за $since"
  rows=$(from_loki)
else
  source_name="logs/*.log (Loki недоступен)"
  rows=$(from_files)
fi

if [ -z "$rows" ]; then
  echo "${YELLOW}Строк с request_id=$id не найдено${RESET} ($source_name)" >&2
  exit 1
fi

echo "${BOLD}request_id=$id${RESET} ${DIM}($source_name)${RESET}"
printf '%s\n' "$rows" | sort -t $'\t' -k1,1 | while IFS=$'\t' read -r _ service line; do
  formatted=$(format_log "$service" "$line")
  case "$line" in
    *'"level":"ERROR"'*) echo "${RED}${formatted}${RESET}" ;;
    *'"level":"WARN"'*) echo "${YELLOW}${formatted}${RESET}" ;;
    *) echo "$formatted" ;;
  esac
done
