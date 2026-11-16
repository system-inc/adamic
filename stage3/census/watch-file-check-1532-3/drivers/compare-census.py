import json,pathlib,re,time,collections
s=pathlib.Path('/workspace/scratch/watch1532-recut');result=s/'logs/candidate-results.json'
while not result.exists():time.sleep(2)
rows=json.loads(result.read_text());assert rows[0]['name']=='candidate-census' and rows[0]['exit']==0,rows
streams={k:[json.loads(x) for x in (s/'logs'/(k+'-census.jsonl')).open()] for k in ['baseline','candidate']}
def findings(k):
 r=streams[k];assert r[0]['latent_mode']=='full' and r[0]['module_format']=='CommonJSBundle';assert not r[0]['checker_rejected'] and not r[0]['diagnostics'];assert r[-1]['status']=='relation_report' and not r[-1]['findings'];files=[x for x in r if x.get('file')];assert len(files)==1 and files[0]['file'].endswith('/src/compiler/sys.ts');return files[0]
f={k:findings(k) for k in streams}
def key(k,x):
 text=x['text'].replace(str(s/('baseline' if k=='baseline' else 'fresh'))+'/','')
 text=re.sub(r':\d+:\d+','',text)
 return x['kind'],x['phase'],x['reason'],text
c={k:collections.Counter(key(k,x) for x in f[k]['findings']) for k in f};gone=c['baseline']-c['candidate'];new=c['candidate']-c['baseline']
serialize=lambda counter:[dict(kind=k[0],phase=k[1],reason=k[2],messageWithoutPosition=k[3],count=n) for k,n in sorted(counter.items())]
report=dict(pairing=['kind','phase','reason','message with position stripped'],baseline=dict(raw=len(f['baseline']['findings']),unique=len(c['baseline']),units=len(f['baseline']['units'])),candidate=dict(raw=len(f['candidate']['findings']),unique=len(c['candidate']),units=len(f['candidate']['units'])),gone=serialize(gone),new=serialize(new),changed=0 if not new else None,baselineTargetRecords=[x for x in f['baseline']['findings'] if 'a function taking PollingInterval seen' in x['reason']],headerChecks=True,relationTrailers=True)
(s/'census-comparison.json').write_text(json.dumps(report,indent=2)+'\n');print(json.dumps(report,indent=2),flush=True)
assert not new,'new or changed census findings'
assert len(gone)==1 and sum(gone.values())==1
assert next(iter(gone))[0]=='Refused' and 'a function taking PollingInterval seen' in next(iter(gone))[2]
