"""Admission checks: unchanged mode-off output, file/unit union, witness scope, and lost/moved data controls."""
from copy import deepcopy
import gzip
import json
import os
from pathlib import Path
import subprocess
from pieces_common import OUT,E,S,load,norm,dump,digest,validate_piece,verify_partition

root=S/'selector-audit'; root.mkdir(exist_ok=True)
sources={'a.a':'import { identity } from "./b.a";\nexport function caller(): number { return identity(1); }\nexport function pieceProbe(): number { return 1; }\n',
         'b.a':'export function identity<T>(value: T): T { return value; }\nfunction another(): number { return 1; }\nconst sentinel: number = 1;\n'}
for file,text in sources.items():(root/file).write_text(text)
env=dict(os.environ,LATENT_FULL='0',LATENT_SPECULATIVE='0',LATENT_ASSERT_NO_OUTPUT='1',LATENT_MUTANT_FUNCTION='pieceProbe',TMPDIR=str(S/'tmp'))
env.pop('ANY_MEMBERS_CENSUS_PIECE',None)

def invoke(binary,name,selection=None):
    local=dict(env,LATENT_TIMES=str(root/(name+'.times')))
    if selection:
        path=root/(name+'.plan.json');dump(path,selection)
        local.update(ANY_MEMBERS_CENSUS_ROOT=str(root),ANY_MEMBERS_CENSUS_PIECE=str(path))
    with (root/(name+'.log')).open('w') as log:
        subprocess.run([str(binary),str(root),str(root/(name+'.jsonl'))],env=local,stdout=log,stderr=subprocess.STDOUT,timeout=30,check=True)
    return load(root/(name+'.jsonl')),load(root/(name+'.times'))

for label in ('base','head'):
    original,_=invoke(S/f'{label}-bin/latent-census',label+'-original')
    selected,timers=invoke(S/f'piece-{label}-bin/latent-census',label+'-off')
    assert original==selected,'mode-off output changed'
    if label=='base':whole,whole_times=norm(selected,root),norm(timers,root)
records={r['file']:r for r in whole[1:]}
assert set(records)==set(sources)
all_units={u['where'] for r in records.values() for u in r['units']}
timed={t['where'] for t in whole_times}
a=[u['where'] for u in records['a.a']['units']]
b=[u['where'] for u in records['b.a']['units']]
pieces=[]
for i,(file,units) in enumerate([('a.a',a),('b.a',b[:2]),('b.a',b[2:])]):
    piece=dict(schema=1,id=str(i),files=[dict(file=file,all_units=False,units=units)],expected_units=units,expected_timed_units=[u for u in units if u in timed])
    pieces.append(piece)
verify_partition(pieces,all_units)
outputs=[]
for piece in pieces:
    rows,times=invoke(S/'piece-base-bin/latent-census','piece-'+piece['id'],piece)
    outputs.append(validate_piece(piece,rows,times,whole[0],root))
first=lambda rows:{u:next((f for r in rows[1:] for f in r['findings'] if f['phase']=='lowering' and f['unit']==u),None) for u in all_units}
joined=[whole[0]]+[r for rows,_ in outputs for r in rows[1:]]
assert first(joined)==first(whole),'fixture first stop union changed'
whole_units={u['where']:u for r in whole[1:] for u in r['units']}
piece_units={u['where']:u for r in joined[1:] for u in r['units']}
witness=[]
for u in all_units:
    before,after=whole_units[u],piece_units[u]
    if before!=after:
        assert {k:v for k,v in before.items() if k!='instantiated'}=={k:v for k,v in after.items() if k!='instantiated'}
        witness.append(dict(unit=u,name=before.get('name'),whole=before.get('instantiated'),piece=after.get('instantiated')))
assert witness==[dict(unit='b.a:1:1',name='identity',whole=True,piece=False)],witness
for mutant in (pieces[:-1],pieces+[pieces[0]]):
    try:verify_partition(mutant,all_units)
    except AssertionError:pass
    else:raise AssertionError('dropped/duplicate piece accepted')
# Preserve the finding's site and owner while moving it to another scheduled piece.
mutant=deepcopy(outputs)
source=next((rows,r,f) for rows,_ in mutant for r in rows[1:] for f in r['findings'] if f['phase']=='lowering')
rows,row,finding=source;row['findings'].remove(finding)
other=next(i for i,p in enumerate(pieces) if finding['unit'] not in p['expected_units'])
mutant[other][0][1]['findings'].append(finding)
try:validate_piece(pieces[other],mutant[other][0],mutant[other][1],whole[0],root)
except AssertionError:pass
else:raise AssertionError('moved-unit finding accepted')
for path in root.glob('*.jsonl'):
    with gzip.open(E/('audit-'+path.name+'.gz'),'wb') as f:f.write(path.read_bytes())
dump(E/'selector-audit.json',dict(status='passed',sources=sources,mode_off_identical_both_compilers=True,disjoint_file_and_unit_selectors=True,first_stops_match=True,actual_witness_scope_difference=witness,controls_rejected=['dropped piece','duplicate piece','moved finding with unchanged global site identity']))
print('PASS: both compilers preserve mode-off output; file/unit partition first stops match; cross-file generic witness difference retained; drop, duplicate and moved-finding mutants rejected.')
