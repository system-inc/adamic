#!/usr/bin/env bash
set -eu
cd "$(dirname "$0")/../../../../.."
source /workspace/adamic-tools/env.sh
scratch=$(mktemp -d /tmp/wave1-11-registration.XXXXXX)
cp -r stage1/cohere/lint/rules "$scratch/rules"
go run ./cmd/lint-registry -root "$scratch"
echo 'baseline exit=0'
mv "$scratch/rules/no-debugger/rule.ts" "$scratch/rules/no-debugger/rule.a"
if go run ./cmd/lint-registry -root "$scratch"; then
    echo 'unexpected: .a-only module accepted' >&2
    exit 1
fi
echo 'a-only module rejected as expected'
