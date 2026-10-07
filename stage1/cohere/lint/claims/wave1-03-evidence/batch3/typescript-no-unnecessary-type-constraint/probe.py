#!/usr/bin/env python3
"""Prove repair-shape blockers using unchanged cohere rule bodies and scratch Go overlays."""
import argparse
import json
import subprocess
from pathlib import Path

owned = Path(__file__).resolve().parent
repo = owned.parents[6]
parser = argparse.ArgumentParser()
parser.add_argument('--scratch', required=True, type=Path)
args = parser.parse_args()
scratch = args.scratch.resolve()
scratch.mkdir(parents=True, exist_ok=True)
cohere = repo / 'cohere'
subjects = [
    ('typescript-no-unnecessary-type-constraint', 'NoUnnecessaryTypeConstraint', 'unexpected suggestion shape'),
    ('typescript-prefer-as-const', 'PreferAsConst', 'unexpected fix shape'),
    ('typescript-prefer-enum-initializers', 'PreferEnumInitializers', 'unexpected suggestion shape'),
]

def command(arguments, log, cwd=repo):
    with log.open('w') as output:
        result = subprocess.run(arguments, cwd=cwd, stdout=output, stderr=subprocess.STDOUT)
    print('exit=' + str(result.returncode) + ' log=' + str(log), flush=True)
    return result.returncode

registry = scratch / 'registry.go'
registry.write_text('package main\nimport "github.com/system-inc/cohere/internal/lint/rule"\n'
    + 'type registered struct { subject rule.Rule; options func([]string) any }\n'
    + 'func registeredRules() []registered { return []registered{\n'
    + ''.join('{oracle' + name + '(), oracle' + name + 'Options},\n' for _, name, _ in subjects)
    + '} }\n')
original = (repo / 'stage1/cohere/lint/testdata/oracle.go').read_text()
# The diagnostic shape printer replaces only serialization. Parsing, rule execution,
# formatting and converging fixes remain the independent shared Go oracle's code.
start = original.index('\t\trepair, replacement, suggestion :=')
end = original.index('\n\t}\n\tresult, err := edit.FixText', start)
full = original[:start] + '''
        fmt.Fprintf(out, "shape %d %d %s fixes=%d suggestions=%d\\n", start, end, d.Message.Id, len(d.Fixes), len(d.Suggestions))
        for _, fix := range d.Fixes {
            fmt.Fprintf(out, "fix %d %d %q\\n", fix.Range.Pos(), fix.Range.End(), fix.Text)
        }
        for _, suggestion := range d.Suggestions {
            fmt.Fprintf(out, "suggestion %s %q\\n", suggestion.Message.Id, suggestion.Message.Description)
            for _, fix := range suggestion.Fixes {
                fmt.Fprintf(out, "edit %d %d %q\\n", fix.Range.Pos(), fix.Range.End(), fix.Text)
            }
        }''' + original[end:]
probe = scratch / 'full-shape.go'
probe.write_text(full)

for label, oracle in [('legacy', repo / 'stage1/cohere/lint/testdata/oracle.go'), ('full', probe)]:
    files = [('oracle', oracle), ('registry', registry)]
    files += [(slug.replace('-', '_'), owned.parent / slug / 'oracle.go') for slug, _, _ in subjects]
    overlay = scratch / (label + '-overlay.json')
    virtual = [(str(cohere / ('adamic_shape_' + name + '.go')), str(path)) for name, path in files]
    overlay.write_text(json.dumps({'Replace': dict(virtual)}, indent=2))
    binary = scratch / (label + '-oracle')
    if command(['go', 'build', '-overlay=' + str(overlay), '-o', str(binary)] + [name for name, _ in virtual], scratch / (label + '-build.log'), cohere):
        raise SystemExit('oracle build failed; no shape evidence')

expected = {
    'NoUnnecessaryTypeConstraint': ['shape 14 15 noUnnecessaryTypeConstraint fixes=0 suggestions=1',
        'edit 15 27 ""', 'suggestion removeTheConstraint "Remove the constraint that does nothing."'],
    'PreferAsConst': ['shape 9 14 preferAsConst fixes=2 suggestions=0', 'fix 7 14 ""', 'fix 22 22 " as const"',
        "fixed\tlet foo = 'bar' as const;\\u000a"],
    'PreferEnumInitializers': ['shape 17 19 defineInitializer fixes=0 suggestions=3',
        'edit 17 19 "Up = 0"', 'edit 17 19 "Up = 1"', 'edit 17 19 "Up = \'Up\'"'],
}
for slug, name, panic in subjects:
    directory = scratch / slug
    directory.mkdir(exist_ok=True)
    witness = owned.parent / slug / 'testdata/repair-shape.ts.txt'
    # Give the raw witness the real TypeScript filename expected by the parser.
    source = directory / 'repair-shape.ts'
    source.write_bytes(witness.read_bytes())
    manifest = directory / 'manifest.txt'
    manifest.write_text(str(source) + '\t@typescript-eslint/' + slug.removeprefix('typescript-') + '\n')
    legacy_log = directory / 'legacy.txt'
    status = command([str(scratch / 'legacy-oracle'), '--manifest', str(manifest)], legacy_log)
    if status == 0 or ('panic: ' + panic) not in legacy_log.read_text():
        raise SystemExit('expected legacy shape refusal absent for ' + name)
    full_log = directory / 'full.txt'
    if command([str(scratch / 'full-oracle'), '--manifest', str(manifest)], full_log):
        raise SystemExit('independent full diagnostic observation failed for ' + name)
    observed = full_log.read_text()
    if any(line not in observed for line in expected[name]):
        raise SystemExit('unexpected independently observed repair shape for ' + name)
    if command(['go', 'test', './internal/lint/rules/typescript', '-count=1', '-v', '-timeout=5m', '-run', '^Test' + name], directory / 'upstream.txt', cohere):
        raise SystemExit('unchanged upstream tests failed for ' + name)
print('PASS: all three legacy oracle refusals and complete Go repair observations reproduced', flush=True)
