#!/usr/bin/env python3
"""Independently reconcile published rows against scratch census streams."""
import collections, gzip, hashlib, json, pathlib, subprocess, sys
out=pathlib.Path(__file__).resolve().parent
scratch=pathlib.Path(sys.argv[1]).resolve()
adapted=scratch/'adapted'
manifest=json.loads((out/'manifest.json').read_text())
states=json.loads((out/'pieces.json').read_text())
assert len({s['file'] for s in states})==len(states)
assert set(s['file'] for s in states)<=set(manifest['files'])
expected=collections.Counter(); owners={}; meter=set(); texts=collections.defaultdict(set); matched=0
for state in states:
    file=state['file']; raw=scratch/'raw'/(file.replace('/','__')+'.jsonl')
    assert hashlib.sha256(raw.read_bytes()).hexdigest()==state['raw_sha256'],file
    records=[]
    for line in raw.open():
        try: records.append(json.loads(line))
        except json.JSONDecodeError: assert not state['complete'],file
    if state['complete']:
        assert state['exit_code']==0,file
        assert records[0]['latent_mode']=='full' and records[0]['module_format']=='CommonJSBundle',file
        assert records[-1]['status']=='relation_report',file
        files=[r for r in records if r.get('file')]
        assert len(files)==1 and files[0]['file']==str(adapted/file),file
    units={u['where']:u for record in records for u in record.get('units',[])}
    for record in records:
        for f in record.get('findings',[]):
            if f.get('kind')!='Refused' or ', whose readonly field pos becomes writable:' not in f.get('reason',''): continue
            where=f['where'].replace(str(adapted)+'/', '')
            fn,line,col=where.rsplit(':',2)
            unit=f['unit'].replace(str(adapted)+'/', '')
            key=(fn,int(line),int(col),f['reason'],unit)
            expected[key]+=1; matched+=1; texts[key].add(f['text'].replace(str(adapted)+'/', ''))
            u=units[f['unit']]
            owner=(u['kind'],u.get('name',''))
            assert key not in owners or owners[key]==owner,key
            owners[key]=owner
            meter.add((f['kind'],where,f['reason'],f['text'].replace(str(adapted)+'/', '')))
rows=[json.loads(line) for line in gzip.open(out/'rows.jsonl.gz','rt')]
keys=[(r['file'],r['line'],r['col'],r['reason'],r['unit']) for r in rows]
assert len(keys)==len(set(keys)) and keys==sorted(keys)
assert set(keys)==set(expected)
for key,r in zip(keys,rows):
    assert r['multiplicity']==expected[key],key
    assert (r['unit_kind'],r['unit_name'])==owners[key],key
    assert r['where']==f"{r['file']}:{r['line']}:{r['col']}",key
assert sum(r['multiplicity'] for r in rows)==matched
assert all(len(values)==1 for values in texts.values()), 'text variation within a requested key'
assert all(r['text'] in texts[key] for key,r in zip(keys,rows))
for file,sha in manifest['source_sha256'].items():
    assert hashlib.sha256((adapted/file).read_bytes()).hexdigest()==sha,file
root=out.parents[2]
changed=subprocess.check_output(['git','diff','--name-only',manifest['sha'],'HEAD'],cwd=root,text=True).splitlines()
assert all(p.startswith('stage3/census/pos-writable-rows/') for p in changed),changed
assert not subprocess.check_output(['git','diff',manifest['sha'],'--','stage3/census/latent/full_audit.py'],cwd=root)
result=dict(status='pass',complete_pieces=sum(s['complete'] for s in states),failed_pieces=[s['file'] for s in states if not s['complete']],unattempted=sorted(set(manifest['files'])-set(s['file'] for s in states)),distinct_rows=len(rows),matching_records=matched,meter_style_distinct=len(meter),source_hashes_checked=len(manifest['source_sha256']),unit_metadata='matched raw unit ledger',full_audit_py='unchanged',compiler_and_adaptation_diff='none')
(out/'verification.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result))
