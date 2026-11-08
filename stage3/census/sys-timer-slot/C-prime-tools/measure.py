import sys
sys.dont_write_bytecode=True
from run import *
import shutil,hashlib
base=pathlib.Path('/workspace/scratch/sys-timer-property/H');old=pathlib.Path('/workspace/scratch/sys-timer-property-prime')
def ignore(directory,names):
 if pathlib.Path(directory)==base:return ['built','.git','node_modules']
 if pathlib.Path(directory)==base/'tests/baselines':return ['local']
 return []
shutil.copytree(base,S/'CP',ignore=ignore);(S/'CP/node_modules').symlink_to(ROOT/'stage3/typescript/node_modules',target_is_directory=True)
for name in ['check.cjs','timer-row.json','host_axioms.json']:shutil.copyfile(old/name,S/name)
expected=json.loads((ROOT/'stage3/census/sys-timer-slot/D-source-scope.json').read_text())['H'];actual={str(p.relative_to(S/'CP')):hashlib.sha256(p.read_bytes()).hexdigest() for p in (S/'CP/src').rglob('*') if p.is_file()};assert actual==expected
assert run('method-owner-replay',['node','stage3/adapt/99-correctness-core-sys-timer-slot/adapt.cjs',S/'CP'])==0
assert run('CP-stock',['node','--max-old-space-size=8192',S/'check.cjs',S/'CP',ROOT,S/'CP-stock.json'])==0
r=json.loads((S/'CP-stock.json').read_text());print('stock adamic='+str(len(r['adamic']))+' upstream='+str(len(r['solution']))+' solutionExit='+str(r['solutionExit']),flush=True)
assert run('CP-latent',[old/'census',S/'CP/src/compiler',S/'CP-latent.jsonl'],extra={'LATENT_FULL':'1','LATENT_ONLY_FILE':str(S/'CP/src/compiler/sys.ts'),'LATENT_ASSERT_NO_OUTPUT':'1','LATENT_TIMES':str(S/'CP-times.jsonl')})==0
for name in expected:assert hashlib.sha256((S/'CP'/name).read_bytes()).hexdigest()==expected[name],name
(S/'source-identity.json').write_text(json.dumps({'sourceFiles':len(actual),'head2SourceIdentical':True,'head2':'66e4051f19854166c95df6be99a4ae0e2eeaadd8','hostRowIdenticalToDprime':(S/'timer-row.json').read_bytes()==(old/'timer-row.json').read_bytes(),'hostTableIdenticalToDprime':(S/'host_axioms.json').read_bytes()==(old/'host_axioms.json').read_bytes(),'censusBinarySha256':hashlib.sha256((old/'census').read_bytes()).hexdigest(),'sourceHashes':actual},indent=2)+'\n')
print('C-prime source and shared stand-in verified',flush=True)
