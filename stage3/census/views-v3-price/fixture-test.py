#!/usr/bin/env python3
"""Prove read-form enumeration; the bound deliberately does not reproduce admission."""
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
declare const union: readonly Node[] | Node[];
const g=union[0];
declare const tuple: readonly [Node];
const h=tuple[0];
declare const mutableTuple: [Node];
const i=mutableTuple[0];
declare const intersection: readonly Node[] & {readonly brand: 1};
const j=intersection[0];
'''
expected = {k:dict(upper_bound=6 if k=='index' else 1,readonly_subset=4 if k=='index' else 1) for k in ['at','for-of','forEach','index','map','slice','spread']}
def verify(data):
    assert data['upper_bound'] == 12, data
    assert data['readonly_subset'] == 10, data
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
