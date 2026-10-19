import json, pathlib, subprocess
p=pathlib.Path(__file__).resolve().parent
root=p.parents[2]
base='dbdf977f5b8b0979d414ebe6f20e7aadd312e309'
source='20fde75fa96b48f3dea2e3612b418acae666332f'
def rows(text):
    result={}
    for line in text.splitlines():
        fields=[field.strip() for field in line.split('|')]
        if len(fields)==9 and fields[1]!='Fixture':
            try: result[fields[1]]=[int(x) for x in fields[2:8]]
            except ValueError: pass
    return result
def git(*args): return subprocess.check_output(['git',*args],cwd=root,text=True)
old=rows(git('show',base+':internal/oracle/counts.md'))
new=rows((root/'internal/oracle/counts.md').read_text())
changes=[dict(fixture=k,before=old.get(k),after=new.get(k)) for k in sorted(old.keys()|new.keys()) if old.get(k)!=new.get(k)]
(p/'counts-row-changes.json').write_text(json.dumps(changes,indent=2)+'\n')
source_old=rows(git('show',source+'~1:internal/oracle/counts.md'))
source_new=rows(git('show',source+':internal/oracle/counts.md'))
source_changes={k:(source_old.get(k),source_new.get(k)) for k in source_old.keys()|source_new.keys() if source_old.get(k)!=source_new.get(k)}
comparison={'base_rows':len(old),'rebased_rows':len(new),'added':sum(x['before'] is None for x in changes),'removed':sum(x['after'] is None for x in changes),'changed':sum(x['before'] is not None and x['after'] is not None for x in changes),'equals_source_numeric_deltas':{x['fixture']:(x['before'],x['after']) for x in changes}==source_changes}
comparison['additional_changed_rows_on_this_base']=[x['fixture'] for x in changes if x['fixture'] not in source_changes]
comparison['all_source_changed_rows_match']=all(any(x['fixture']==k and (x['before'],x['after'])==v for x in changes) for k,v in source_changes.items())
(p/'counts-summary.json').write_text(json.dumps(comparison,indent=2)+'\n')
all_events=[]; summaries={}
for log in sorted(p.glob('*.log')):
    events=[]
    for line in log.read_text(errors='replace').splitlines():
        try: value=json.loads(line)
        except ValueError: continue
        if not isinstance(value,dict) or value.get('Action') not in ('pass','fail','skip') or not value.get('Test'): continue
        events.append(dict(log=log.name,package=value['Package'],test=value['Test'],result=value['Action'],seconds=value.get('Elapsed',0)))
    if events:
        summaries[log.name]={'events':len(events),'pass':sum(x['result']=='pass' for x in events),'fail':sum(x['result']=='fail' for x in events),'skip':sum(x['result']=='skip' for x in events),'max_seconds':max(x['seconds'] for x in events)}
        all_events+=events
(p/'test-seconds.json').write_text(json.dumps(all_events,indent=2)+'\n')
(p/'test-summary.json').write_text(json.dumps(summaries,indent=2)+'\n')
print(json.dumps(comparison))
for key,value in summaries.items(): print(key,json.dumps(value))
