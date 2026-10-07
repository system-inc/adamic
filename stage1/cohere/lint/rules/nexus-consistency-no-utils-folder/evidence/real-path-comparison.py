import json, os, pathlib, re, shutil, subprocess, time
repo = pathlib.Path('/workspace/adamic')
lint = repo / 'stage1/cohere/lint'
root = pathlib.Path('/workspace/wave-25-no-utils-manual')
root.mkdir(exist_ok=True)

def run(name, args, cwd=repo):
    with (root / (name + '.stdout')).open('wb') as out, (root / (name + '.stderr')).open('wb') as err:
        start = time.perf_counter()
        result = subprocess.run([str(x) for x in args], cwd=cwd, stdout=out, stderr=err)
        elapsed = time.perf_counter() - start
    assert result.returncode == 0, (name, result.returncode, (root / (name + '.stderr')).read_text())
    return (root / (name + '.stdout')).read_bytes(), elapsed

replacements = {}
virtual = []
for name, source in [('oracle', lint / 'testdata/oracle.go'), ('registry', lint / '.generated/registry.go')] + [(p.parent.name.replace('-', '_'), p) for p in sorted((lint / 'rules').glob('*/oracle.go'))]:
    target = repo / 'cohere' / ('adamic_lint_' + name + '.go')
    replacements[str(target)] = str(source)
    virtual.append(str(target))
overlay = root / 'overlay.json'
overlay.write_text(json.dumps({'Replace': replacements}))
run('oracle-build', ['go', 'build', '-overlay=' + str(overlay), '-o', root / 'oracle', *virtual], repo / 'cohere')
rows = []
for directory, count in [('utils', 1), ('_utils', 1), ('_utils/utils', 2), ('utilities', 0), ('_utilities', 0), ('utilsomething', 0), ('myutils', 0)]:
    source = root / directory / 'Thing.ts'
    source.parent.mkdir(parents=True, exist_ok=True)
    source.write_text('export const Value = 1;\n')
    for selection in ['nexus/consistency-no-utils-folder', 'all']:
        rows.append(str(source) + '\t' + selection)
    rows.append(str(source) + '\tnexus/consistency-no-utils-folder\t\t\tfalse\t{"ignored":true}')
manifest = root / 'manifest.tsv'
manifest.write_text('\n'.join(rows) + '\n')
truth, _ = run('truth', [root / 'oracle', '--manifest', manifest])
count, _ = run('truth-count', [root / 'oracle', '--manifest', manifest, '--count'])
assert count == b'12\n', count
compiler = '/workspace/wave-25-landing-second/fifth/adamic'
for label, directory in [('normal', lint), ('mutant', root / 'mutant')]:
    if label == 'mutant':
        for source in lint.rglob('*'):
            if not source.is_file() or source.suffix not in ['.ts', '.a', '.json', '.go', '.txt']:
                continue
            if 'testdata' in source.parts and source.suffix != '.txt':
                continue
            target = directory / source.relative_to(lint)
            target.parent.mkdir(parents=True, exist_ok=True)
            text = source.read_text()
            if source.suffix in ['.ts', '.a']:
                def rewrite(match):
                    module = match.group(2)
                    resolved = (source.parent / module).resolve()
                    if module.startswith('.') and not resolved.is_relative_to(lint):
                        return match.group(1) + str(resolved) + match.group(3)
                    return match.group(0)
                text = re.sub(r"(from ['\"])([^'\"]+)(['\"])", rewrite, text)
            target.write_text(text)
        module = directory / 'rules/nexus-consistency-no-utils-folder/rule.a'
        metadata = json.loads((lint / 'rules/nexus-consistency-no-utils-folder/mutant.json').read_text())
        text = module.read_text()
        assert text.count(metadata['from']) == 1
        module.write_text(text.replace(metadata['from'], metadata['to']))
        run('mutant-registry', ['go', 'run', './cmd/lint-registry', '-root', str(directory)])
    entry = directory / 'main.ts'
    binary = root / (label + '-native')
    run(label + '-build', [compiler, 'build', entry, '-o', binary, '--sanitize'])
    js, _ = run(label + '-emit', [compiler, 'js', entry])
    emitted = root / (label + '.mjs')
    emitted.write_bytes(js)
    for side, args in [('Node', ['node', '--disable-warning=ExperimentalWarning', repo / 'oracle/node.mjs', entry, '--manifest', manifest]), ('emitted', ['node', '--disable-warning=ExperimentalWarning', repo / 'oracle/node.mjs', emitted, '--manifest', manifest]), ('native', [binary, '--manifest', manifest])]:
        output, elapsed = run(label + '-' + side, args)
        assert (output == truth) == (label == 'normal'), (label, side)
        if label == 'mutant':
            assert (root / (label + '-' + side + '.stderr')).read_bytes() == b''
        print(label, side, 'Go bytes agree' if label == 'normal' else 'compiled, exit 0, empty stderr; Go bytes catch mutant', len(output), elapsed, flush=True)
print('real paths: 21 selected/all/options rows, 12 findings; clean near-matches preserved', flush=True)
