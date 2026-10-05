#!/bin/sh
# check.sh: the dedication two ways, assembly and Adamic, one answer.
set -e
cd "$(dirname "$0")"
./build.sh
./dedication > native.out
node dedication.a > adamic.out
cmp native.out adamic.out
rm native.out adamic.out
echo "ok: $(./dedication | wc -c | tr -d ' ') bytes, both ways"
