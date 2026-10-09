from pathlib import Path
import os,time,json,subprocess,difflib
p=Path('review/test-defend/internal-lower-prototype')
plans=[
 {'id':'D1','target':'TestRegExpNativeRefusals','file':'internal/regexp/native.go','old':'\t\t\tif i.min != nil && !i.min.IsUint64() || i.max != nil && !i.max.IsUint64() {\n\t\t\t\treturn fmt.Errorf("native regexp quantifier bounds above uint64 are not yet supported")\n\t\t\t}\n','new':'','kind':'drop statement','reason':'Drop the counter-width refusal; huge quantifier counts truncate during Uint64 serialization.'},
 {'id':'D2','target':'TestRepresentationClockSourceCheckedTypes','file':'internal/lower/expression.go','old':'substituted, isKnown := l.substitution[proven]\n\t\treturn substituted, isKnown','new':'substituted, _ := l.substitution[proven]\n\t\treturn substituted, true','kind':'change constant','reason':'Claim unresolved generic parameters have a known ABI even without a substitution; discard now-unused known binding.'},
 {'id':'D3','target':'TestSuppressionDirectiveSoundNeighbors','file':'internal/lower/refusals.go','old':'func (l *lowering) refuse(module *ast.SourceFile) error {','new':'func (l *lowering) refuse(module *ast.SourceFile) error {\n\tif len(module.CommentDirectives) == 0 && strings.Contains(module.Text(), "@ts-expect-error") {\n\t\treturn &Refused{Where: l.program.FileName(module), What: "suppression directive", Fix: "remove it and fix the type error"}\n\t}','kind':'return early','reason':'Early reject the directive token lexically even when parser says prose/literals contain no directive. No input path/name selection.'}
]
for m in plans:
 original=Path(m['file']).read_text();assert original.count(m['old'])==1
 m['line']=original[:original.index(m['old'])].count('\n')+1
 (p/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),original.replace(m['old'],m['new'],1).splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
(p/'plan.json').write_text(json.dumps(plans,indent=2))
results=[]
def run(argv,name,env):
 start=time.monotonic()
 with (p/(name+'.log')).open('w') as log: r=subprocess.run(argv,stdout=log,stderr=subprocess.STDOUT,env=env)
 events=[]
 for line in (p/(name+'.log')).read_text().splitlines():
  try:events.append(json.loads(line))
  except:pass
 item={'id':name,'command':argv,'exit':r.returncode,'wall':time.monotonic()-start,'failed':sorted(set(x['Test'].split('/')[0] for x in events if x.get('Test') and x.get('Action')=='fail')),'passed':sorted(set(x['Test'] for x in events if x.get('Test') and '/' not in x['Test'] and x.get('Action')=='pass')),'skipped':sorted(set(x['Test'] for x in events if x.get('Test') and x.get('Action')=='skip')),'binary_seconds':next((x.get('Elapsed') for x in events[::-1] if not x.get('Test') and x.get('Action') in ['pass','fail']),None),'fail_lines':[x.get('Output','').strip() for x in events if x.get('Test') and any(s in x.get('Output','') for s in ['expected loud','representation =','sound neighbor:','panic:'])]}
 results.append(item);(p/'results.json').write_text(json.dumps(results,indent=2));print(name,r.returncode,item['wall'],item['failed'],flush=True)
 return item
for m in plans:
 path=Path(m['file']);original=path.read_text()
 env=os.environ|{'ADAMIC_GATE_UNCACHED':'1','ADAMIC_BUILD_CACHE_DIR':'/tmp/defend-prototype/cache/'+m['id']}
 try:
  path.write_text(original.replace(m['old'],m['new'],1))
  check=run(['go','vet','./'+str(path.parent)+'/'],m['id']+'-vet',env)
  if check['exit']:break
  matrix=run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'],m['id'],env)
 finally:path.write_text(original)
