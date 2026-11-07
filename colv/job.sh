#!/bin/bash
# Scratch probe job for #zjndrx5 head 1's view rule. Never submitted.
set -u
root=$(pwd)
work=/tmp/colv-work
rm -rf "$work"; mkdir -p "$work"
echo "== build new"; go build -o "$work/new" ./cmd/adamic && echo built-new
cp internal/lower/library_intl.go "$work/tip_library_intl.go"
cp colv/base_library_intl.go.txt internal/lower/library_intl.go
echo "== build old"; go build -o "$work/old" ./cmd/adamic && echo built-old
cp "$work/tip_library_intl.go" internal/lower/library_intl.go
one() { tr '\n' '|' | cut -c1-300; }
mkdir -p "$work/node_modules/adamic"
printf '{"name":"adamic","type":"module","main":"index.mjs"}' > "$work/node_modules/adamic/package.json"
ln -s "$root/oracle/adamic.mjs" "$work/node_modules/adamic/index.mjs"
for source in colv/probes/*.a; do
	name=$(basename "$source" .a)
	mkdir -p "$work/$name"
	cp "$source" "$work/$name/p.a"; cp "$source" "$work/$name/p.ts"
	(cd "$work/$name" && timeout 60 node --disable-warning=ExperimentalWarning "$root/oracle/node.mjs" p.a > node.out 2>node.err; echo "node exit=$? out=$(one < node.out) err=$(head -c 200 node.err | one)") | sed "s/^/PROBE $name /"
	for binary in old new; do
		for extension in a ts; do
			(
				cd "$work/$name"
				timeout 120 "$work/$binary" build "p.$extension" -o "native-$binary-$extension" > build.out 2> build.err
				code=$?
				if [ $code -ne 0 ]; then
					echo "native compile exit=$code err=$(one < build.err)"
				else
					timeout 60 "./native-$binary-$extension" > run.out 2> run.err
					echo "native exit=$? out=$(one < run.out) err=$(head -c 200 run.err | one)"
				fi
				timeout 120 "$work/$binary" js "p.$extension" > "p-$binary-$extension.mjs" 2> js.err
				code=$?
				if [ $code -ne 0 ]; then
					echo "js compile exit=$code err=$(one < js.err)"
				else
					timeout 60 node "p-$binary-$extension.mjs" > jsrun.out 2> jsrun.err
					echo "js exit=$? out=$(one < jsrun.out) err=$(head -c 200 jsrun.err | one)"
				fi
			) | sed "s/^/PROBE $name $binary .$extension /"
		done
	done
done
COLV_LIGHT=1 bash colv/final.sh
echo "== admission delta"
python3 cloud/admission-corpus/manifest.py --sha HEAD > /tmp/colv-manifest.json && timeout 1200 go run ./cmd/adamic-admission-delta --base 4e562449b536fb8a3539c1adc6554e7d470b80c9 --head HEAD --manifest /tmp/colv-manifest.json --manifest-generator cloud/admission-corpus/manifest.py --corpus-filter diff --json > /tmp/colv-delta.json 2> /tmp/colv-delta.err; echo "delta exit=$?"
tail -20 /tmp/colv-delta.err
python3 -c "import json; r=json.load(open('/tmp/colv-delta.json')); print('verdict', r.get('verdict'), 'diff', r.get('diff_count'), 'stale', r.get('stale_runtime_stops')); [print(e['path'], e['class'], (e.get('compile') or {}).get('stderr','')[:200].replace(chr(10),' ')) for e in r.get('programs', [])]"
echo "== job end"
