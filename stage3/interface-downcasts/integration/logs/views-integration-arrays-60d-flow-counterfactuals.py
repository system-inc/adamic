from pathlib import Path
import subprocess,os
root=Path('/workspace/adamic');path=root/'internal/lower/graph_flow.go';original=path.read_bytes()
before='''\t\tcase ir.ArrayLiteral:
\t\t\tif element := f.arrayLiteralElement(proven); element != nil {
\t\t\t\tfor _, item := range value.Elements {
\t\t\t\t\tmembers(item, element)
\t\t\t\t}
\t\t\t}'''
after='''\t\tcase ir.ArrayLiteral:
\t\t\targuments := f.l.checker.GetTypeArguments(proven)
\t\t\tif len(arguments) > 0 {
\t\t\t\tfor _, element := range value.Elements {
\t\t\t\t\tmembers(element, arguments[0])
\t\t\t\t}
\t\t\t}'''
assert original.decode().count(before)==1
try:
 path.write_text(original.decode().replace(before,after))
 log=Path('/tmp/views-integration-arrays-60d-flow-crash.log')
 with log.open('wb') as out:r=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewRanked4ArrayContracts/ranked4-label-push$','-count=1','-timeout','10m'],cwd=root,stdout=out,stderr=subprocess.STDOUT,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'})
 observed=log.read_text();assert r.returncode!=0 and 'GetTypeArguments' in observed and 'panic:' in observed,observed
 print('Old union GetTypeArguments call: frontend crash reproduced; no runtime mutant credit.',flush=True)
finally:path.write_bytes(original)
before='\telements := []*checker.Type{}'
assert original.decode().count(before)==1
try:
 path.write_text(original.decode().replace(before,'\tif proven.Flags()&checker.TypeFlagsUnion != 0 { return nil }\n'+before))
 log=Path('/tmp/views-integration-arrays-60d-flow-seed-survivor.log')
 with log.open('wb') as out:r=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewRanked4ArrayContracts$','-count=1','-timeout','10m'],cwd=root,stdout=out,stderr=subprocess.STDOUT,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'})
 assert r.returncode==0,log.read_text()
 print('Union-specific element seeds omitted: ranked4 survives; no check proof or mutant credit.',flush=True)
finally:path.write_bytes(original)
