#!/usr/bin/env bash
set -euo pipefail
here=$(cd "$(dirname "$0")" && pwd)
repo=$(cd "$here/../../.." && pwd)
if [ "$#" -lt 2 ]; then
    echo 'usage: run.sh <parser-tree> <new-output-directory> [native-binary] [--inputs <corpus-tree>]' >&2
    exit 2
fi
tree=$(realpath "$1")
out=$(realpath -m "$2")
shift 2
inputs="$tree"
native=''
while [ "$#" -gt 0 ]; do
    if [ "$1" = '--inputs' ] && [ "$#" -ge 2 ]; then
        inputs=$(realpath "$2")
        shift 2
    elif [ -z "$native" ] && [ "$1" != '--inputs' ]; then
        native=$(realpath "$1")
        shift
    else
        echo 'invalid arguments' >&2
        exit 2
    fi
done
[ ! -e "$out" ] || { echo "refusing existing output: $out" >&2; exit 2; }
mkdir -p "$out"
export PARSER_TYPESCRIPT=${PARSER_TYPESCRIPT:-${STAGE3_CACHE:-$HOME/.cache/adamic-stage3}/api/node_modules/typescript/lib/typescript.js}
export PARSER_RUNTIME="$repo/oracle/adamic.mjs"
python3 - "$tree" "$out" "$here" "$inputs" <<'PY'
from pathlib import Path
import shutil, sys
root, out, here, inputs = map(Path, sys.argv[1:])
files = sorted(p.relative_to(inputs / 'src/compiler').as_posix() for p in (inputs / 'src/compiler').rglob('*') if p.is_file())
(out / 'manifest').write_text(''.join(p + '\n' for p in files))
(out / 'source-manifest').write_text(''.join(str(inputs / 'src/compiler' / p) + '\n' for p in files if p.endswith('.ts')))
# Staging adds no files to src/compiler or to the corpus.
shutil.copyfile(here / 'main.a', root / 'parser-proof-main.a')
shutil.copyfile(here / 'kinds.a', root / 'kinds.a')
source = (here / 'main.a').read_text()
needle = 'function dump(node: Node): void {\n'
assert source.count(needle) == 1
source = source.replace(needle, 'let planted = false;\n' + needle + '    if(!planted && node.kind === SyntaxKind.Identifier) { node.end += 1; planted = true; }\n')
(root / 'parser-proof-mutant.a').write_text(source)
PY
node --disable-warning=ExperimentalWarning "$here/node.mjs" "$tree/parser-proof-main.a" "$inputs" "$out/manifest" > "$out/node.dump" 2> "$out/node.stderr"
node --disable-warning=ExperimentalWarning "$here/node.mjs" "$tree/parser-proof-mutant.a" "$inputs" "$out/manifest" > "$out/mutant.dump" 2> "$out/mutant.stderr"
set +e
cmp "$out/node.dump" "$out/mutant.dump" > "$out/mutant-comparison.log" 2>&1
comparison=$?
set -e
[ "$comparison" -eq 1 ] || { echo "expected mutant byte difference, cmp exited $comparison" >&2; exit 1; }
python3 - "$out" <<'PY'
import hashlib, json, sys
from pathlib import Path
out = Path(sys.argv[1])
rows = {}
for name in ['node', 'mutant']:
    data = (out / (name + '.dump')).read_bytes()
    rows[name] = {'bytes': len(data), 'sha256': hashlib.sha256(data).hexdigest(), 'exit': 0, 'stderr_bytes': (out / (name + '.stderr')).stat().st_size}
normal = (out / 'node.dump').read_text().splitlines()
mutant = (out / 'mutant.dump').read_text().splitlines()
assert len(normal) == len(mutant)
differences = [(index + 1, a, b) for index, (a, b) in enumerate(zip(normal, mutant)) if a != b]
assert len(differences) == 1, differences[:3]
line, before, after = differences[0]
a, b = before.split('\t')[0].split(), after.split('\t')[0].split()
assert a[0] == b[0] == 'Identifier' and a[1] == b[1] and int(b[2]) == int(a[2]) + 1 and a[3] == b[3]
assert before.split('\t')[1:] == after.split('\t')[1:]
rows['mutation'] = {'line': line, 'before': before, 'after': after, 'changed_lines': 1}
rows['files'] = len((out / 'manifest').read_text().splitlines())
rows['schema'] = 'parser-preorder-with-jsdoc-v2'
rows['coverage'] = {
    'jsdoc_nodes': sum(line.startswith('JSDoc ') for line in normal),
    'tag_nodes': sum('\ttagName ' in line for line in normal),
    'type_expressions': sum(line.startswith('JSDocTypeExpression ') for line in normal),
    'jsdoc_diagnostics': sum(line.startswith('jsDocDiagnostic ') for line in normal),
}
rows['mutant'] ['caught_by'] = 'cmp of complete preorder dump; both Node executions exited zero'
(out / 'report.json').write_text(json.dumps(rows, indent=2) + '\n')
print(json.dumps(rows, indent=2))
PY
if [ -n "$native" ]; then
    "$native" "$inputs" "$out/manifest" > "$out/native.dump" 2> "$out/native.stderr"
    cmp "$out/node.dump" "$out/native.dump" > "$out/native-comparison.log" 2>&1
    cmp "$out/node.stderr" "$out/native.stderr" >> "$out/native-comparison.log" 2>&1
fi
