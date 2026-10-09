from pathlib import Path
import subprocess,json
out=Path(__file__).parent/"resolved"
source=Path('internal/lower/interface_cast.go');original=source.read_bytes()
try:
 subprocess.run(['git','apply','--unidiff-zero','review/compiler/fx6-candidates-4/mutants/spread-method.patch'],check=True,timeout=10)
 with (out/'original-spread.log').open('w') as log:
  r=subprocess.run(['go','test','./internal/lower','-run','^TestViewSpreadMethodRefused$','-count=1','-v','-timeout','90s'],stdout=log,stderr=subprocess.STDOUT,timeout=120)
 (out/'original-spread.json').write_text(json.dumps({'name':'original-spread-method','exit':r.returncode,'caught':r.returncode==1,'masked_by':'checkViewMembers creation admission'},indent=2)+'\n')
 print('original spread-method mutant exit',r.returncode,flush=True)
finally: source.write_bytes(original)
