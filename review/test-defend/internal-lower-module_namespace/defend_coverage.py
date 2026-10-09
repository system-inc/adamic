import pathlib,subprocess,json,time
p=pathlib.Path('review/test-defend/internal-lower-module_namespace'); (p/'coverage').mkdir(exist_ok=True)
ref='origin/test-audit/internal-lower-module_namespace'; root='review/test-audit/internal-lower-module_namespace/'
audit=json.loads(subprocess.check_output(['git','show',ref+':'+root+'report.json']))
targets=['TestModuleNamespaceInitializedReadProof','TestCallableNamespaceReceiverStaysLoud','TestNamespaceClassEarlyConstructionStaysLoud','TestNamespaceAmbientHostInitialization','TestNamespaceAmbientContextsDoNotExecute','TestNamespaceCallGraphLinearWork','TestNamespaceCallGraphCycleUnion','TestNamespaceEnumInitializationIndependentOfModuleAnalysis','TestParserFactoryBindingHoisting','TestTscNamespaceDeclarationShapes']
selected=[r for r in audit if r['test'] in targets]; (p/'audit-rows.json').write_text(json.dumps(selected,indent=2))
names=sorted(set(targets+[s for r in selected for s in r['subsumed_by']]))
alltests=[s for s in (p/'logs/list.log').read_text().splitlines() if s.startswith('Test')]
(p/'scope.json').write_text(json.dumps({'base':subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),'targets':targets,'package_tests':alltests,'missing':[s for s in targets if s not in alltests]},indent=2))
results=[]
for name in names+['REST-'+r['test'] for r in selected if r['verdict']=='untrue']:
 regex='^'+name+'$' if not name.startswith('REST-') else '^('+'|'.join(s for s in alltests if s!=name[5:])+')$'
 cmd=['timeout','120','go','test','-count=1','-timeout','90s','./internal/lower/','-run',regex,'-coverpkg=./internal/lower','-coverprofile='+str(p/'coverage'/ (name+'.out'))]
 t=time.monotonic()
 with (p/'logs'/('coverage-'+name+'.log')).open('w') as log: code=subprocess.call(cmd,stdout=log,stderr=log)
 results.append({'name':name,'command':cmd,'exit':code,'seconds':time.monotonic()-t}); (p/'coverage-runs.json').write_text(json.dumps(results,indent=2));print(name,code,round(results[-1]['seconds'],2),flush=True)
 if code: break
