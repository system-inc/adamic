from pathlib import Path
import subprocess,json,difflib,os,time
p=Path('review/test-audit/stage1-cohere-estree-recovery_lowered_recipe');root=Path.cwd();e=os.environ.copy();e.update(ADAMIC_ESTREE_LIBRARY='/tmp/u087-estree-library',ADAMIC_BUILD_CACHE_DIR='/tmp/u087/cache/baseline');records=[]
def run(id,name,regex,overlay=None,cache=None):
 env=e.copy()
 if cache:env['ADAMIC_BUILD_CACHE_DIR']=cache
 cmd=['timeout','120','go','test']+(['-overlay='+overlay] if overlay else [])+['-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run',regex];s=time.monotonic()
 with (p/(id+'.'+name+'.log')).open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 records.append(dict(id=id,test=name,command=cmd,exit=r.returncode,seconds=time.monotonic()-s,cache=env['ADAMIC_BUILD_CACHE_DIR']));(p/'construction-commands.json').write_text(json.dumps(records,indent=2)+'\n');print(id,name,r.returncode,round(records[-1]['seconds'],2),flush=True)
 return r.returncode
def overlay(id,file,before,after):
 text=(root/file).read_text();assert text.count(before)==1,(id,text.count(before));changed=text.replace(before,after)
 if id=='S03': changed=changed.replace('\t"fmt"\n','')
 if id=='P02':
  changed=changed.replace('\t"fmt"\n','').replace('\t"path/filepath"\n','')
 (p/'diffs'/(id+'.diff')).write_text(''.join(difflib.unified_diff(text.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 dest=p/(id+'.go.txt');dest.write_text(changed);ov=p/(id+'-overlay.json');ov.write_text(json.dumps({'Replace':{str((root/file).resolve()):str(dest.resolve())}}))
 with (p/(id+'.vet.log')).open('w') as f:v=subprocess.run(['timeout','90','go','vet','-overlay='+str(ov.resolve()),'./stage1/cohere/estree/'],stdout=f,stderr=subprocess.STDOUT)
 records.append(dict(id=id,validation='go vet',exit=v.returncode));return str(ov.resolve())
# An isolated preparation control may fit after partial products from prior bounded runs.
setup_green=True
for n in range(1,4):
 if run('baseline-'+str(n),'TestRecoveryMutants_Setup','^TestRecoveryMutants_Setup$'):
  setup_green=False; break
if setup_green:
 for n in range(1,4):
  if run('baseline-'+str(n),'TestRecoveryMutants_family','^(TestRecoveryMutants|TestRecoveryMutants_000|TestRecoveryMutants_001|TestRecoveryMutants_002|TestRecoveryMutantsUnion)$'):
   setup_green=False; break
 if setup_green:
  run('W01','TestRecoveryMutants_family','^(TestRecoveryMutants|TestRecoveryMutants_000|TestRecoveryMutants_001|TestRecoveryMutants_002|TestRecoveryMutantsUnion)$',str((p/'W01-overlay.json').resolve()))

# Finish three baseline timings for preparation-only rows.
rows=[('TestScalarEdges preparation family','^(TestScalarEdges|TestScalarEdges_Setup)$'),('TestProduct_scalar_edges_go_oracle','^TestProduct_scalar_edges_go_oracle$'),('TestProduct_scalar_edges_lowered','^TestProduct_scalar_edges_lowered$'),('TestProduct_scalar_edges_sanitized_native','^TestProduct_scalar_edges_sanitized_native$')]
for name,regex in rows:
 for n in range(1,4):
  if run('baseline-'+str(n),name.replace(' ','_'),regex):break
# Source always remains original; overlays change only the permitted construction or witness.
text=(root/'stage1/cohere/estree/estree_test.go').read_text(); start=text.index('func firstDifference('); body=text[start:text.index('\n}',start)+2]
ov=overlay('W02','stage1/cohere/estree/estree_test.go',body,'func firstDifference(want, got []byte) string { return "" }')
run('W02','TestScalarEdgesPlantedFailure','^TestScalarEdgesPlantedFailure$',ov)
ov=overlay('S01','stage1/cohere/estree/scalar_edges_split_products_test.go','scalarEdgePreparation.ready = ready','// readiness publication deliberately dropped')
run('S01','TestScalarEdges_preparation_family','^(TestScalarEdges|TestScalarEdges_Setup)$',ov)
ov=overlay('S02','stage1/cohere/estree/scalar_edges_split_products_test.go','if err := os.WriteFile(filepath.Join(dir, "port.c"), []byte(native.C(ir)), 0644); err != nil {\n\t\t\treturn err\n\t\t}', '// C artifact write deliberately dropped')
run('S02','TestProduct_scalar_edges_lowered','^TestProduct_scalar_edges_lowered$',ov,'/tmp/u087/cache/S02')
ov=overlay('S03','stage1/cohere/estree/scalar_edges_split_products_test.go','output, err := command.CombinedOutput()\n\t\tif err != nil {\n\t\t\treturn fmt.Errorf("oracle: %w\\n%s", err, output)\n\t\t}', '// oracle build deliberately dropped')
run('S03','TestProduct_scalar_edges_go_oracle','^TestProduct_scalar_edges_go_oracle$',ov,'/tmp/u087/cache/S03')
ov=overlay('S04','stage1/cohere/estree/scalar_edges_split_products_test.go','return native.Build(string(source), filepath.Join(dir, "port"), native.Options{Sanitize: true, Split: true})','_ = source\n return nil')
run('S04','TestProduct_scalar_edges_sanitized_native','^TestProduct_scalar_edges_sanitized_native$',ov,'/tmp/u087/cache/S04')
# Break the independence of the recipe's two input paths while retaining all variable uses.
ov=overlay('S05','stage1/cohere/estree/recovery_lowered_recipe_test.go','[]string{main, copied}','[]string{main, main + copied[len(copied):]}')
run('S05','TestRecoveryLoweredRecipe','^TestRecoveryLoweredRecipe$',ov)
text=(root/'internal/lower/lower.go').read_text(); start=text.index('func Lower('); body=text[start:text.index('\n}',start)+2]
ov=overlay('P02','internal/lower/lower.go',body,'func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) { return &ir.Program{}, nil }')
run('P02','TestRecoveryLoweredRecipe','^TestRecoveryLoweredRecipe$',ov,'/tmp/u087/cache/P02')
run('restored','control','^(TestRecoveredGrammar|TestScalarEdgesUnion|TestScalarEdges_000|TestScalarEdges_001|TestScalarEdges_002|TestScalarEdges_003|TestRecoveryMutantsShardSurvivor|TestRecoveryMutantsTopSurvivor|TestScalarEdgesPlantedFailure|TestRecoveryLoweredRecipe)$')
