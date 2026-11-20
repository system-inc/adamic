set -u
# Is consumers.ts's `v + '!'` stop new with the fix, or already on main? Each instance alone, at head and at main.
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
mkdir -p "$TMPDIR/j"
for instance in 'undefined' 'string | undefined' 'string'; do
	name=$(echo "$instance" | tr -d ' |')
	cat > "$TMPDIR/j/$name.ts" <<EOF
interface Box<T> { f: T }
function joined<T extends string | undefined>(o: Box<T>): string { const v = o.f; if (v === undefined) { return 'none'; } return v + '!'; }
console.log(joined<$instance>({ f: undefined }));
EOF
done
go build -o "$TMPDIR/adamic-head" ./cmd/adamic || { echo FAIL head build; exit 1; }
cp internal/lower/narrowed.go "$TMPDIR/narrowed.go.head"
cp scratch-rnbfix/narrowed.go.main internal/lower/narrowed.go
go build -o "$TMPDIR/adamic-main" ./cmd/adamic || { echo FAIL main build; exit 1; }
cp "$TMPDIR/narrowed.go.head" internal/lower/narrowed.go
for which in head main; do
	for file in "$TMPDIR"/j/*.ts; do
		out=$( cd "$TMPDIR/j" && "$TMPDIR/adamic-$which" build "$(basename "$file")" -o "$TMPDIR/j/bin" 2>&1 | grep -m1 -E "error|can't|NotYet|stage" )
		code=$?
		echo "$which $(basename "$file"): ${out:-built} | run: $("$TMPDIR/j/bin" 2>&1 | tr '\n' ' ')"
		rm -f "$TMPDIR/j/bin"
	done
done
echo done
