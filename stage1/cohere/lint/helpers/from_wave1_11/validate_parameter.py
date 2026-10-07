import argparse,json,subprocess,os,re,shutil,hashlib,posixpath
from pathlib import Path
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[4];COHERE=ROOT/'cohere'
p=argparse.ArgumentParser();p.add_argument('--scratch',type=Path,required=True);p.add_argument('--replay',action='store_true');a=p.parse_args();S=a.scratch.resolve();S.mkdir(parents=True,exist_ok=True)
def run(label,cmd,cwd=ROOT,env=None):
 with (S/(label+'.log')).open('wb') as out,(S/(label+'.stderr')).open('wb') as err:r=subprocess.run(list(map(str,cmd)),cwd=cwd,env=env,stdout=out,stderr=err)
 assert r.returncode==0 and not (S/(label+'.stderr')).read_bytes(),label
 return (S/(label+'.log')).read_bytes()
if not a.replay:
 consumers=[('structure/'+slug.replace('_','-'),'structure',slug,prefix) for slug,prefix in [('network_require_hook_options_parameter','TestNetworkRequireHookOptionsParameter'),('network_require_hook_variables_type','TestNetworkRequireHookVariablesType'),('next_require_api_parameter_name','TestNextRequireApiParameterName'),('react_component_require_properties_type_suffix','TestReactComponentRequirePropertiesTypeSuffix')]]
 source=COHERE/'internal/lint/rules/structure/network_file_analysis.go';text=source.read_text();anchor='func parameterTypeNode(';assert text.count(anchor)==1;patched=S/'parameters.go';patched.write_text(text.replace(anchor,'func wave11OriginalParameterTypeNode(',1))
 extra=source.parent/'wave11_capture.go';overlay=S/'capture-overlay.json';overlay.write_text(json.dumps({'Replace':{str(source):str(patched),str(extra):str(HERE/'capture_parameter.go.txt')}}));capture=S/'captured.jsonl';env=os.environ|{'WAVE11_PARAMETER_CAPTURE':str(capture)}
 original=[];coverage={}
 for consumer,pkg,slug,prefix in consumers:
  capture.write_text('');run('consumer-Go-'+slug,['go','test','-overlay='+str(overlay),'./internal/lint/rules/'+pkg,'-run','^'+prefix,'-count=1','-v','-timeout=10m'],COHERE,env)
  records=[json.loads(line) for line in capture.read_text().splitlines()];coverage[consumer]=len(records)
  for row in records:row['consumer']=consumer
  original+=records
 assert all(coverage.values()),coverage
 capture.write_text(''.join(json.dumps(row)+'\n' for row in original))
 # Pure predicate: deduplicate metadata, preserving every distinct actual consumer input.
 unique={(r['present'],r['kind'],r['typeId'],r['typeKind']) for r in original}
 rows=[dict(zip(['present','kind','typeId','typeKind'],r)) for r in sorted(unique)]
 for kind in range(501):
  for present in [False,True]:
   for typeId in [-1,0,1,7]:rows.append({'present':present,'kind':kind,'typeId':typeId,'typeKind':80})
 for kind in range(501):rows.append({'present':True,'kind':170,'typeId':7,'typeKind':kind})
 casefile=S/'cases.json';casefile.write_text(json.dumps(rows));print('actual helper calls',len(original),'distinct',len(unique),'coverage',coverage,'total cases',len(rows),flush=True)
else:
 original=[json.loads(line) for line in (HERE/'parameter_capture.jsonl').read_text().splitlines()]
 coverage=json.loads((HERE/'parameter_capture_summary.json').read_text())['consumers']
 rows=json.loads((HERE/'parameter_cases.json').read_text());casefile=S/'cases.json';casefile.write_text(json.dumps(rows))
 print('replaying',len(original),'actual consumer calls through',len(rows),'metadata cases',flush=True)
virtual=COHERE/'wave11_parameter_oracle.go';overlay=S/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(HERE/'oracle_parameter.go.txt'),str(COHERE/'internal/lint/rules/structure/wave11_export.go'):str(HERE/'export_parameter.go.txt')}}));run('Go-build',['go','build','-overlay='+str(overlay),'-o',S/'oracle',virtual],COHERE);want=run('Go',[S/'oracle',casefile])
run('build',['go','run',HERE/'build.go',HERE/'main_parameter.a',S/'native',S/'emitted.mjs'])
def check(label,source,native,js,mutant=False):
 for backend,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',source,casefile]),('emitted',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',js,casefile]),('native',[native,casefile])]:
  actual=run(label+'-'+backend,cmd);assert (actual!=want)==mutant,label+' '+backend;print(label,backend,'byte-only catch' if mutant else 'equal',flush=True)
check('baseline',HERE/'main_parameter.a',S/'native',S/'emitted.mjs')
change=json.loads((HERE/'mutant_parameter.json').read_text());copy=S/'mutant';shutil.copytree(ROOT/'stage1',copy/'stage1',dirs_exist_ok=True);file=copy/HERE.relative_to(ROOT)/change['file'];text=file.read_text();assert text.count(change['from'])==1;file.write_text(text.replace(change['from'],change['to'],1));source=copy/HERE.relative_to(ROOT)/'main_parameter.a';run('mutant-build',['go','run',HERE/'build.go',source,copy/'native',copy/'emitted.mjs']);check(change['name'],source,copy/'native',copy/'emitted.mjs',True)
(S/'summary.json').write_text(json.dumps({'actualCalls':len(original),'consumers':coverage,'cases':len(rows),'bytes':len(want),'stdoutSha256':hashlib.sha256(want).hexdigest(),'mutant':change['name'],'backends':['Node','emitted','native'],'onlyByteComparison':True},indent=2));print('PASS',len(rows),'cases',len(want),'bytes',flush=True)
