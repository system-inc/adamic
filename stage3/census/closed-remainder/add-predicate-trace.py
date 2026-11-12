import pathlib,json,sys
repo,out=map(pathlib.Path,sys.argv[1:]); m=json.loads((out/'overlay.json').read_text());s=(repo/'internal/lower/predicates_checked.go').read_text();needle='message func(branch string) string) ([]ir.Statement, error) {';assert s.count(needle)==1
s=s.replace(needle,'''message func(branch string) string) (statements []ir.Statement, traceFailure error) {
 if os.Getenv("LWP_TRACE")!="" { defer func(){ reason:=""; if traceFailure!=nil {reason=traceFailure.Error()}; lwp2Write(map[string]any{"event":"predicate_check","unit":latentFindingOwner,"phase":latentFindingPhase,"where":l.program.Where(node),"source":l.typeString(source),"target":l.typeString(target),"false_too":falseToo,"failure":reason}) }() }
''');s=s.replace('import (','import (\n"os"',1);p=out/'predicates_checked.go';p.write_text(s);m['Replace'][str(repo/'internal/lower/predicates_checked.go')]=str(p);(out/'overlay.json').write_text(json.dumps(m,indent=2)+'\n')
