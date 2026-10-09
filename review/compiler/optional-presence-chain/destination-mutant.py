from pathlib import Path
import json,subprocess
r=Path.cwd();out=r/'review/compiler/optional-presence-chain';p=r/'internal/lower/optional_fields.go';s=p.read_text();before='own[index].Optional = field != nil && field.Flags&ast.SymbolFlagsOptional != 0';assert s.count(before)==1
mut=out/'destination-metadata.go.txt';mut.write_text(s.replace(before,'own[index].Optional = false && field != nil && field.Flags&ast.SymbolFlagsOptional != 0'))
overlay=out/'destination-metadata.json';overlay.write_text(json.dumps({'Replace':{str(p):str(mut)}}))
with (out/'destination-metadata.log').open('w') as f:
 result=subprocess.run(['go','test','-overlay',str(overlay),'./internal/oracle','-run','TestNativeAgreesWithNode/internal/oracle/testdata/optional_field_checked_view_string_undefined.a$','-count=1','-v','-timeout','90s'],stdout=f,stderr=subprocess.STDOUT,timeout=120)
log=(out/'destination-metadata.log').read_text();assert result.returncode==1 and 'expected string, found undefined' in log and 'build failed' not in log,log
print('Construction optionality mutant caught by Node in both backends: expected string, found undefined')
