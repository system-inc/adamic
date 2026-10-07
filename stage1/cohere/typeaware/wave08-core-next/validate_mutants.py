"""Compiling production-rule mutants must fail only the byte oracle."""
import argparse,json,pathlib,subprocess
p=argparse.ArgumentParser();p.add_argument('--artifacts',required=True);p.add_argument('--baseline',required=True);p.add_argument('--stage0',required=True);p.add_argument('--archive',required=True);p.add_argument('--fixtures',required=True);a=p.parse_args()
s=pathlib.Path(__file__).resolve().parent;r=s.parents[3];out=pathlib.Path(a.artifacts).resolve();out.mkdir(parents=True,exist_ok=True);baseline=pathlib.Path(a.baseline)
def run(name,cmd,expected=0):
 with (out/(name+'.stdout')).open('wb') as stdout,(out/(name+'.stderr')).open('wb') as stderr:proc=subprocess.run([str(x) for x in cmd],cwd=r,stdout=stdout,stderr=stderr)
 data=(out/(name+'.stdout')).read_bytes();errors=(out/(name+'.stderr')).read_bytes();assert proc.returncode==expected,(name,proc.returncode,errors);return data,errors
records=[];cases=json.loads((baseline/'results.json').read_text())
mutants=[('atomic','require_atomic_updates.a','state.outdated.has(event.symbol)','!state.outdated.has(event.symbol)'),('await','require_await.a',"? ';' : ''","? '' : ';'"),('symbol','symbol_description.a',"callee.text !== 'Symbol'","callee.text === 'Symbol'")]
for slug,filename,before,after in mutants:
 folder=out/slug;folder.mkdir(exist_ok=True)
 for file in s.glob('*.a'):
  text=file.read_text().replace("'../","'"+str(s.parent)+"/")
  if file.name==filename:assert text.count(before)==1;text=text.replace(before,after)
  (folder/file.name).write_text(text)
 run(slug+'-build',[a.stage0,'build',folder/'suite.a','-o',out/slug/'native','--tsgo',a.archive,'--sanitize'])
 caught=[]
 for case in cases:
  if case['description'] in ['compiler','repository'] or case['flags'][0]!='--'+slug:continue
  fixture=pathlib.Path(a.fixtures)/'cases'/case['name']
  data,errors=run(slug+'-'+case['name'],[out/slug/'native',case.get('config',str(fixture/'tsconfig.json')),case.get('manifest',str(fixture/'roots.manifest')),*case['flags']]);assert errors==b''
  truth=(baseline/(case['name']+'-go.stdout')).read_bytes()
  if data!=truth:
   first=next((i for i,(g,n) in enumerate(zip(truth,data)) if g!=n),min(len(truth),len(data)))
   caught.append(dict(case=case['name'],first_difference=first))
 assert caught,slug
 records.append(dict(mutant=slug,compiles=True,exits=0,stderr='',caught=caught))
 print(slug,'compiled, all runs exit zero with empty stderr; byte comparison catches',len(caught),'cases, first',caught[0]['first_difference'],flush=True)
(out/'results.json').write_text(json.dumps(records,indent=2)+'\n')
run('handle-build',[a.stage0,'build',s/'testdata/handle_probe.a','-o',out/'handle','--tsgo',a.archive,'--sanitize'])
fixture=pathlib.Path(a.fixtures)/'cases'/cases[0]['name'];path=(fixture/'roots.manifest').read_text().splitlines()[0]
for question in ['wave08-symbol-origins','wave08-syntax-snapshot','wave08-await-contract','wave08-await-type\n1\n0\n']:
 slug=question.split('\n')[0];data,errors=run('released-'+slug,[out/'handle',fixture/'tsconfig.json',path,question,'--release'],70)
 assert data==b'' and b'invalid or released checker handle' in errors
 print(slug,'released handle refused before output, exit 70',flush=True)
print('PASS',flush=True)
