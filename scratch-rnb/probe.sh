set -u
go build -o "$TMPDIR/adamic" ./cmd/adamic || { echo FAIL build; exit 1; }
echo "== node"; node --experimental-strip-types scratch-rnb/p.ts 2>&1
echo "== c"; "$TMPDIR/adamic" c scratch-rnb/p.ts > "$TMPDIR/p.c" 2>&1; echo "exit=$?"
wc -l "$TMPDIR/p.c"
echo "== functions holding temporary_body, whole"
awk '/^[a-zA-Z].*\) \{$/{buf=""; keep=0} {buf=buf"\n"$0} /temporary_body/{keep=1} /^\}$/{ if (keep) print buf; buf=""; keep=0 }' "$TMPDIR/p.c" | head -300
echo "== js"; "$TMPDIR/adamic" js scratch-rnb/p.ts > "$TMPDIR/p.mjs" 2>&1; echo "exit=$?"
echo "== build"; "$TMPDIR/adamic" build scratch-rnb/p.ts -o "$TMPDIR/p" > "$TMPDIR/b.log" 2>&1; echo "exit=$?"; grep -m3 "error" "$TMPDIR/b.log"
