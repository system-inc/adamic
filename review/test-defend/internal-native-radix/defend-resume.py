import pathlib,json,os,subprocess,time,difflib,re
root=pathlib.Path('/workspace/adamic');os.chdir(root);p=root/'review/test-defend/internal-native-radix';env=dict(os.environ,ADAMIC_RECORD_BENCH='1');base=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip();records=[]
names=[n for n in (p/'list.log').read_text().splitlines() if n.startswith('Test') and any(x in n for x in ['RegExp','Regex','String','Release','Record','Radix','Map'])]
pattern='^('+'|'.join(names)+')$';(p/'matrix-scope.json').write_text(json.dumps({'base':base,'tests':names,'pattern':pattern,'limits':'Direct runtime consumers selected by current test names and source callers. Other emitted-program and decoder suites remain unknown.'},indent=2))
def run(id,cmd,extra={}):
 t=time.monotonic()
 with (p/(id+'.log')).open('w') as f:r=subprocess.run(cmd,shell=True,stdout=f,stderr=subprocess.STDOUT,env=dict(env,**extra))
 es=[]
 for l in (p/(id+'.log')).read_text().splitlines():
  try:e=json.loads(l)
  except:continue
  if isinstance(e,dict):es.append(e)
 out=dict(id=id,command=cmd,env=extra,exit=r.returncode,wall=time.monotonic()-t,failed=[e['Test'] for e in es if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test']],passed=[e['Test'] for e in es if e.get('Action')=='pass' and e.get('Test') and '/' not in e['Test']],skipped=[e['Test'] for e in es if e.get('Action')=='skip' and e.get('Test')],cooked=any('test timed out' in e.get('Output','') for e in es) or r.returncode==124)
 records.append(out);(p/'runs.json').write_text(json.dumps(records,indent=2));print(id,r.returncode,round(out['wall'],2),out['failed'],flush=True);return out
records=json.loads((p/'runs.json').read_text())
(p/'D1.log').rename(p/'D1-cooked.log')
for r in records:
 if r['id']=='D1':r['id']='D1-cooked'
plan=json.loads((p/'mutants.json').read_text())
for m in plan:
 id=m['id'];f=m['file'];s=(root/f).read_text(); subset=[n for n in names if ('ToStringWithARadix' in n if id=='D1' else any(x in n for x in ['RegExp','Regex','Record','RuntimeStringEquality','StringsMatchJavaScript']) if id=='D2' else any(x in n for x in ['RegExp','Regex']))]; pattern='^('+'|'.join(subset)+')$'; (p/(id+'-scope.json')).write_text(json.dumps(subset,indent=2))
 try:
  (root/f).write_text(s.replace(m['before'],m['after']));extra={'ADAMIC_BUILD_CACHE_DIR':'/tmp/defend-radix/cache/'+id}
  if f.endswith('.go'):v=run(id+'-compile','timeout 90 go vet ./internal/regexp/',extra)
  else:
   tu='internal/native/runtime/string.c' if f.endswith('.h') else f
   v=run(id+'-compile',f'clang -std=c11 -Wall -Wextra -Werror -Wcast-function-type-strict -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O1 -g -fsanitize=address,undefined -fno-omit-frame-pointer -DADAMIC_COUNT -I internal/native/runtime -c {tu} -o /tmp/defend-{id}.o',extra)
  if v['exit']:continue
  d=run(id,f"timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '{pattern}'",extra)
  if d['cooked']:break
 finally:(root/f).write_text(s)
 run(id+'-apply',f'git read-tree {base} && git apply --cached --check {p.relative_to(root)}/{id}.diff',{'GIT_INDEX_FILE':'/tmp/defend-radix-index'})
