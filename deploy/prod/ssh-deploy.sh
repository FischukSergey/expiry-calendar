#!/usr/bin/env bash
# Forced command для ключа GitHub Actions (authorized_keys command=).
# Клиент передаёт только SHA или ref — он попадает в SSH_ORIGINAL_COMMAND.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REF="${SSH_ORIGINAL_COMMAND:-origin/main}"

exec "${SCRIPT_DIR}/deploy.sh" "${REF}"
