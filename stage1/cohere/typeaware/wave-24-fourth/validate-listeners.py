#!/usr/bin/env python3
"""Compare numeric listener declarations with the pinned Go parser enum."""
import argparse
import json
from pathlib import Path
import subprocess

OWN = Path(__file__).resolve().parent
ROOT = OWN.parents[3]

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('directory', type=Path)
    parser.add_argument('--stage0', required=True, type=Path)
    parser.add_argument('--checker', required=True, type=Path)
    args = parser.parse_args()
    dest = args.directory.resolve()
    dest.mkdir(parents=True, exist_ok=True)
    def run(label, command):
        with (dest/(label+'.stdout')).open('wb') as out, (dest/(label+'.stderr')).open('wb') as err:
            result = subprocess.run([str(value) for value in command], cwd=ROOT, stdout=out, stderr=err)
        assert result.returncode == 0, (label, result.returncode)
        return (dest/(label+'.stdout')).read_bytes()
    truth = run('go', ['go','run',OWN/'listener_oracle.go'])
    binary = dest/'listeners'
    run('native-build',[args.stage0,'build',OWN/'listener-probe.a','-o',binary,'--tsgo',args.checker])
    got = run('native',[binary])
    assert got == truth, 'numeric listener drift'
    names = ['prefer-return-this-type', 'return-await', 'use-unknown-in-catch-callback-variable',
             'no-process-exit-after-output', 'no-uncleared-race-timeout', 'require-blocking-standard-streams',
             'prefer-promise-reject-errors', 'prefer-regex-literals', 'prefer-rest-params']
    declarations = []
    for name in names:
        metadata = json.loads((OWN/'listeners'/name/'rule.json').read_text())
        assert metadata['name'] and all(type(kind) is int for kind in metadata['kinds'])
        declarations.append(','.join(str(kind) for kind in metadata['kinds']))
    encoded = ('\n'.join(declarations)+'\n').encode()
    assert encoded == truth, 'rule.json numeric listener drift'
    mutant_json = dest/'mutant-rule.json'
    metadata = json.loads((OWN/'listeners/prefer-rest-params/rule.json').read_text())
    metadata['kinds'] = [214]
    mutant_json.write_text(json.dumps(metadata))
    changed = declarations[:-1]+[','.join(str(kind) for kind in json.loads(mutant_json.read_text())['kinds'])]
    assert ('\n'.join(changed)+'\n').encode() != truth, 'JSON listener mutant survived'
    original = OWN.parent/'wave-24-third/prefer_rest_params.a'
    source = original.read_text()
    needle = 'SyntaxKinds: number[] = [79]'
    assert source.count(needle) == 1
    source = source.replace(needle, 'SyntaxKinds: number[] = [214]')
    source = source.replace("'../", "'"+str(OWN.parent)+'/').replace("'./", "'"+str(original.parent)+'/')
    mutant = dest/'mutant-rule.a'
    mutant.write_text(source)
    probe = dest/'mutant-probe.a'
    source = (OWN/'listener-probe.a').read_text().replace("'../", "'"+str(OWN.parent)+'/')
    source = source.replace(str(original), str(mutant))
    probe.write_text(source)
    binary = dest/'mutant'
    run('mutant-build',[args.stage0,'build',probe,'-o',binary,'--tsgo',args.checker])
    got = run('mutant',[binary])
    assert got != truth, 'runnable listener mutant survived'
    print('PASS nine native and rule.json declarations match pinned numeric SyntaxKind; native and JSON Identifier-to-CallExpression mutants caught')

if __name__ == '__main__':
    main()
