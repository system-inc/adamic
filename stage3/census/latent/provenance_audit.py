"""Prove symbol-linked rollback echoes on real guarded lowering findings."""
import json
import os
from pathlib import Path
import subprocess
import sys

SOURCE = '''declare function external(): number;
export function rollback(): void {
    const poison: number = external();
    const secondary: number = poison;
    const holder = { poison };
    if (poison) { }
    if (secondary) { }
    { const poison: number = 7; if (poison) { } }
}
export function clean(poison: number): void { if (poison) { } }
'''

def check(records):
    findings = [f for r in records[1:] for f in r['findings'] if f['phase']=='lowering']
    root = next(f for f in findings if f['kind']=='NotYet' and 'without a body' in f['reason'] and not f.get('blocked_by') and f['unit'].endswith('probe.a:2:1'))
    echoes = [f for f in findings if f['kind']=='NotYet' and f['reason']=='reading poison']
    assert len(echoes)==3, ('poison echoes',echoes)
    signature = {k:root[k] for k in ['unit','phase','kind','where','reason','text']}
    for echo in echoes:
        assert echo.get('blocked_by')==signature, 'poison echo must carry blocked_by pointing at its actual root'
        assert echo['blocked_symbol_declaration'].endswith('probe.a:3:11'), 'same declaration symbol'
    secondary = next(f for f in findings if f['kind']=='NotYet' and f['reason']=='reading secondary')
    assert secondary['blocked_by']==signature, 'transitive echo must flatten to the first root'
    assert secondary['blocked_symbol_declaration'].endswith('probe.a:4:11'), 'secondary declaration identity'
    assert not any(f['reason']=='reading poison' and int(f['where'].rsplit(':',2)[1]) in (8,10) for f in findings), 'shadowed and independent parameter symbols remain usable'
    return root,echoes,secondary

def main(binary, output):
    binary=Path(binary).resolve();output=Path(output).resolve();output.mkdir(parents=True,exist_ok=True)
    source=output/'probe.a';source.write_text(SOURCE)
    records={}
    for name,extra in [('positive',{}),('drop-tag-mutant',{'LATENT_MUTANT_DROP_BLOCKED_BY':'1'})]:
        destination=output/(name+'.jsonl')
        with (output/(name+'.log.txt')).open('w') as log:
            subprocess.run([str(binary),str(source),str(destination)],env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',**extra),stdout=log,stderr=subprocess.STDOUT,check=True)
        records[name]=[json.loads(line) for line in destination.read_text().splitlines()]
    root,echoes,secondary=check(records['positive'])
    try:check(records['drop-tag-mutant'])
    except AssertionError as failure:
        assert str(failure)=='poison echo must carry blocked_by pointing at its actual root', str(failure)
        print('drop-tagging mutant caught by poison blocked_by assertion:',failure)
    else:raise AssertionError('drop-tagging mutant survived')
    print('PASS: three poison reads including shorthand carry the actual failed initializer root; secondary flattens to that root; shadow and independent parameter symbols are not poisoned; no-output guards')
    print('root:',json.dumps(root,sort_keys=True))
    print('echo:',json.dumps(echoes[0],sort_keys=True))

if __name__=='__main__':main(*sys.argv[1:])
