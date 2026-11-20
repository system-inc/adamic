set -u
# For each first-parent commit from df1c19ae to 358b95d5 that touches Go, build adamic there and
# try `adamic build` of the program: the first green sha after a red names the fix.
root=$(pwd)
cp scratch-rnb/p.ts "$TMPDIR/p.ts"
git diff --quiet df1c19ae 358b95d5 -- cohere || echo "NOTE cohere pin moved in range"
shas="df1c19ae $(git rev-list --reverse --first-parent df1c19ae..358b95d5)"
for s in $shas; do
	short=$(git rev-parse --short=10 "$s")
	if [ "$short" != "$(git rev-parse --short=10 df1c19ae)" ] && git diff --quiet "$s~1" "$s" -- '*.go' 'internal/native_runtime/c'; then
		echo "SKIP $short no Go"
		continue
	fi
	w="$TMPDIR/w-$short"
	git worktree add -q --detach "$w" "$s" 2>/dev/null || { echo "GAP $short worktree"; continue; }
	rmdir "$w/cohere" 2>/dev/null; rm -rf "$w/cohere"; ln -s "$root/cohere" "$w/cohere"
	( cd "$w" && GOFLAGS=-buildvcs=false go build -o "$TMPDIR/adamic-$short" ./cmd/adamic ) > "$TMPDIR/b-$short.log" 2>&1 || { echo "GAP $short go build"; tail -5 "$TMPDIR/b-$short.log"; continue; }
	( cd "$TMPDIR" && "$TMPDIR/adamic-$short" build p.ts -o "$TMPDIR/p-$short" ) > "$TMPDIR/r-$short.log" 2>&1
	code=$?
	if [ $code -eq 0 ]; then
		echo "GREEN $short $(git log -1 --format=%s "$s" | cut -c1-90) | run: $("$TMPDIR/p-$short" 2>&1 | tr '\n' ' ')"
	else
		echo "RED $short exit $code $(grep -m1 -o 'error: [^[]*' "$TMPDIR/r-$short.log")"
	fi
done
echo bisect done
