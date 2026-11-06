#!/usr/bin/env python3
"""Prove read-form enumeration; this does not certify admission equivalence."""
import json, os, pathlib, subprocess, tempfile
counter = pathlib.Path(__file__).with_name('count.cjs').resolve()
source = '''interface Node { readonly kind: number; }
declare const a: readonly Node[];
declare const mutable: Node[];
const one=a[0];
a[1] = one!;
const b=a.at(0);
for(const n of a) console.log(n);
const c=[...a];
const d=a.slice();
const e=a.map(x=>x);
a.forEach(x=>console.log(x));
const f=mutable[0];
'''
expected = dict(at=1, **{'for-of':1}, forEach=1, index=1, map=1, slice=1, spread=1)
def verify(data):
    assert data['candidates'] == 7, data
    assert data['unresolved'] == 0, data
    assert data['forms'] == expected, data
with tempfile.TemporaryDirectory(prefix='v3-price-fixture-') as scratch:
    scratch = pathlib.Path(scratch)
    (scratch/'fixture.ts').write_text(source)
    for mutant in (False,True):
        env=dict(os.environ)
        env.pop('V3_MUTANT_SKIP_INDEX',None)
        if mutant: env['V3_MUTANT_SKIP_INDEX']='1'
        prefix=scratch/('mutant' if mutant else 'normal')
        with (scratch/'run.log').open('w') as log:
            subprocess.run(['node',str(counter),str(scratch),str(prefix),'fixture'],env=env,stdout=log,stderr=log,check=True)
        data=json.loads(prefix.with_suffix('.json').read_text())
        if not mutant: verify(data)
        else:
            try: verify(data)
            except AssertionError: pass
            else: raise AssertionError('index-skipping mutant survived')
        print(json.dumps({'mutant':mutant,'summary':data,'assertion':'rejected' if mutant else 'passed'}))
