from pathlib import Path
import subprocess,json,difflib
p=Path('review/test-audit/stage1-cohere-graphql-printer-gaps');file='internal/lower/class.go';src=subprocess.check_output(['git','show','HEAD:'+file],text=True)
guard='''if field := l.checker.GetSymbolAtLocation(parent.Name()); field != nil && len(field.Declarations) > 0 && (field.Declarations[0].Kind == ast.KindPropertyDeclaration || parameterProperty(field.Declarations[0])) {
			return nil
		}'''
plan=[('M01','last = statement.End()','last = 0','change constant','if auditSelected("M01") { last = 0 } else { last = statement.End() }'),('M02','node.Pos() >= l.unsetUntil','node.Pos() < l.unsetUntil','flip condition','(node.Pos() >= l.unsetUntil) != auditSelected("M02")'),('M03',guard,guard.replace('\n\t\t\treturn nil',''),'drop statement',guard.replace('return nil','if !auditSelected("M03") { return nil }'))]
metadata=[]
for mid,old,new,menu,switch in plan:
 assert src.count(old)==1,(mid,src.count(old))
 changed=src.replace(old,new);(p/(mid+'.diff')).write_text(''.join(difflib.unified_diff(src.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 metadata.append(dict(id=mid,file=file,line=src[:src.index(old)].count('\n')+1,before=old,after=new,menu=menu))
(p/'plan.json').write_text(json.dumps(metadata,indent=2))
# All source-derived production changes are fixed before mutant execution.
changed=src
for mid,old,new,menu,switch in plan:changed=changed.replace(old,switch)
Path(file).write_text(changed)
Path('internal/lower/audit_selector.go').write_text('package lower\nimport "os"\nfunc auditSelected(id string) bool { return os.Getenv("ADAMIC_MUTANT") == id }\n')
f='internal/lower/lower.go';base=subprocess.check_output(['git','show','HEAD:'+f],text=True);marker='func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {'
Path(f).write_text(base.replace(marker,marker+'\n\tif auditSelected("P01") { return nil, nil }'))
start=base.index(marker);end=base.index('\ntype lowering struct',start);probe=base[:start]+marker+'\n\treturn nil, nil\n}\n'+base[end:]
# Keep only imports still used by the remaining file.
probe=probe.replace('\n\t"fmt"','').replace('\n\t"path/filepath"','')
(p/'P01.diff').write_text(''.join(difflib.unified_diff(base.splitlines(True),probe.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
# Fix checker and construction faults before their runs, in isolated test-file overlays.
checks=[{'id':'W01','file':'stage1/cohere/graphql/printer/shards_test.go','change':'printerShardDisagreement returns nil at entry','kind':'weaken comparison'}, {'id':'S01','file':'stage1/cohere/graphql/printer/gaps_test.go','change':'whitespaceShards allocation bound adds one','kind':'off-by-one construction bound'},{'id':'S02','file':'stage1/cohere/graphql/printer/gaps_test.go','change':'drop entire case-assignment loop','kind':'drop construction loop'},{'id':'P02','file':'stage1/cohere/graphql/printer/gaps_test.go','change':'whitespaceShards returns nil at entry','kind':'empty construction probe'}]
(p/'check-plan.json').write_text(json.dumps(checks,indent=2))
(p/'scope.json').write_text(json.dumps({'base':subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),'rows':['TestPrinterConstructorGap','TestPrinterWhitespacePlantedDisagreement','TestPrinterWhitespaceShardGrowth'],'bounded':True,'lowering_inventory':'reached-lowering-functions.txt','harness_reach':{'TestPrinterWhitespacePlantedDisagreement':['whitespaceCaseID','whitespaceOwner','whitespaceShards','printerShardDisagreement','firstDifference'],'TestPrinterWhitespaceShardGrowth':['whitespaceCaseID','whitespaceOwner','whitespaceShards','printerShardUnion']},'nproc':5,'setup_seconds':0},indent=2))
