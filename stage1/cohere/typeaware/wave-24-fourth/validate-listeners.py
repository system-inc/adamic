#!/usr/bin/env python3
"""Compare numeric listener declarations with the pinned Go parser enum."""
import argparse
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
    print('PASS nine declarations match pinned numeric SyntaxKind; runnable Identifier-to-CallExpression mutant caught')

if __name__ == '__main__':
    main()
