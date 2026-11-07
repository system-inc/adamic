from run import *
import shutil,hashlib
rules=json.loads((ROOT/'stage3/adapt/41-explicit-any-remaining/rules.json').read_text());files=sorted({r['file'] for r in rules});tree=S/'main-replay';tree.mkdir()
for file in files:
 p=tree/file;p.parent.mkdir(parents=True,exist_ok=True);shutil.copy2(S/'main'/file,p)
def hashes(root):return {file:hashlib.sha256((root/file).read_bytes()).hexdigest() for file in files}
before=hashes(tree);code=run('main-full-tree-replay',['node',S/'base-repo/stage3/adapt/41-explicit-any-remaining/adapt.cjs',tree]);assert code==1;assert before==hashes(tree)
assert 'missing or duplicate reviewed site: 31' in (S/'main-full-tree-replay.log').read_text();assert 'missing or duplicate reviewed site: 31' in (S/'head1-replay.log').read_text()
(S/'replay-control.json').write_text(json.dumps({'originalMainFullTreeReplayExit':1,'amendedFullTreeReplayExit':1,'samePreexistingRule':31,'originalReplayNoWrite':True,'isolatedOwnStageIdempotenceExit':0,'action':'No unrelated guard changed; own-stage replay and LF/CRLF contract pass.'},indent=2)+'\n')
