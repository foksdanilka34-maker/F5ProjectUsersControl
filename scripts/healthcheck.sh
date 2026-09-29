#!/usr/bin/env bash
# Состояние всего окружения F5 одной командой.
#
#   ./scripts/healthcheck.sh
#
# Код выхода: 0 — сервисы и инфраструктура в порядке, 1 — что-то из них недоступно.
# Стек наблюдаемости проверяется, но на код выхода не влияет (он может быть не поднят).
set -uo pipefail
source "$(dirname "$0")/lib.sh"
need curl

failed=0

# check <раздел: core|obs> <название> <команда...>
check() {
  local kind=$1 name=$2; shift 2
  local out
  if out=$("$@" 2>&1); then
    printf '  %-26s %sOK%s   %s\n' "$name" "$GREEN" "$RESET" "${DIM}${out}${RESET}"
  else
    if [ "$kind" = core ]; then
      failed=1
      printf '  %-26s %sFAIL%s %s\n' "$name" "$RED" "$RESET" "$out"
    else
      printf '  %-26s %sDOWN%s %s\n' "$name" "$YELLOW" "$RESET" "${DIM}${out}${RESET}"
    fi
  fi
}

http_ok() {
  local res
  res=$(curl -sf -m 3 -o /dev/null -w '%{http_code} за %{time_total}s' "$1") || { echo "нет ответа"; return 1; }
  echo "$res"
}

docker_up=0
docker info >/dev/null 2>&1 && docker_up=1

in_docker() {
  [ "$docker_up" = 1 ] || { echo "Docker не запущен"; return 1; }
  docker ps --format '{{.Names}}' | grep -qx "$1" || { echo "контейнер $1 не запущен (make docker-up)"; return 1; }
}
pg_ready() { in_docker "$1" && docker exec "$1" pg_isready -q -U "$2" -d "$2" && echo "принимает соединения"; }
rabbit_ok() { in_docker rabbitmq && docker exec rabbitmq rabbitmq-diagnostics -q ping >/dev/null && echo "ping ok"; }

# Цели Prometheus, которые не отвечают (например, не запущен один из сервисов).
prom_targets() {
  local down
  down=$(curl -sf -m 3 "$PROM_URL/api/v1/targets" |
    jq -r '[.data.activeTargets[] | select(.health != "up") | .labels.job] | join(", ")') || { echo "нет ответа"; return 1; }
  [ -z "$down" ] && echo "все цели up" || { echo "не отвечают: $down"; return 1; }
}

firing_alerts() {
  local alerts
  alerts=$(curl -sf -m 3 "$ALERTMANAGER_URL/api/v2/alerts?active=true&silenced=false&inhibited=false" |
    jq -r '[.[] | "\(.labels.alertname)/\(.labels.severity)"] | join(", ")') || { echo "нет ответа"; return 1; }
  [ -z "$alerts" ] && echo "активных алертов нет" || { echo "горят: $alerts"; return 1; }
}

echo "${BOLD}Сервисы F5${RESET}"
check core "gateway      :8080" http_ok http://localhost:8080/health
check core "employee     :8081" http_ok http://localhost:8081/health
check core "business     :8082" http_ok http://localhost:8082/health

echo "${BOLD}Инфраструктура${RESET}"
check core "identity_db" pg_ready identity_db identity
check core "business_db" pg_ready business_db business
check core "rabbitmq" rabbit_ok

if command -v jq >/dev/null; then
  echo "${BOLD}Наблюдаемость${RESET}"
  check obs "prometheus   :9090" http_ok "$PROM_URL/-/ready"
  check obs "  targets" prom_targets
  check obs "alertmanager :9093" http_ok "$ALERTMANAGER_URL/-/ready"
  check obs "  firing alerts" firing_alerts
  check obs "loki         :3100" http_ok "$LOKI_URL/ready"
  check obs "grafana      :3000" http_ok "$GRAFANA_URL/api/health"
fi

exit $failed
