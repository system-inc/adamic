# Run from the repository root with the toolchain environment sourced.
# Arguments: file, old text, new text, oracle test pattern, log label.
from pathlib import Path
import subprocess,sys
file,old,new,pattern,label=sys.argv[1:]
p=Path(file); original=p.read_text(); assert original.count(old)==1, (file,original.count(old))
try:
 p.write_text(original.replace(old,new))
 with open('/tmp/object-prototype-'+label+'.log','w') as log:
  run=subprocess.run(['go','test','-count=1','-timeout','10m','./internal/oracle','-run',pattern],stdout=log,stderr=subprocess.STDOUT)
 output=Path('/tmp/object-prototype-'+label+'.log').read_text()
 assert run.returncode!=0 and ('stdout differs' in output or 'exit codes differ' in output),output
 assert 'Lower:' not in output,output
 print(label+': caught by Node comparison')
finally:
 p.write_text(original)
