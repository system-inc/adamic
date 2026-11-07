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
for source in colv/probes/*.a; do
	name=$(basename "$source" .a)
	mkdir -p "$work/$name"
	cp "$source" "$work/$name/p.a"; cp "$source" "$work/$name/p.ts"
	(cd "$work/$name" && timeout 60 node p.ts > node.out 2>node.err; echo "node exit=$? out=$(one < node.out) err=$(head -c 200 node.err | one)") | sed "s/^/PROBE $name /"
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
echo "== gofmt"; gofmt -l internal/lower internal/oracle
echo "== vet"; go vet ./internal/lower ./internal/oracle && echo vet-ok
echo "== lower Intl"; timeout 900 go test ./internal/lower -run 'Intl|SlotType|RefuseNode' -count=1 2>&1 | tail -150
echo "== oracle Intl"; timeout 900 go test ./internal/oracle -run 'Intl' -count=1 2>&1 | tail -80
echo "== job end"
