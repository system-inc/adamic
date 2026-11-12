import collections,json,os,pathlib,signal,subprocess,sys
S=pathlib.Path('/workspace/ldpi')
def run(kind,tag,binary,files=None):
 p=S/'fixture-control'/f'{kind}-{tag}';p.mkdir(parents=True,exist_ok=True)
 cmd=[sys.executable,str(S/'run.py'),str(p),'env','LATENT_FULL=0','LATENT_SPECULATIVE=0','LATENT_ASSERT_NO_OUTPUT=1',f'LDPI_INSTANTIATIONS={p}/instantiations.json']
 if files is not None:cmd.append('LATENT_ONLY_FILE='+'\n'.join(files))
 cmd += [str(binary),str(S/'split-fixture/main.ts'),str(p/'census.jsonl')]
 code=subprocess.call(cmd,env=dict(os.environ,LDPI_DEADLINE_SECONDS='40'))
 if code:
  if code<0:signal.signal(-code,signal.SIG_DFL);os.kill(os.getpid(),-code)
  sys.exit(code)
 rows=[json.loads(l) for l in (p/'census.jsonl').read_text().splitlines()]
 assert rows[0]['status']=='measurement' and rows[-1]['status']=='relation_report'
 return rows,p

def witness_key(v):
 # Match json.Marshal(LatentInstantiation)'s struct field order and HTML escaping.
 text=json.dumps({k:v[k] for k in ['unit','call_site','type_arguments']},ensure_ascii=False,separators=(',',':'))
 return text.replace('<',r'\u003c').replace('>',r'\u003e').replace('&',r'\u0026').replace('\u2028',r'\u2028').replace('\u2029',r'\u2029')

def merge(rows,states):
 files={}
 for data in rows:
  for row in data[1:-1]:
   assert row['file'] not in files
   files[row['file']]=row
 entered={};witnesses=collections.defaultdict(dict)
 for st in states:
  entered.update({k:v for k,v in st['entered'].items() if v})
  for k,values in (st['witnesses'] or {}).items():
   for v in values:witnesses[k][witness_key(v)]=v
 for row in files.values():
  for u in row['units']:
   if u['status']=='deferred_generic_instantiation':
    u['instantiated']=entered.get(u['where'],False)
    values=[v for k,v in sorted(witnesses[u['where']].items())][:5]
    if values:u['instantiated_by']=values
    else:u.pop('instantiated_by',None)
 return files
if __name__=='__main__':
 result={}
 for kind in ['shallow','fixed']:
  whole,p=run(kind,'whole',S/f'{kind}-entry-census')
  data=[];states=[]
  for i,path in enumerate(whole[0]['sources']):
   rows,p=run(kind,str(i),S/f'{kind}-split-census',[path]);data.append(rows)
   states.append(json.loads((p/'instantiations.json').read_text()))
  combined=merge(data,states)
  expected={r['file']:r for r in whole[1:-1]}
  assert combined==expected,kind
  instantiated=[u for r in expected.values() for u in r['units'] if u['status']=='deferred_generic_instantiation' and u['instantiated']]
  assert instantiated,'fixture must exercise cross-file instantiation union'
  assert any(len(u.get('instantiated_by',[]))==5 for u in instantiated),'fixture must exercise the five-witness cap'
  result[kind]={'whole_vs_slices_file_records_equal':True,'files':len(expected),'instantiated_generic_units':len(instantiated)}
 (S/'fixture-control/summary.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
