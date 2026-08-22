#!/usr/bin/env bash
# Первичное получение сертификата Let's Encrypt (standalone), как в my-chat.
#
#   cp .env.example .env   # в корне репозитория на VPS, заполнить DOMAIN и LETSENCRYPT_EMAIL
#   cd deploy/prod
#   bash init-ssl.sh --staging
#   bash init-ssl.sh
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
COMPOSE_FILE="${SCRIPT_DIR}/docker-compose.prod.yml"
ENV_FILE="${REPO_ROOT}/.env"

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "нет ${ENV_FILE} — скопируйте .env.example и заполните секреты" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
source "${ENV_FILE}"
set +a

if [[ -z "${DOMAIN:-}" || -z "${LETSENCRYPT_EMAIL:-}" ]]; then
  echo "в .env нужны DOMAIN и LETSENCRYPT_EMAIL" >&2
  exit 1
fi

CERTBOT_FLAGS="--non-interactive --agree-tos --email ${LETSENCRYPT_EMAIL}"
STAGING=""

for arg in "$@"; do
  case "$arg" in
    --staging)
      STAGING="--staging"
      echo "режим staging: тестовый сертификат"
      ;;
    *)
      echo "неизвестный аргумент: $arg" >&2
      exit 1
      ;;
  esac
done

COMPOSE="docker compose -f ${COMPOSE_FILE} --env-file ${ENV_FILE}"

echo "==> ${SCRIPT_DIR}"
cd "${SCRIPT_DIR}"

echo "==> останавливаем nginx (порт 80 для certbot)"
${COMPOSE} stop nginx 2>/dev/null || true

echo "==> поднимаем db и backend"
${COMPOSE} up -d db
${COMPOSE} exec db sh -c 'until pg_isready -U "${POSTGRES_USER}" -d "${POSTGRES_DB}"; do sleep 1; done'
${COMPOSE} up -d backend frontend

echo "==> certbot standalone для ${DOMAIN}"
mkdir -p "${SCRIPT_DIR}/certbot/conf" "${SCRIPT_DIR}/certbot/www"

docker run --rm \
  -v "${SCRIPT_DIR}/certbot/conf:/etc/letsencrypt" \
  -v "${SCRIPT_DIR}/certbot/www:/var/www/certbot" \
  -p 80:80 \
  certbot/certbot:latest certonly \
    --standalone \
    --domain "${DOMAIN}" \
    ${CERTBOT_FLAGS} \
    ${STAGING}

echo "==> nginx + автопродление"
${COMPOSE} up -d nginx
${COMPOSE} up -d certbot nginx-reloader

echo "готово: https://${DOMAIN}/healthz"
if [[ -n "${STAGING}" ]]; then
  echo "staging. Боевой сертификат: bash ${BASH_SOURCE[0]}"
fi
