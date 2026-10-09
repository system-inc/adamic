"""Scope the existing no-output census to the 26 surveyed roots.

The historical refusal rewriter cannot collect two newer outer metadata hooks.
Record those errors before invoking that unchanged visitor rewriter. Production
Lower and Load remain disabled by the existing census overlay.
"""
import gzip
import json
from pathlib import Path
import runpy
import subprocess
import sys

repository, scratch = map(lambda value: Path(value).resolve(), sys.argv[1:3])
# review/compiler/generic-values -> repository root
root = Path(__file__).resolve().parents[3]
territory = root / 'stage3/census/latent'
original_run = subprocess.run
scratch.mkdir(parents=True, exist_ok=True)

def invoke(args, *positionals, **kwargs):
    if isinstance(args, list) and any('refusalrewrite/cmd' in str(arg) for arg in args):
        args = list(args)
        index = args.index('-input') + 1
        source = Path(args[index]).read_text()
        for hook in ['mergedFieldRulings', 'functionAnnotations']:
            block = '\tif err := l.' + hook + '(module); err != nil {\n\t\treturn err\n\t}'
            assert source.count(block) == 1, ('outer metadata hook changed', hook)
            source = source.replace(block, block.replace('return err', 'latentRecord(err)'))
        supplied = scratch / 'refusal-input.go.txt'
        supplied.write_text(source)
        args[index] = str(supplied)
    return original_run(args, *positionals, **kwargs)

subprocess.run = invoke
sys.argv = [str(territory / 'make_overlay.py'), str(repository), str(scratch)]
runpy.run_path(str(territory / 'make_overlay.py'), run_name='__main__')
subprocess.run = original_run

ledger = json.load(gzip.open(root / 'review/compiler/lowering-chain/source-member-4/step-16-generics/baseline.json.gz', 'rt'))
selected = sorted({row['unit'].split('/src/compiler/')[-1] for row in ledger['findings'] if row['reason'] == 'a generic function as a value'})
assert len(selected) == 26
units = scratch / 'internal_lower_latent_units.go'
source = units.read_text().replace('"fmt"', '"fmt"\n"strings"', 1)
literal = 'map[string]bool{' + ','.join(json.dumps(name) + ':true' for name in selected) + '}'
needle = '\t\tfor _, candidate := range latentCandidates(program, file) {\n\t\t\tnode := candidate.node'
assert source.count(needle) == 1
source = source.replace(needle, needle + '\n selected := ' + literal + '\n parts := strings.Split(program.Where(node), "/src/compiler/")\n if !selected[parts[len(parts)-1]] { continue }')
files = sorted({name.rsplit(':', 2)[0] for name in selected})
file_literal = 'map[string]bool{' + ','.join(json.dumps(name) + ':true' for name in files) + '}'
file_needle = '\tfor _, file := range program.Files() {'
assert source.count(file_needle) == 1
source = source.replace(file_needle, file_needle + '\n files := ' + file_literal + '\n filename := strings.Split(program.FileName(file), "/src/compiler/")\n if !files[filename[len(filename)-1]] { continue }')
units.write_text(source)
manifest = json.loads((scratch / 'overlay.json').read_text())
if repository != root:
    # Restore the dependency's changed lower files in the real build, then remap
    # all measurement hooks generated against its scratch source view.
    replacements = {}
    changed = subprocess.check_output(['git', 'diff', '--name-only', '50654a40', '--', 'internal/lower'], cwd=root, text=True).splitlines()
    for name in changed:
        if name.endswith('.go') and (repository / name).exists(): replacements[str(root / name)] = str(repository / name)
    empty = scratch / 'empty.go.txt'; empty.write_text('package lower\n')
    for name in subprocess.check_output(['git', 'ls-files', '--cached', '--others', '--exclude-standard', 'internal/lower'], cwd=root, text=True).splitlines():
        if name.endswith('.go') and not name.endswith('_test.go') and not (repository / name).exists():
            replacements[str(root / name)] = str(empty)
    for name, replacement in manifest['Replace'].items():
        replacements[str(root / Path(name).relative_to(repository))] = replacement
    manifest['Replace'] = replacements
(scratch / 'overlay.json').write_text(json.dumps(manifest, indent=2) + '\n')
(scratch / 'selected.json').write_text(json.dumps(selected, indent=2) + '\n')
print('selected roots:', len(selected))
