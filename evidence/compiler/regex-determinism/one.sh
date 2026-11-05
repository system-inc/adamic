#!/bin/bash
# usage: one.sh <worktree> <binary> <file>   prints: file exit stderr-sha stdout-sha
w=$1; bin=$2; f=$3
cd "$w" || exit 1
e=$(mktemp); o=$(mktemp)
"$bin" c "$f" >"$o" 2>"$e"; rc=$?
echo "$(basename $f) $rc $(sed "s#$w##g" "$e" | sha256sum | cut -c1-12) $(sha256sum <"$o" | cut -c1-12)"
rm -f "$e" "$o"
