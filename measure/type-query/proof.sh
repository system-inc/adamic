#!/bin/bash
# #7yz6rtp's proof at this head: the fixture through every family, its count row, the mutant, and a
# typeof <host member> in a type through each sibling walk. Never lands.
set -u
T=$(mktemp -d)
echo "box $(hostname), at $(git rev-parse --short=10 HEAD)"
go build -buildvcs=false -o $T/adamic ./cmd/adamic > $T/b.log 2>&1 && echo "PASS adamic built" || { echo "FAIL build"; tail -30 $T/b.log; exit 1; }
F=internal/oracle/testdata/timers_source/type_query.a
$T/adamic c $F > $T/f.c 2> $T/f.err && echo "PASS type_query lowers" || { echo "FAIL type_query: $(head -3 $T/f.err)"; }
node $F > $T/node.out 2>&1; echo "node: $(tr '\n' '|' < $T/node.out)"
go test ./internal/oracle -run 'TestCounts' -count=1 -args -update-counts > $T/uc.log 2>&1; echo "update-counts exit $?"
echo "== recorded: $(git status --porcelain internal/oracle/testdata/counts | tr '\n' ' ')"
for f in $(git status --porcelain internal/oracle/testdata/counts | awk '{print $2}'); do echo "FILE $f"; cat "$f"; done
go test ./internal/oracle -run '^Test[A-Za-z0-9]+_[0-9]{3}$/type_query|TestWASIRefus|TestCounts|TestFamiliesOwnEveryCaseOnce|TestFamilyWrappersAreCurrent' -count=1 > $T/o.log 2>&1; e=$?
grep -E -- '--- FAIL|^(ok|FAIL)\s' $T/o.log | head -30; [ $e = 0 ] && echo "PASS oracle (type_query families, wasi, counts)" || { echo "FAIL oracle exit $e"; grep -B3 -A25 -- '--- FAIL' $T/o.log | head -120; }
go test ./internal/lower -count=1 > $T/l.log 2>&1 && echo "PASS lower whole" || { echo "FAIL lower"; grep -B3 -A20 -- '--- FAIL' $T/l.log | head -80; tail -5 $T/l.log; }
echo "== mutant: the inTypeQuery clause dropped"
python3 - <<'PY'
p='internal/lower/library_node_timers.go'; s=open(p).read()
old='ast.IsPartOfTypeNode(node) || inTypeQuery(node) || called(node)'
assert s.count(old)==1; open(p,'w').write(s.replace(old,'ast.IsPartOfTypeNode(node) || called(node)'))
PY
go build -buildvcs=false -o $T/adamic.m ./cmd/adamic > $T/bm.log 2>&1 || { echo "FAIL mutant build"; tail -5 $T/bm.log; }
$T/adamic.m c $F > /dev/null 2> $T/m.err && echo "MUTANT SURVIVED: type_query lowers" || echo "mutant: type_query stops: $(head -1 $T/m.err)"
go test ./internal/oracle -run '^Test[A-Za-z0-9]+_[0-9]{3}$/type_query' -count=1 > $T/om.log 2>&1 && echo "MUTANT SURVIVED: oracle passed" || echo "mutant caught by: $(grep -oE -- '--- FAIL: [^ ]+' $T/om.log | sort -u | tr '\n' ' ')"
python3 - <<'PY'
p='internal/lower/library_node_timers.go'; s=open(p).read()
old='ast.IsPartOfTypeNode(node) || called(node)'
assert s.count(old)==1; open(p,'w').write(s.replace(old,'ast.IsPartOfTypeNode(node) || inTypeQuery(node) || called(node)'))
PY
git diff --quiet internal/lower && echo "PASS mutant undone" || echo "FAIL mutant not undone"
echo "== sibling walks: typeof <host member> in a type, one program each, at the head"
probe() { # name, imports, type
  printf '%s\ntype Held = %s;\nfunction absent(value: Held | undefined): boolean {\n    return value === undefined;\n}\nconsole.log(String(absent(undefined)));\n' "$2" "$3" > $T/p_$1.a
  for b in adamic adamic.m; do
    if $T/$b c $T/p_$1.a > /dev/null 2> $T/p_$1.err; then echo "probe $1 ($3) [$b]: lowers"; else echo "probe $1 ($3) [$b]: $(head -2 $T/p_$1.err | tr '\n' ' ')"; fi
  done
}
probe clear "import 'node:timers';" "typeof clearTimeout"
probe set "import 'node:timers';" "typeof setTimeout"
probe params "import 'node:timers';" "Parameters<typeof clearTimeout>[0]"
probe stdout "" "typeof process.stdout"
probe stdoutset "" "typeof process.stdout.setBlocking"
probe env "" "typeof process.env"
probe platform "" "typeof process.platform"
probe process "" "typeof process"
probe stack "" "typeof Error.stackTraceLimit"
probe require "" "typeof require"
probe module "" "typeof module"
probe dirname "" "typeof __dirname"
probe performance "" "typeof performance"
probe profiler "" "typeof onProfilerEvent"
probe realpath "import * as fs from 'node:fs';" "typeof fs.realpathSync"
probe native "import * as fs from 'node:fs';" "typeof fs.realpathSync.native"
probe watch "import * as fs from 'node:fs';" "typeof fs.watch"
probe stat "import * as fs from 'node:fs';" "ReturnType<typeof fs.statSync>"
probe global "" "typeof globalThis"
echo "== end"
