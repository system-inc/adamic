#!/usr/bin/env python3
import json,sys
from pathlib import Path
base=Path(__file__).resolve().parent
sites=json.loads((base/'evidence/sites.json').read_text())
result={}
for label,directory in zip(sys.argv[1::2],sys.argv[2::2]):
    directory=Path(directory)
    total={r['site']:dict(calls=0,completed=0,identity=0,frozen_arrays=0,undefined_identity=0) for r in sites}
    processes=[]
    throws=[]
    observer_errors=[]
    for file in sorted(directory.glob('*.counts.json')):
        row=json.loads(file.read_text());processes.append({k:v for k,v in row.items() if k not in ('counts','scripts')})
        for site,values in row['counts'].items():
            values=dict(values,completed=values.get('completed',values['calls']))
            if site not in total: total[site]={k:0 for k in values}
            for k,v in values.items():total[site][k]+=v
    for file in sorted(directory.glob('*.throws.jsonl')):
        for line in file.read_text().splitlines():
            row=json.loads(line)
            process=json.loads((directory/(str(row['pid'])+'.counts.json')).read_text())
            for frame in row['frames']:frame['url']=frame['url'] or process['scripts'].get(frame['scriptId'],'')
            if (row['description'].startswith('TypeError: Cannot add initializers after decoration has completed\n')
                and row['frames'][0]['function'] == 'context.addInitializer'
                and 'esDecorators' in (row.get('case') or {}).get('title','')
                and not row['frozen_aliases']):
                row['classification']='unrelated-explicit-decorator-error'
                row['classification_evidence']='src/compiler/factory/emitHelpers.ts:783 explicitly throws this TypeError when done, before extraInitializers.push; six esDecorators cannot-call-after-decoration tests assert it.'
            elif row['frozen_aliases']:
                row['classification']='freeze-write'
            else:
                row['classification']='unclassified'
            throws.append(row)
    for file in directory.glob('*.observer-errors.jsonl'):
        observer_errors.extend(json.loads(line) for line in file.read_text().splitlines())
    for values in total.values():
        assert values['calls'] >= values['completed'] >= values['identity']
        assert values['identity'] == values['frozen_arrays'] + values['undefined_identity']
    assert sum(p['freezeThrows'] for p in processes) == len(throws), 'missing throw records'
    assert sum(p['freezeEvents'] for p in processes) == sum(v['frozen_arrays'] for v in total.values())
    entry_available=all(p.get('counter_mode')=='entry-and-completion' for p in processes)
    result[label]={'entry_count_available':entry_available,'call_count_kind':'entries' if entry_available else 'completed returns (legacy counter)','counts':total,'totals':{k:sum(v[k] for v in total.values()) for k in ['calls','completed','identity','frozen_arrays','undefined_identity']},'processes':processes,'throws':throws,'freeze_write_throws':[r for r in throws if r['classification']=='freeze-write'],'unrelated_throws':[r for r in throws if r['classification'].startswith('unrelated-')],'unclassified_throws':[r for r in throws if r['classification']=='unclassified'],'observer_errors':observer_errors}
(base/'evidence/counts.json').write_text(json.dumps(result,indent=2)+'\n')
for k,r in result.items():print(k,r['totals'],'candidates',len(r['throws']),'freeze writes',len(r['freeze_write_throws']),'unclassified',len(r['unclassified_throws']),'observer errors',len(r['observer_errors']))
