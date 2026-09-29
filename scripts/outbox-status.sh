#!/usr/bin/env bash
# Застряли ли события в outbox и вебхуки GitLab во входящей очереди.
#
#   ./scripts/outbox-status.sh
#
# HTTP может быть «зелёным», а событие не уйти в RabbitMQ — тогда другие сервисы
# и расширения о нём не узнают. Скрипт смотрит на это со стороны БД.
#
# Код выхода 1, если есть записи FAILED или PENDING старше MAX_PENDING_AGE секунд (по умолчанию 60).
set -euo pipefail
source "$(dirname "$0")/lib.sh"
need docker
docker ps --format '{{.Names}}' 2>/dev/null | grep -qx business_db ||
  { echo "Контейнеры БД не запущены (нужен Docker и make docker-up)" >&2; exit 2; }

max_age=${MAX_PENDING_AGE:-60}
problems=0

# report <база> <таблица> <описание>
report() {
  local db=$1 table=$2 title=$3
  echo "${BOLD}${title}${RESET} ${DIM}(${table})${RESET}"

  psql_db "$db" -P footer=off -c "
    SELECT status,
           count(*)                                       AS records,
           date_trunc('second', now() - min(created_at)) AS oldest
    FROM $table GROUP BY status ORDER BY status"

  local stuck
  stuck=$(psql_db "$db" -tA -c "
    SELECT count(*) FROM $table
    WHERE status = 'FAILED'
       OR (status = 'PENDING' AND created_at < now() - make_interval(secs => $max_age))")
  if [ "$stuck" -gt 0 ]; then
    problems=1
    echo "${RED}Проблемных записей: $stuck (FAILED или PENDING дольше ${max_age}s). Последние ошибки:${RESET}"
    psql_db "$db" -P footer=off -c "
      SELECT created_at::timestamp(0) AS created, status, event_type, retry_count AS retries,
             left(coalesce(error_message, '-'), 90) AS error
      FROM $table
      WHERE status <> 'PUBLISHED' AND status <> 'PROCESSED'
      ORDER BY created_at DESC LIMIT 5"
  else
    echo "${GREEN}Застрявших записей нет${RESET}"
  fi
  echo
}

report identity identity.outbox "Outbox сервиса employee"
report business business.outbox "Outbox сервиса business"
report business business.gitlab_webhook_events "Входящие вебхуки GitLab"

exit $problems
