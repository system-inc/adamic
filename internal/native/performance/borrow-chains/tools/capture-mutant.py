"""Make the captured chained string borrowed in generated C, then run ASan."""
from pathlib import Path
import subprocess, re, json
root=Path(__file__).resolve().parents[5]
source=subprocess.check_output(['/tmp/borrow-chains-after','c',str(root/'internal/oracle/testdata/borrow_chain_capture.a')],cwd=root).decode()
pattern=r'(adamic_cell_new\(\(adamic_value\)\{\.reference = )adamic_retain\((adamic_temporary_\d+)\)(\}, true\))'
source,count=re.subn(pattern,r'\1\2\3',source)
assert count==1
path=Path('/tmp/borrow-chains-capture-mutant.c');path.write_text(source)
runtime=root/'internal/native/runtime'
command=['clang','-std=c11','-Wall','-Wextra','-Werror','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all','-I',str(runtime),'-o','/tmp/borrow-chains-capture-mutant',str(path),*[str(p) for p in sorted(runtime.glob('*.c'))],'-lm']
with open('/tmp/borrow-chains-mutant-capture-build.log','wb') as log:subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,check=True)
with open('/tmp/borrow-chains-mutant-capture.log','wb') as log:result=subprocess.run(['/tmp/borrow-chains-capture-mutant'],stdout=log,stderr=subprocess.STDOUT)
text=Path('/tmp/borrow-chains-mutant-capture.log').read_text()
summary={'mutant':'captured chain stored in a cell without retain','exit':result.returncode,'killed':'heap-use-after-free' in text,'command':command}
Path('/tmp/borrow-chains-capture-mutant.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary,indent=2));assert summary['killed']
