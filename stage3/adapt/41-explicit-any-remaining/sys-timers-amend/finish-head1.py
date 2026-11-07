from run import *
import hashlib,shutil
base=S/'main';after=S/'head1'
def hashes(tree):return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (tree/'src').rglob('*') if p.is_file()}
a=hashes(base);b=hashes(after);candidate=hashes(S/'head1-candidate');assert b==candidate
changed=[k for k in a if a[k]!=b.get(k)];assert changed==['src/compiler/sys.ts'];assert set(a)==set(b)
(S/'head1-source-scope.json').write_text(json.dumps({'sourceFiles':len(a),'changed':changed,'actualFullApplyEqualsCensusCandidate':True,'before':a,'after':b},indent=2)+'\n')
roots=sorted(str(p) for p in (after/'src/compiler').rglob('*.ts'));assert len(roots)==79;assert run('head1-actual-types',[S/'bin/adamic','types',*roots])==0
(after/'node_modules').symlink_to('/workspace/scratch/sys-timer-slot/deps/node_modules',target_is_directory=True)
assert run('head1-build-compiler',['npm','run','build:compiler','--','--lkg'],after)==0
js={str(p.relative_to(after/'built/local')):hashlib.sha256(p.read_bytes()).hexdigest() for p in (after/'built/local').rglob('*.js')};old=json.loads((S/'main-js.json').read_text());assert len(js)==9;assert js==old
outputs={}
for name in ['typescript.d.ts','typescript.internal.d.ts','tsserverlibrary.d.ts']:
 x=(base/'built/local'/name).read_bytes();y=(after/'built/local'/name).read_bytes();assert x==y;outputs[name]={'bytes':len(x),'sha256':hashlib.sha256(x).hexdigest(),'identical':True}
(S/'head1-artifacts.json').write_text(json.dumps({'javascript':js,'javascriptCount':9,'allJavascriptIdentical':True,'declarations':outputs},indent=2)+'\n')
replayExit=run('head1-replay',['node',ROOT/'stage3/adapt/41-explicit-any-remaining/adapt.cjs',after])
if replayExit:
 assert replayExit==1 and 'missing or duplicate reviewed site: 31' in (S/'head1-replay.log').read_text()
 control=json.loads((S/'replay-control.json').read_text())
 assert control['originalMainFullTreeReplayExit']==1 and control['samePreexistingRule']==31
 print('Finished-tree replay retains the paired, pre-existing rule-31 stop; own-stage contract passes.',flush=True)
