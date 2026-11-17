#!/usr/bin/env bash
# #rv75m9k road A's gates on the adapted tree at this commit, before and after the adaptation, for the files named:
# each file's full latent census (LATENT_FULL=1, LATENT_ONLY_FILE), paired by compare.py. With --whole it also checks
# `adamic types` over the 79 roots, the emitted JavaScript of every touched file (transpileModule) and the declarations
# of the 79 roots emitted with stripInternal, each before against after, and it runs the restoring mutant last.
# Usage: bash gate.sh [--whole] <src/compiler/file.ts>...   Sealed: it sources the image's environment itself.
set -uo pipefail
source /opt/adamic-block/environment.sh && source cloud/sealed-job.sh || { echo "FAIL sealed environment"; exit 1; }
here=$(cd "$(dirname "$0")" && pwd)
whole=0; [ "${1:-}" = --whole ] && { whole=1; shift; }
sealed_typescript_source >/dev/null && sealed_npm stage3/api >/dev/null && echo "PASS sealed inputs" || { echo "FAIL sealed inputs"; exit 1; }
export STAGE3_PRODUCT_STORE="$TMPDIR/store" STAGE3_CACHE="$TMPDIR/stage3-cache" NODE_PATH="$PWD/stage3/api/node_modules"
tree="$TMPDIR/tree"
bash stage3/apply.sh --build-product "$tree" > "$TMPDIR/apply.log" 2>&1 && echo "PASS adapted tree built" || { echo "FAIL apply"; tail -30 "$TMPDIR/apply.log"; exit 1; }
overlay="$TMPDIR/overlay"
{ python3 stage3/census/latent/make_overlay.py "$PWD" "$overlay" && gofmt -w "$overlay"/*.go && go build -buildvcs=false -overlay="$overlay/overlay.json" -o "$TMPDIR/latent-census" ./stage3/census/latent/tool; } > "$TMPDIR/census-build.log" 2>&1 && echo "PASS census tool built" || { echo "FAIL census tool"; tail -20 "$TMPDIR/census-build.log"; exit 1; }
if [ "$whole" = 1 ]; then
	go build -buildvcs=false -o "$TMPDIR/adamic" ./cmd/adamic > "$TMPDIR/adamic-build.log" 2>&1 && echo "PASS adamic built" || { echo "FAIL adamic build"; tail -20 "$TMPDIR/adamic-build.log"; exit 1; }
fi
work="$TMPDIR/adaptation"; mkdir -p "$work"; cp "$here/adapt.cjs" "$here/mutant.cjs" "$work/"
node "$here/scan.cjs" "$tree" "$here/names.json" "$TMPDIR/before.json" > /dev/null || { echo "FAIL scan"; exit 1; }
node "$here/owners.cjs" "$tree" "$TMPDIR/before.json" "$work/owners.json" | head -1
touched=$(node -e 'console.log([...new Set(require(process.argv[1]).map(o => o.file))].join(" "))' "$work/owners.json")
roots=$(cd "$tree" && ls src/compiler/*.ts src/compiler/*/*.ts src/compiler/*/*/*.ts 2>/dev/null | grep -v '\.d\.ts$')
echo "roots: $(echo $roots | wc -w)"

measure() { # measure <label>
	local label=$1 file
	for file in $files; do
		LATENT_FULL=1 LATENT_ONLY_FILE="$tree/$file" LATENT_ASSERT_NO_OUTPUT=1 timeout 1500 "$TMPDIR/latent-census" "$tree/src/compiler" "$TMPDIR/$label.$(echo "$file" | tr / _).jsonl" > "$TMPDIR/$label.$(echo "$file" | tr / _).log" 2>&1
		echo "census $label $file exit=$? $(tail -1 "$TMPDIR/$label.$(echo "$file" | tr / _).log" | cut -c1-200)"
	done
	if [ "$whole" = 1 ]; then
		( cd "$tree" && "$TMPDIR/adamic" types $roots ) > "$TMPDIR/$label.types.out" 2> "$TMPDIR/$label.types.err"
		echo "types $label exit=$? declarations=$(wc -l < "$TMPDIR/$label.types.out") diagnostics=$(wc -l < "$TMPDIR/$label.types.err")"
		node -e '
const ts = require("typescript"), fs = require("node:fs"), crypto = require("node:crypto"), path = require("node:path");
const [tree, out, ...touched] = process.argv.slice(1);
const lines = [];
for (const file of touched) lines.push("js " + crypto.createHash("sha256").update(ts.transpileModule(fs.readFileSync(path.join(tree, file), "utf8"), { compilerOptions: { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS } }).outputText).digest("hex") + " " + file);
const config = ts.getParsedCommandLineOfConfigFile(path.join(tree, "src/compiler/tsconfig.json"), {}, { ...ts.sys, onUnRecoverableConfigFileDiagnostic: () => {} });
const options = { ...config.options, noEmit: false, declaration: true, emitDeclarationOnly: true, stripInternal: true, outDir: undefined, declarationDir: undefined, composite: false, incremental: false, typeRoots: [path.join(process.env.NODE_PATH, "@types")] };
const program = ts.createProgram(config.fileNames, options);
const declarations = new Map();
// A declaration map records source columns, which a retyped member moves; the declarations themselves are the API.
program.emit(undefined, (name, text) => { if (!name.endsWith(".map")) declarations.set(path.relative(tree, name), text); });
for (const [name, text] of [...declarations].sort()) lines.push("d.ts " + crypto.createHash("sha256").update(text).digest("hex") + " " + name);
lines.push("stock diagnostics " + ts.getPreEmitDiagnostics(program).length);
fs.writeFileSync(out, lines.join("\n") + "\n");
' "$tree" "$TMPDIR/$label.emit.txt" $touched
		echo "emit $label: $(grep -c '^js ' "$TMPDIR/$label.emit.txt") files transpiled, $(grep -c '^d.ts ' "$TMPDIR/$label.emit.txt") declaration files, $(grep '^stock' "$TMPDIR/$label.emit.txt")"
	fi
}
files="$*"
echo "== before"; measure before
node "$work/adapt.cjs" "$tree" | tr -d '\n ' && echo && echo "PASS adapt"
echo "== after"; measure after
echo "== compare"
for file in $files; do
	base="$(echo "$file" | tr / _).jsonl"
	python3 "$here/compare.py" "$file" "$TMPDIR/before.$base" "$TMPDIR/after.$base"
done
if [ "$whole" = 1 ]; then
	sort "$TMPDIR/before.types.err" > "$TMPDIR/b.err"; sort "$TMPDIR/after.types.err" > "$TMPDIR/a.err"
	cmp -s "$TMPDIR/b.err" "$TMPDIR/a.err" && echo "PASS adamic types diagnostics identical ($(wc -l < "$TMPDIR/a.err") lines)" || { echo "DIFF adamic types diagnostics:"; diff "$TMPDIR/b.err" "$TMPDIR/a.err" | head -60; }
	cmp -s "$TMPDIR/before.emit.txt" "$TMPDIR/after.emit.txt" && echo "PASS emitted JavaScript of touched files and stripInternal declarations identical" || { echo "DIFF emit:"; diff "$TMPDIR/before.emit.txt" "$TMPDIR/after.emit.txt" | head -40; }
	node "$work/mutant.cjs" "$tree" > /dev/null && echo "PASS mutant ran"
fi
echo "== sample before record"
head -c 600 "$TMPDIR/before.$(echo "${files%% *}" | tr / _).jsonl"; echo
echo "== end"
