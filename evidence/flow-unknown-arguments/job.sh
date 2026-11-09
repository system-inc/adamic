set -u
. ~/adamic-tools/env.sh
git -c protocol.file.allow=always submodule update --init --recursive --depth 1 > /tmp/fb-sub.log 2>&1; echo "submodule exit=$?"
HEAD_SHA=$(git rev-parse HEAD); PARENT=0f44327f7aaf29fd302014e2ae438ce74a8074e1
echo "head $HEAD_SHA parent $PARENT"; nproc; uptime
O=/tmp/flowbench-$(date +%s); mkdir -p $O
cat > $O/bench.a <<'BENCH'
// A tsc-shaped loop for #41vb71c's cost: a graph program (a cycle of Nodes), and interface-method calls
// through a parameter, whose targets the analysis can't bound, each handed fresh objects it doesn't keep.
interface Node { name: string; next: Node[]; kids: Node[][] }
interface Info { depth: number; flags: number }
interface Visitor { visit(n: Node, info: Info, path: number[]): number }
const a: Node = { name: 'a' + '', next: [], kids: [] };
const b: Node = { name: 'b' + '', next: [a], kids: [] };
a.next.push(b); // a cycle: the program is a graph program
function run(v: Visitor, n: Node, count: number): number {
	let sum = 0;
	for (let i = 0; i < count; i++) {
		sum += v.visit(n, { depth: i & 7, flags: i & 255 }, [i & 3, i & 5]);
	}
	return sum;
}
const visitors: Visitor[] = [
	{ visit: (n: Node, info: Info, path: number[]): number => info.depth + n.name.length + path.length },
	{ visit: (n: Node, info: Info, path: number[]): number => (info.flags & 1) + n.next.length + (path[0] ?? 0) },
];
let total = 0;
for (const v of visitors) {
	total += run(v, a, 5000000);
}
console.log(`${total}`);
BENCH
go build -o $O/adamic-head ./cmd/adamic > $O/build-head.log 2>&1; echo "build head exit=$?"
git worktree add --detach $O/parent $PARENT > /dev/null 2>&1; (cd $O/parent && git -c protocol.file.allow=always submodule update --init --recursive --depth 1 > /dev/null 2>&1; go build -o $O/adamic-parent ./cmd/adamic > $O/build-parent.log 2>&1; echo "build parent exit=$?")
$O/adamic-head build --help > $O/help.txt 2>&1; grep -iE 'count|release|sanit|-o ' $O/help.txt | head -12
for side in parent head; do
  (cd $O && timeout 600 ./adamic-$side build bench.a -o bench-$side > build-bench-$side.log 2>&1); echo "compile $side exit=$?"; tail -3 $O/build-bench-$side.log
done
for side in parent head; do
  [ -x $O/bench-$side ] || continue
  $O/bench-$side > $O/out-$side.txt 2>&1; echo "$side output: $(cat $O/out-$side.txt)"
done
for round in 1 2 3 4 5; do for side in parent head; do
  [ -x $O/bench-$side ] || continue
  /usr/bin/time -f "$side round=$round wall=%e user=%U rss_kb=%M" $O/bench-$side > /dev/null 2>> $O/times.txt
done; done
cat $O/times.txt
python3 - $O/times.txt <<'PY'
import sys,re,statistics as st
d={}
for line in open(sys.argv[1]):
    m=re.match(r'(\w+) round=\d+ wall=([\d.]+) user=([\d.]+) rss_kb=(\d+)',line)
    if m: d.setdefault(m.group(1),[]).append((float(m.group(2)),int(m.group(4))))
for k,v in d.items(): print(k,'median wall',st.median(x[0] for x in v),'min',min(x[0] for x in v),'max',max(x[0] for x in v),'rss_kb',max(x[1] for x in v))
if 'parent' in d and 'head' in d:
    a=st.median(x[0] for x in d['parent']); b=st.median(x[0] for x in d['head']); print('head/parent wall', round(b/a,4))
PY
for side in parent head; do
  (cd $O && timeout 600 ./adamic-$side build --count bench.a -o counted-$side > build-counted-$side.log 2>&1); echo "counted compile $side exit=$?"; tail -2 $O/build-counted-$side.log
  [ -x $O/counted-$side ] && { $O/counted-$side > $O/counted-out-$side.txt 2> $O/counted-err-$side.txt; echo "counted $side stdout: $(head -c 200 $O/counted-out-$side.txt)"; echo "counted $side stderr:"; head -c 1500 $O/counted-err-$side.txt; echo; }
done
