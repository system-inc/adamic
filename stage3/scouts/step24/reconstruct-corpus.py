#!/usr/bin/env python3
"""Reproduce the historical parser corpus and reject drift against reference.json.

Requires the pinned local TypeScript git mirror and npm API from stage3 setup.
All subprocess output is retained in the new output directory. No target changes.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[2]
ADAPT_PIN = 'a3ef0dc93d5b2a6cf58f74669c763dc83a1aad0e'
SOURCE_PIN = '050880ce59e30b356b686bd3144efe24f875ebc8'


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def check_dump(path):
    reference = json.loads((REPO / 'stage3/drivers/parser/reference.json').read_text())
    actual = {'bytes': path.stat().st_size, 'sha256': digest(path)}
    expected = {key: reference[key] for key in actual}
    if actual != expected:
        raise ValueError(f'parser dump drift: expected {expected}, actual {actual}')
    return actual


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('output', type=Path, help='new scratch directory')
    parser.add_argument('--cache', type=Path, default=Path(os.environ.get('STAGE3_CACHE', str(Path.home() / '.cache/adamic-stage3'))))
    args = parser.parse_args()
    out = args.output.resolve()
    out.mkdir()  # Refuse existing output, including an already adapted tree.
    tree = out / 'corpus'
    api = args.cache.resolve() / 'api/node_modules'
    env = dict(os.environ, NODE_PATH=str(api), PARSER_TYPESCRIPT=str(api / 'typescript/lib/typescript.js'), PARSER_RUNTIME=str(REPO / 'oracle/adamic.mjs'))

    def run(name, command, cwd=REPO):
        with (out / (name + '.log')).open('wb') as log:
            subprocess.run(command, cwd=cwd, env=env, stdout=log, stderr=subprocess.STDOUT, check=True)

    run('clone', ['git', 'clone', '--quiet', str(args.cache.resolve() / 'typescript.git'), str(tree)])
    run('checkout', ['git', '-C', str(tree), 'checkout', '--quiet', SOURCE_PIN])
    adapters = {}
    for lane in ['00-setup', '10-type-imports']:
        name = f'stage3/adapt/{lane}/adapt.cjs'
        script = out / (lane + '.cjs')
        script.write_bytes(subprocess.check_output(['git', 'show', f'{ADAPT_PIN}:{name}'], cwd=REPO))
        adapters[name] = digest(script)
        run(lane, ['node', str(script), str(tree)])
    # The argument itself is embedded in generated source. It must be relative.
    run('generate', ['node', 'scripts/processDiagnosticMessages.mjs', 'src/compiler/diagnosticMessages.json'], cwd=tree)
    files = sorted(p.relative_to(tree / 'src/compiler').as_posix() for p in (tree / 'src/compiler').rglob('*') if p.is_file())
    manifest = out / 'manifest'
    manifest.write_text(''.join(p + '\n' for p in files))
    assert manifest.read_bytes() == (HERE / 'evidence/corpus.manifest').read_bytes(), 'corpus membership drift'
    hashes = {'src/compiler/' + name: digest(tree / 'src/compiler' / name) for name in files}
    expected_hashes = json.loads((HERE / 'evidence/reconstructed-input-hashes.json').read_text())
    def check_inputs():
        actual = {'src/compiler/' + name: digest(tree / 'src/compiler' / name) for name in files}
        if actual != expected_hashes:
            raise ValueError('corpus byte drift')
    check_inputs()
    driver = REPO / 'stage3/drivers/parser'
    parser_tree = out / 'parser'
    shutil.copytree(tree / 'src/compiler', parser_tree / 'src/compiler')
    for source, target in [('main.a', 'parser-proof-main.a'), ('kinds.a', 'kinds.a')]:
        shutil.copyfile(driver / source, parser_tree / target)

    def dump(name):
        with (out / (name + '.dump')).open('wb') as stdout, (out / (name + '.stderr')).open('wb') as stderr:
            subprocess.run(['node', '--disable-warning=ExperimentalWarning', str(driver / 'node.mjs'), str(parser_tree / 'parser-proof-main.a'), str(tree), str(manifest)], env=env, stdout=stdout, stderr=stderr, check=True)
        assert not (out / (name + '.stderr')).stat().st_size
        return out / (name + '.dump')

    result = check_dump(dump('node'))
    generated = tree / 'src/compiler/diagnosticInformationMap.generated.ts'
    original = generated.read_bytes()
    offset = original.index(b"DiagnosticCategory")
    mutant = bytearray(original)
    assert mutant[offset] == ord('D')
    mutant[offset] = ord('X')  # Exactly one ASCII byte in the first import identifier.
    assert sum(a != b for a, b in zip(original, mutant)) == 1
    generated.write_bytes(mutant)
    try:
        mutant_dump = dump('one-byte-mutant')
        try:
            check_dump(mutant_dump)
        except ValueError as error:
            caught = str(error)
        else:
            raise AssertionError('one-byte corpus mutant survived')
        try:
            check_inputs()
        except ValueError:
            pass
        else:
            raise AssertionError('input hash guard mutant survived')
    finally:
        generated.write_bytes(original)
    report = {'source_pin': SOURCE_PIN, 'adaptation_pin': ADAPT_PIN, 'adapters': adapters, 'files': len(files), 'reference': result,
              'mutant': {'file': 'src/compiler/diagnosticInformationMap.generated.ts', 'offset_zero_based': offset, 'before_hex': '44', 'after_hex': '58', 'changed_bytes': 1, 'caught': caught, 'input_hash_guard': 'caught'}}
    (out / 'report.json').write_text(json.dumps(report, indent=2) + '\n')
    print(json.dumps(report, indent=2))


if __name__ == '__main__':
    main()
