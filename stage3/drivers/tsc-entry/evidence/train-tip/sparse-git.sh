#!/usr/bin/env bash
set -euo pipefail
if [ "${1:-}" = clone ] && [ "${2:-}" = --no-hardlinks ]; then
    /usr/bin/git clone --no-checkout "${@:2}"
    destination=${@: -1}
    /usr/bin/git -C "$destination" sparse-checkout set --no-cone /src/ /scripts/ /package.json /package-lock.json /tests/baselines/reference/api/typescript.d.ts
else
    exec /usr/bin/git "$@"
fi
