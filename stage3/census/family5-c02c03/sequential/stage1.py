import os,sys,time,json,re,subprocess,shutil,hashlib
from pathlib import Path
R=Path('/tmp/family5-land');S=Path('/tmp/family5-sequential');BASE='fbe5b72ba115a8f3060d4b5229fd578574f4505f';C02='28a2e22015d0cf51667fcefa7acb246cab553817'
os.chdir(R)
os.environ.update(ADAMIC_TYPESCRIPT_SOURCE='/workspace/mapset-upstream',ADAMIC_BUILD_CACHE_DIR='/workspace/family5-build-cache')
def run(name,cmd,env=None):
 code=subprocess.call([sys.executable,'/tmp/family5-proof.py',str(S/name),*map(str,cmd)],env=env)
 return code
status=S/'c02/lane-alone/status.json'
while not status.exists():time.sleep(5)
lane=json.loads(status.read_text());report=json.loads(Path('/workspace/family5-c02-lane-alone/report.json').read_text());oracle=json.loads(Path('/workspace/family5-c02-lane-alone/oracle/report.json').read_text())
if lane['exit_code']==0:
 assert report['counts']=={'passing':106366,'failing':1,'pending':0}
 print('PASS isolated standard-eight-worker lane child_exit=0; 106366 passing, 1 sanctioned failure',flush=True)
 code=run('c02/product-alone',['go','test','./stage3/lane/adaptedtree','-run','^TestProduct_Stage3AdaptedTree$','-count=1','-timeout','45m','-v'])
 (S/'stage1-result.json').write_text(json.dumps({'lane_green':True,'product_exit':code,'next':'paired full censuses' if code==0 else 'stop at product failure'},indent=2)+'\n')
 sys.exit(code)
# Measure both original cases and every timeout from the isolated lane; no parallel execution.
cases={'codeFixClassImplementInterfaceNoTruncation.ts':'fourslash','relationComplexityError.ts':'compiler'}
for f in oracle.get('failures',[]):
 if f.get('cause')!='timeout':continue
 title=f['title'];paths=re.findall(r'tests/cases/(?:compiler|conformance|fourslash)/[^"\s]+?\.ts',title)
 for path in paths:
  runner='conformance' if '/conformance/' in path else 'fourslash' if '/fourslash/' in path else 'compiler'
  cases[Path(path).name]=runner
measurements=[]
for side,sha,tree in [('c02',C02,Path('/workspace/family5-c02-lane-alone/adapted-tree')),('main',BASE,Path('/tmp/family5-base-tree'))]:
 subprocess.run(['git','switch','--detach',sha],check=True)
 assert subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip()==sha
 for filename,runner in cases.items():
  name='targeted/'+side+'/'+filename.removesuffix('.ts');out=S/name;timing=out/'runnable-times';timing.mkdir(parents=True)
  local=tree/'tests/baselines/local'
  if local.exists():shutil.rmtree(local)
  env=dict(os.environ,FAMILY5_TIMING_DIR=str(timing));env['NODE_OPTIONS']=env.get('NODE_OPTIONS','')+' --require='+str(S/'measure-runnables.cjs')
  cmd=['bash','stage3/oracle/run.sh',tree,out/'oracle','--runners',runner,'--tests',re.escape(filename)]
  code=run(name,cmd,env)
  data=json.loads((out/'oracle/report.json').read_text());events=[]
  for file in timing.glob('*.jsonl'):events.extend(json.loads(l) for l in file.read_text().splitlines())
  matching=[x for x in events if filename in x['title']]
  timed=[x for x in matching if (x['kind']=='test' if runner=='fourslash' else x['kind']=='hook' and 'before all' in x['title'])]
  row={'side':side,'sha':sha,'case':filename,'runner':runner,'oracle_exit':code,'test_command_exit':data.get('phases',{}).get('tests',{}).get('exit'),'test_phase_seconds':data.get('phases',{}).get('tests',{}).get('seconds'),'counts':data.get('counts'),'case_elapsed_ms':max((x['elapsed_ms'] for x in timed),default=None),'runnable_measurement':'test body' if runner=='fourslash' else 'before-all compilation hook','timing_events':matching,'failures':[{k:v for k,v in f.items() if k in ['title','cause','message','elapsed_ms','timeout_ms']} for f in data.get('failures',[])]}
  measurements.append(row);(out/'measurement.json').write_text(json.dumps(row,indent=2)+'\n')
  print(json.dumps({k:v for k,v in row.items() if k!='timing_events' and k!='failures'}),flush=True)
 # Hash all emitted JavaScript and the public declaration, after the same locked build.
 built=tree/'built/local';manifest={str(p.relative_to(built)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(built.rglob('*.js')) if p.is_file()}
 manifest['typescript.d.ts']=hashlib.sha256((built/'typescript.d.ts').read_bytes()).hexdigest()
 (S/('built-'+side+'.json')).write_text(json.dumps(manifest,indent=2)+'\n')
subprocess.run(['git','switch','typescript/line-and-character-1'],check=True)
summary=[]
for filename in cases:
 a=next(x for x in measurements if x['case']==filename and x['side']=='c02');b=next(x for x in measurements if x['case']==filename and x['side']=='main')
 sa=any(f.get('cause')=='timeout' for f in a['failures']);sb=any(f.get('cause')=='timeout' for f in b['failures'])
 grade='timeout on both sides' if sa and sb else 'C02-only timeout' if sa else 'main-only timeout' if sb else 'no isolated timeout on either side'
 summary.append({'case':filename,'c02_ms':a['case_elapsed_ms'],'main_ms':b['case_elapsed_ms'],'c02_exit':a['oracle_exit'],'main_exit':b['oracle_exit'],'grade':grade})
a=json.loads((S/'built-c02.json').read_text());b=json.loads((S/'built-main.json').read_text());diff={k:{'c02':a.get(k),'main':b.get(k)} for k in sorted(a.keys()|b.keys()) if a.get(k)!=b.get(k)}
result={'lane_green':False,'lane_counts':report['counts'],'timings':summary,'built_js_and_public_dts_equal':not diff,'built_difference':diff,'next':'stop: standard lane remains red; no landing push'}
(S/'stage1-result.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2),flush=True)
sys.exit(1)
