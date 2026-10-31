#!/bin/bash
# usage: compare.sh <file>  ; runs both binaries on one .a file, writes outputs under /tmp/ev-out
f="$1"; rel="${f#/home/user/adamic/}"; key="${rel//\//__}"
for v in main change; do
  /tmp/ev-bin/$v/adamic c "$f" > "/tmp/ev-out/$v/$key.stdout" 2> "/tmp/ev-out/$v/$key.stderr"
  echo $? > "/tmp/ev-out/$v/$key.exit"
done
