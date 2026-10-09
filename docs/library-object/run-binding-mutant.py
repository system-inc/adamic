"""Prove that a replaced literal binding cannot authorize a uniform Object.values read."""
import os
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[2]
scratch = Path('/tmp/object-binding-mutant')
scratch.mkdir(exist_ok=True)
path = root / 'internal/lower/library_object.go'
source = root / 'internal/oracle/testdata/library_object_replaced.a'
original = path.read_bytes()
before = 'declaration.Parent.Flags&ast.NodeFlagsConst == 0 && l.objectBindingAssigned(declaration, symbol)'
assert original.decode().count(before) == 1
binary = scratch / 'adamic'
module = scratch / 'source.mts'
module.write_bytes(source.read_bytes())

def run(arguments):
 result = subprocess.run([str(item) for item in arguments], cwd=root, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
 print('command:', ' '.join(map(str, arguments)), 'exit:', result.returncode, flush=True)
 print(result.stdout.decode(errors='replace'), result.stderr.decode(errors='replace'), flush=True)
 return result

assert run(['go', 'build', '-o', binary, './cmd/adamic']).returncode == 0
baseline = run([binary, 'c', source])
assert baseline.returncode == 1 and b"shape not proven" in baseline.stderr
node = run(['node', '--disable-warning=ExperimentalWarning', module])
assert node.returncode == 0
try:
 path.write_text(original.decode().replace(before, 'false', 1))
 assert run(['go', 'build', '-o', binary, './cmd/adamic']).returncode == 0
 assert run([binary, 'build', source, '-o', scratch / 'mutant.bin']).returncode == 0
 native = run([scratch / 'mutant.bin'])
 assert native.returncode == 0 and native.stderr == b''
 assert native.stdout != node.stdout
 print('replaced-binding-trusted: caught only by Node stdout comparison', flush=True)
finally:
 path.write_bytes(original)
