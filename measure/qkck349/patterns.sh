#!/bin/bash
# #qkck349: regexredux's cost split by pattern (Runtime's ask on #qkck349). One program per operation:
# the port's cleanup replace, then that one match (its count printed) or one replace on the cleaned
# text (its length printed), and a base that does the cleanup alone. Each built natively and run on
# Node over fasta's 1M output, output checked, median of 3 for each, the base's time beside them so an
# operation's own cost is its time less the base's. The five replaces each run on the cleaned text,
# not on the previous replace's output as the port chains them, so their cost is close, not exact.
# Then perf annotate of adamic_regex_regular_find over the whole port at 1M: its hottest lines, and
# whether the lazy DFA's code in regexp_regular.c runs at all. Never lands.
set -u
T=$(mktemp -d)
echo "box $(hostname) $(nproc) cpus, $(uname -r), tree $(git rev-parse HEAD)"
echo "load before $(cut -d' ' -f1-3 /proc/loadavg); busy: $(ps -eo pcpu=,comm= | awk '$1 > 5 {printf "%s(%s) ", $2, $1}')"
go build -buildvcs=false -o "$T/adamic" ./cmd/adamic || { echo "FAIL adamic build"; exit 1; }
"$T/adamic" build bench/holdout/fasta.a -o "$T/fasta" > /dev/null 2>&1 || { echo "FAIL fasta build"; exit 1; }
"$T/fasta" 1000000 > "$T/in.txt"
mkdir -p bench/holdout/patterns
operations=(
	'base|'
	'match|/agggtaaa|tttaccct/ig'
	'match|/[cgt]gggtaaa|tttaccc[acg]/ig'
	'match|/a[act]ggtaaa|tttacc[agt]t/ig'
	'match|/ag[act]gtaaa|tttac[agt]ct/ig'
	'match|/agg[act]taaa|ttta[agt]cct/ig'
	'match|/aggg[acg]aaa|ttt[cgt]ccct/ig'
	'match|/agggt[cgt]aa|tt[acg]accct/ig'
	'match|/agggta[cgt]a|t[acg]taccct/ig'
	'match|/agggtaa[cgt]|[acg]ttaccct/ig'
	"replace|/tHa[Nt]/g|'<4>'"
	"replace|/aND|caN|Ha[DS]|WaS/g|'<3>'"
	"replace|/a[NSt]|BY/g|'<2>'"
	"replace|/<[^>]*>/g|'|'"
	"replace|/\\|[^|][^|]*\\|/g|'-'"
)
index=0
for operation in "${operations[@]}"; do
	kind=${operation%%|*}; rest=${operation#*|}
	source="bench/holdout/patterns/op$index.a"
	{
		echo "import { panic, programArguments, readTextFile } from 'adamic';"
		echo "const input = readTextFile(programArguments()[0] ?? panic('the input file must be the first argument'));"
		echo "if (input.kind === 'Error') panic(input.message);"
		echo "const data = input.text.replace(/^>.*\\n|\\n/mg, '');"
		case $kind in
			base) echo "console.log(\`\${data.length}\`);" ;;
			match) echo "const m = data.match($rest);"; echo "console.log(\`\${m ? m.length : 0}\`);" ;;
			replace) pattern=${rest%|*}; with=${rest##*|}; echo "console.log(\`\${data.replace($pattern, $with).length}\`);" ;;
		esac
	} > "$source"
	"$T/adamic" build "$source" -o "$T/op$index" > "$T/b$index.log" 2>&1 || { echo "FAIL build op$index ($operation): $(head -c 300 "$T/b$index.log")"; index=$((index + 1)); continue; }
	want=$(node --disable-warning=ExperimentalWarning oracle/node.mjs "$source" "$T/in.txt" 2>&1)
	got=$("$T/op$index" "$T/in.txt" 2>&1)
	[ "$got" = "$want" ] || { echo "FAIL op$index output: native $got, Node $want"; index=$((index + 1)); continue; }
	for round in 1 2 3; do
		/usr/bin/time -f "%U" -o "$T/n$index.$round" "$T/op$index" "$T/in.txt" > /dev/null
		/usr/bin/time -f "%U" -o "$T/j$index.$round" node --disable-warning=ExperimentalWarning oracle/node.mjs "$source" "$T/in.txt" > /dev/null
	done
	native=$(cat "$T"/n$index.* | sort -n | sed -n 2p); node=$(cat "$T"/j$index.* | sort -n | sed -n 2p)
	printf 'op%-2d %-8s native %6ss  node %6ss  prints %-10s %s\n' "$index" "$kind" "$native" "$node" "$got" "${rest:-cleanup only}"
	index=$((index + 1))
done
echo "== perf annotate, adamic_regex_regular_find over the whole port at 1M, the hottest lines"
"$T/adamic" build bench/holdout/regexredux.a -o "$T/rr" > /dev/null 2>&1 || { echo "FAIL regexredux build"; exit 1; }
perf record -q -e cycles:u -F 2999 -o "$T/rr.data" "$T/rr" "$T/in.txt" > /dev/null 2>&1
perf report -q -i "$T/rr.data" --stdio --no-children --sort symbol --percent-limit 0.3 2>/dev/null | head -12 | cut -c1-140
perf annotate -q -i "$T/rr.data" --stdio -s adamic_regex_regular_find 2>/dev/null | awk '$1 ~ /^[0-9.]+$/ && $1 + 0 >= 0.8' | sort -rn | head -30 | cut -c1-160
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
