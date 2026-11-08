import sys
sys.dont_write_bytecode=True
from run import *
import shutil,hashlib,difflib
assert run('head',['git','rev-parse','HEAD'])==0;assert (S/'head.log').read_text().strip()=='66e4051f19854166c95df6be99a4ae0e2eeaadd8'
assert run('branch',['git','switch','-c','typescript/sys-timer-slot-property'])==0
assert run('restore-history',['git','restore','--source=949b9c3c4c883d964bba19ee9fbdae9fd1909288','--staged','--worktree','--','stage3/census/sys-timer-slot'])==0
base=pathlib.Path('/workspace/scratch/sys-timer-stack/head2')
expected=json.loads(pathlib.Path('/workspace/scratch/sys-timer-stack/head2-source-scope.json').read_text())['after']
def sources(tree):return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (tree/'src').rglob('*') if p.is_file()}
assert sources(base)==expected
assert run('head2-owner-replay',['node','stage3/adapt/99-correctness-core-sys-timer-slot/adapt.cjs',base])==0
for name in ['H','D']:
 def ignore(directory,names):
  if pathlib.Path(directory)==base:return ['.git','node_modules','built']
  if pathlib.Path(directory)==base/'tests/baselines':return ['local']
  return []
 shutil.copytree(base,S/name,ignore=ignore)
 (S/name/'node_modules').symlink_to(ROOT/'stage3/typescript/node_modules',target_is_directory=True)
 assert sources(S/name)==expected
p=S/'D/src/compiler/sys.ts';before=p.read_bytes();after=before
changes=[(b'setTimeout?<A extends unknown[]>(callback: (...args: NoInfer<A>) => void, ms: number, ...args: A): unknown;',b'setTimeout?: <A extends unknown[]>(callback: (...args: NoInfer<A>) => void, ms: number, ...args: A) => unknown;'),(b'clearTimeout?(timeoutId: unknown): void;',b'clearTimeout?: (timeoutId: unknown) => void;')]
for a,b in changes:assert after.count(a)==1;after=after.replace(a,b)
p.write_bytes(after)
(S/'D.diff.txt').write_text(''.join(difflib.unified_diff(before.decode().replace('\r\n','\n').splitlines(True),after.decode().replace('\r\n','\n').splitlines(True),fromfile='head2/src/compiler/sys.ts',tofile='D/src/compiler/sys.ts',n=0)))
shutil.copyfile(ROOT/'stage3/adapt/99-correctness-core-sys-timer-slot/stock-check.cjs',S/'check.cjs')
shutil.copyfile(ROOT/'stage3/census/sys-timer-slot/finish-shim.mjs',S/'finish-shim.mjs')
(S/'source-scope.json').write_text(json.dumps({'sourceFiles':len(expected),'HMatchesAcceptedHead2FullApply':True,'H':sources(S/'H'),'D':sources(S/'D'),'changes':[(a.decode(),b.decode()) for a,b in changes]},indent=2)+'\n')
print('prepared H and D; source-only change is exactly two System member spellings',flush=True)
