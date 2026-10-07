"""Compare owned numeric profiles with unmodified production Go rules."""
import argparse,concurrent.futures,json,os,pathlib,subprocess,time
p=argparse.ArgumentParser();p.add_argument('--artifacts',required=True);p.add_argument('--fixtures',required=True);p.add_argument('--native',required=True);p.add_argument('--compiler-root',required=True);a=p.parse_args()
s=pathlib.Path(__file__).resolve().parent;r=s.parents[3];out=pathlib.Path(a.artifacts).resolve();out.mkdir(parents=True,exist_ok=True)
def run(name,cmd,cwd=r):
 with (out/(name+'.stdout')).open('wb') as stdout,(out/(name+'.stderr')).open('wb') as stderr:
  begin=time.perf_counter();proc=subprocess.run([str(x) for x in cmd],cwd=cwd,stdout=stdout,stderr=stderr);elapsed=time.perf_counter()-begin
 return proc.returncode,(out/(name+'.stdout')).read_bytes(),(out/(name+'.stderr')).read_bytes(),elapsed
virtual=r/'cohere/adamic_wave08_core_oracle.go';overlay=out/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(s/'testdata/oracle.go')}}))
status,_,errors,_=run('oracle-build',['go','build','-overlay',overlay,'-o',out/'oracle',virtual],r/'cohere');assert status==0,errors
cases=[]
for folder in sorted((pathlib.Path(a.fixtures)/'cases').iterdir()):
 metadata=json.loads((folder/'metadata.json').read_text());name=metadata['Rule'];flag={'require-await':'--await','require-atomic-updates':'--atomic','symbol-description':'--symbol'}[name]
 if name=='symbol-description' and not metadata['Typed']:continue
 options=metadata['Options'];args=[flag]
 if isinstance(options,dict) and options.get('allowProperties',False):args+=['--allow-properties']
 cases.append((folder.name,folder/'tsconfig.json',folder/'roots.manifest',args,metadata['Name']))
for corpus,prefix in [('compiler',a.compiler_root),('repository',str(r))]:
 manifest=out/(corpus+'.manifest');manifest.write_text(''.join(str(pathlib.Path(prefix)/path)+'\n' for path in (s.parent/'validation-coverage'/(corpus+'.manifest')).read_text().splitlines()))
 config=pathlib.Path(a.compiler_root)/'src/compiler/tsconfig.json' if corpus=='compiler' else r/'tsconfig.json'
 for flag in ['--atomic','--await','--symbol']:cases.append((corpus+flag,config,manifest,[flag],corpus))
def compare(case):
 name,config,manifest,args,description=case
 gs,go,ge,gt=run(name+'-go',[out/'oracle',config,manifest,*args]);ns,native,ne,nt=run(name+'-native',[a.native,config,manifest,*args])
 same=gs==ns==0 and go==native and ne==b''
 result=dict(name=name,description=description,flags=args,go_status=gs,native_status=ns,match=same,go_seconds=gt,native_seconds=nt,go_findings=go.splitlines()[-1].decode(errors='replace') if go else '',native_findings=native.splitlines()[-1].decode(errors='replace') if native else '',native_error=ne.decode(errors='replace'))
 if not same:
  result['first_difference']=next((i for i,(g,n) in enumerate(zip(go,native)) if g!=n),min(len(go),len(native)))
 return result
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:results=list(pool.map(compare,cases))
(out/'results.json').write_text(json.dumps(results,indent=2)+'\n')
for result in results:
 if not result['match']:print(json.dumps(result),flush=True)
print('matches',sum(x['match'] for x in results),'of',len(results),flush=True)
assert all(x['match'] for x in results),'production byte differences recorded'
