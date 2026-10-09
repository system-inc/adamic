from pathlib import Path
import subprocess,json
E=Path('/workspace/adamic/review/test-audit/internal-unicodeproperties-alias');scratch=Path('/tmp/u075-vet');dest=scratch/'internal/unicodeproperties'
original={f:Path('/tmp/u075_'+f).read_text()for f in ('unicodeproperties.go','tables.go')}
(dest/'audit_witness_test.go').write_text('''package unicodeproperties
import("fmt";"testing")
func TestAuditBehaviorWitness(t *testing.T) { fmt.Printf("expressionOK(empty)=%v expressionOK(gc=Lu=X)=%v\\n",expressionOK(""),expressionOK("gc=Lu=X")); a,ok:=Lookup("ASCII",false);fmt.Printf("Lookup(ASCII)=%v Contains(127)=%v ranges=%v\\n",ok,a.Contains(127),a.Set.Ranges) }
''')
for mid in ['clean','M15','M16','M20']:
 for f,s in original.items():(dest/f).write_text(s)
 if mid!='clean':subprocess.run(['git','apply',str(E/'diffs'/f'{mid}.diff')],cwd=scratch,check=True)
 with (E/('witness-'+mid+'.log')).open('w')as out:subprocess.run(['go','test','-v','-count=1','./internal/unicodeproperties/','-run','^TestAuditBehaviorWitness$'],cwd=scratch,stdout=out,stderr=subprocess.STDOUT)
