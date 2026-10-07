#!/usr/bin/env python3
"""Reproduce the JSX parser blocker without modifying shared files."""
import argparse
import json
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[4]
OWN = Path(__file__).resolve().parent

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('directory', type=Path)
    parser.add_argument('--stage0', required=True, type=Path)
    args = parser.parse_args()
    dest = args.directory.resolve()
    dest.mkdir(parents=True, exist_ok=True)
    def run(label, command, cwd=ROOT):
        with (dest/(label+'.stdout')).open('wb') as out, (dest/(label+'.stderr')).open('wb') as err:
            result = subprocess.run([str(item) for item in command], cwd=cwd, stdout=out, stderr=err)
        return result.returncode, (dest/(label+'.stdout')).read_bytes(), (dest/(label+'.stderr')).read_bytes()
    native = dest/'parser'
    assert run('native-build', [args.stage0, 'build', OWN/'parser-probe.a', '-o', native])[0] == 0
    control = OWN/'testdata/control.txt'
    jsx = OWN/'testdata/static-components.txt'
    assert run('native-control', [native, control]) == (0, b'15\n', b'')
    code, out, err = run('native-jsx', [native, jsx])
    assert code == 70 and not out and b'expected GreaterThanToken, got SlashToken' in err
    source = dest/'input.tsx'
    source.write_text(jsx.read_text())
    manifest = dest/'manifest'
    manifest.write_text(str(source)+'\n')
    config = dest/'tsconfig.json'
    config.write_text(json.dumps({'compilerOptions': {'jsx':'preserve','target':'ES2022'}, 'files':[str(source)]}))
    virtual = ROOT/'cohere/adamic_wave24_fourth_oracle.go'
    overlay = dest/'overlay.json'
    overlay.write_text(json.dumps({'Replace':{str(virtual):str(OWN/'testdata/oracle.go')}}))
    oracle = dest/'oracle'
    assert run('oracle-build', ['go','build','-overlay',overlay,'-o',oracle,virtual], ROOT/'cohere')[0] == 0
    code, out, err = run('oracle-jsx', [oracle,config,manifest])
    assert code == 0 and b'react-hooks/static-components\tstaticComponents\t' in out and out.endswith(b'findings 1\n')
    # A runnable probe mutant silently substitutes the ordinary TS control for JSX.
    # The independent Go finding and missing refusal expose this false measurement.
    mutant = dest/'mutant.a'
    mutant.write_text((OWN/'parser-probe.a').read_text().replace('source.text', json.dumps(control.read_text())).replace('../../../typescript/parser/', str(ROOT/'stage1/typescript/parser')+'/'))
    binary = dest/'mutant'
    assert run('mutant-build', [args.stage0,'build',mutant,'-o',binary])[0] == 0
    code, out, err = run('mutant-jsx', [binary,jsx])
    assert (code, out, err) == (0,b'15\n',b'')
    assert code != 70, 'substitution mutant must fail the required JSX refusal check'
    print('PASS ordinary TypeScript; Go JSX parses and reports 1 finding; native JSX refusal 70; runnable probe mutant caught')

if __name__ == '__main__':
    main()
