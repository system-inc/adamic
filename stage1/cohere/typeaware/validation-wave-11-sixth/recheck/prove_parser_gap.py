# Run from the repository root with the Adamic toolchain environment sourced.
from pathlib import Path
import subprocess,json,sys
scratch=Path(sys.argv[1]).resolve();scratch.mkdir(parents=True,exist_ok=True)
source=scratch/'jsx-control.tsx'
source.write_text('declare function make(): () => unknown;\nfunction Component() { const C = make(); return <C />; }\n')
compiler=Path(sys.argv[2]).resolve()
with (scratch/'build.stdout').open('wb') as out,(scratch/'build.stderr').open('wb') as err:
 subprocess.run([str(compiler),'build','stage1/typescript/parser/main.ts','-o',str(scratch/'parser')],stdout=out,stderr=err,check=True)
with (scratch/'native.stdout').open('wb') as out,(scratch/'native.stderr').open('wb') as err:
 result=subprocess.run([str(scratch/'parser'),str(source),'--whole'],stdout=out,stderr=err,timeout=10)
error=(scratch/'native.stderr').read_text()
assert result.returncode==70,result.returncode
assert 'parser slice expected GreaterThanToken, got SlashToken at 91' in error,error
assert (scratch/'native.stdout').read_bytes()==b''
print(json.dumps({'native_exit':result.returncode,'blocker':error.strip()}))
