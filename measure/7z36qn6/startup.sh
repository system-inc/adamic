#!/bin/bash
# #7z36qn6: where a one-line Adamic hello world's startup goes against C's. strace -f of each, every system
# call in order with the thread that made it; which signals the runtime sigactions, and with what; the
# parent's ignored signals (SigIgn) as this job inherits them; and what the cloned thread does before it's
# joined. Never lands.
set -u
T=$(mktemp -d)
trap 'rm -rf "$T"' EXIT
echo "box $(hostname) $(nproc) cpus, $(uname -r), tree $(git rev-parse HEAD)"
grep -E "^Sig(Ign|Cgt|Blk)" /proc/self/status
command -v strace > /dev/null || { echo "FAIL no strace"; exit 1; }
go build -buildvcs=false -o "$T/adamic" ./cmd/adamic || { echo "FAIL adamic build"; exit 1; }
printf 'console.log("hello world");\n' > "$T/hello.a"
printf '#include <stdio.h>\nint main(void) { puts("hello world"); return 0; }\n' > "$T/hello.c"
"$T/adamic" build "$T/hello.a" -o "$T/adamic-hello" > "$T/b.log" 2>&1 || { cat "$T/b.log"; echo "FAIL build"; exit 1; }
clang -O2 -o "$T/c-hello" "$T/hello.c" || { echo "FAIL clang"; exit 1; }
for name in adamic-hello c-hello; do
	strace -f -c -o "$T/$name.count" "$T/$name" > /dev/null
	echo "== $name: system calls by name"
	sed -n '1,30p' "$T/$name.count"
done
echo "== adamic-hello, every call in order with its thread (strace -f), arguments cut"
strace -f -o "$T/trace" "$T/adamic-hello" > /dev/null
cut -c1-160 "$T/trace" | grep -v "rt_sigaction" | head -120
echo "== adamic-hello's rt_sigaction calls: how many set a handler, set SIG_DFL, set SIG_IGN, or only read"
python3 -I - "$T/trace" <<'PY'
import re, sys, collections
kinds = collections.Counter()
signals = collections.defaultdict(set)
for line in open(sys.argv[1]):
    m = re.search(r"rt_sigaction\((\w+), (NULL|\{sa_handler=([^,]+))", line)
    if not m:
        continue
    signal, new, handler = m.group(1), m.group(2), m.group(3)
    kind = "read only" if new == "NULL" else ("SIG_DFL" if handler == "SIG_DFL" else "SIG_IGN" if handler == "SIG_IGN" else "a handler")
    kinds[kind] += 1
    signals[kind].add(signal)
for kind, count in kinds.most_common():
    names = sorted(signals[kind])
    print(f"{kind}: {count} calls over {len(names)} signals: {' '.join(names[:40])}")
PY
echo "== the cloned thread: its id and every call it made"
tid=$(grep -oE "clone3\(.*= [0-9]+" "$T/trace" | grep -oE "[0-9]+$" | head -1)
echo "clone3 returned thread $tid"
[ -n "$tid" ] && grep -E "^$tid " "$T/trace" | cut -c1-200 | head -40
echo "== where the clone happens: the 15 calls before it in the main thread"
main=$(head -1 "$T/trace" | awk '{print $1}')
grep -nE "^$main " "$T/trace" | grep -B15 "clone3" | cut -c1-200 | head -20
echo PASS startup
