#!/usr/bin/env bash
# Выкладка прода на этой машине: fetch + checkout SHA, compose --build, healthz.
# Не читает и не перезаписывает .env, не трогает тома Postgres.
# Повтор на том же SHA безопасен (пересборка, стек остаётся).
# После healthz 200 — dangling-образы и неиспользуемый build cache (не -a, не тома).
#
#   ./deploy/prod/deploy.sh              # origin/main
#   ./deploy/prod/deploy.sh <sha|ref>
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
COMPOSE_FILE="${SCRIPT_DIR}/docker-compose.prod.yml"
ENV_FILE="${REPO_ROOT}/.env"
REF="${1:-origin/main}"
# HTTPS с публичного имени: так же проверяет Actions и откат.
HEALTH_URL="${HEALTH_URL:-https://duekeep.ru/healthz}"
# Сборка Go+Node на 2 ГБ RAM может занять несколько минут.
HEALTH_TIMEOUT="${HEALTH_TIMEOUT:-600}"
HEALTH_INTERVAL=5

# Только после живого стека. Не -a / не system prune / не volume: текущие теги и том БД остаются.
# Ошибка prune не валит выкладку: сайт уже отвечает.
prune_build_leftovers() {
  echo "==> prune dangling images и неиспользуемый build cache"
  if ! docker image prune -f; then
    echo "image prune не удался (стек уже живой)" >&2
  fi
  if ! docker builder prune -f; then
    echo "builder prune не удался (стек уже живой)" >&2
  fi
}

if [[ ! "${REF}" =~ ^[a-zA-Z0-9._/-]+$ ]]; then
  echo "недопустимый ref: ${REF}" >&2
  exit 1
fi

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "нет ${ENV_FILE} — секреты только на диске VPS, скрипт их не создаёт" >&2
  exit 1
fi

if [[ ! -f "${COMPOSE_FILE}" ]]; then
  echo "нет ${COMPOSE_FILE}" >&2
  exit 1
fi

cd "${REPO_ROOT}"

echo "==> fetch origin"
git fetch --prune origin

if ! git rev-parse --verify --quiet "${REF}^{commit}" >/dev/null; then
  echo "==> fetch ${REF}"
  git fetch origin "${REF}"
fi

SHA="$(git rev-parse --verify "${REF}^{commit}")"
TAG="$(git describe --tags --exact-match "${SHA}" 2>/dev/null || true)"
if [[ "${TAG}" == "v1.0.0" ]]; then
  echo "тег v1.0.0 — сдача с общим каталогом, на прод не выкладываем" >&2
  exit 1
fi

echo "==> checkout --detach ${SHA}"
# -f: на сервере не должно быть локальных правок tracked-файлов. .env не в git.
git checkout --force --detach "${SHA}"

echo "==> compose up --build (не трогаем тома)"
docker compose -f "${COMPOSE_FILE}" --env-file "${ENV_FILE}" up -d --build

echo "==> ждём ${HEALTH_URL} (таймаут ${HEALTH_TIMEOUT}s)"
deadline=$((SECONDS + HEALTH_TIMEOUT))
code="000"
while ((SECONDS < deadline)); do
  code="$(curl -sS -o /dev/null -w '%{http_code}' --max-time 15 "${HEALTH_URL}" || true)"
  if [[ "${code}" == "200" ]]; then
    echo "healthz 200, SHA ${SHA}"
    prune_build_leftovers
    exit 0
  fi
  sleep "${HEALTH_INTERVAL}"
done

echo "healthz не 200 за ${HEALTH_TIMEOUT}s (последний код ${code})" >&2
exit 1
