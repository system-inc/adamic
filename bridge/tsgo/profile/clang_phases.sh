#!/usr/bin/env bash
# Put a copy named clang first on PATH for a scratch profiling build only.
set -euo pipefail
: "${ADAMIC_PROFILE_CLANG:?set the absolute real clang path}"
: "${ADAMIC_PHASE_INSTRUMENT:?set the absolute native_phases.py path}"
for argument in "$@"; do
    case "$argument" in
        */main.c) python3 "$ADAMIC_PHASE_INSTRUMENT" "$argument" ;;
    esac
done
exec "$ADAMIC_PROFILE_CLANG" -g "$@"
