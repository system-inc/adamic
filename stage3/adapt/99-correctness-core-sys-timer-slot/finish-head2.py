import sys
sys.dont_write_bytecode=True
from run import *
import hashlib,shutil,difflib
base=S/'head1';after=S/'head2';D=ROOT/'stage3/adapt/99-correctness-core-sys-timer-slot'
def hashes(tree):return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (tree/'src').rglob('*') if p.is_file()}
a=hashes(base);b=hashes(after);candidate=hashes(S/'head2-candidate');assert b==candidate
changed=[k for k in a if a[k]!=b.get(k)];assert changed==['src/compiler/sys.ts'];assert set(a)==set(b)
(S/'head2-source-scope.json').write_text(json.dumps({'sourceFiles':len(a),'changed':changed,'actualFullApplyEqualsCensusCandidate':True,'before':a,'after':b},indent=2)+'\n')
old=(base/'src/compiler/sys.ts').read_text();new=(after/'src/compiler/sys.ts').read_text();(S/'head2-source.diff.txt').write_text(''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='head1/src/compiler/sys.ts',tofile='head2/src/compiler/sys.ts',n=0)))
roots=sorted(str(p) for p in (after/'src/compiler').rglob('*.ts'));assert len(roots)==79;assert run('head2-actual-types',[S/'bin/adamic','types',*roots])==0
assert run('head2-actual-latent',[S/'bin/census',after/'src/compiler',S/'head2-actual-latent.jsonl'],extra={'LATENT_FULL':'1','LATENT_ONLY_FILE':str(after/'src/compiler/sys.ts'),'LATENT_ASSERT_NO_OUTPUT':'1','LATENT_TIMES':str(S/'head2-actual-times.jsonl')})==0
expected=json.loads((D/'expected-new-sites.json').read_text())
comparison=json.loads((S/'head2-comparison.json').read_text())
assert [{k:v for k,v in x.items() if k!='rawWhere'} for x in comparison['head2']['new']]==expected
assert comparison['head2']['unexpected']==[]
def ledger(label,tree):
 rows=[json.loads(l) for l in (S/(label+'-latent.jsonl')).read_text().splitlines()];assert rows[0]['diagnostics']==[]
 return sorted((f['where'].replace(str(tree),''),f['kind'],f['reason'],f['phase']) for f in rows[1]['findings'] if f['kind'] in ['Refused','NotYet'])
assert ledger('head2-actual',after)==ledger('head2-candidate',S/'head2-candidate'),'actual census differs; stop'
(after/'node_modules').symlink_to(ROOT/'stage3/typescript/node_modules',target_is_directory=True)
control=json.loads((S/'head1-stock.json').read_text());assert control['adamic']==[] and control['solution']==[] and control['solutionExit']==0
for label,tree in [('head2',after)]:
 assert run(label+'-stock',['node',S/'stock-check.cjs',tree,ROOT,S/(label+'-stock.json')])==0
 result=json.loads((S/(label+'-stock.json')).read_text());assert result['adamic']==[] and result['solution']==[] and result['solutionExit']==0
assert run('head2-build-compiler',['npm','run','build:compiler','--','--lkg'],after)==0
oldjs=json.loads((S/'main-js.json').read_text());js={name:hashlib.sha256((after/'built/local'/name).read_bytes()).hexdigest() for name in oldjs};assert len(js)==9;assert js==oldjs
for name in js:assert (after/'built/local'/name).read_bytes()==(base/'built/local'/name).read_bytes()==(S/'main/built/local'/name).read_bytes()
outputs={}
for name in ['typescript.d.ts','typescript.internal.d.ts','tsserverlibrary.d.ts']:
 x=(base/'built/local'/name).read_text();y=(after/'built/local'/name).read_text();diff=''.join(difflib.unified_diff(x.splitlines(True),y.splitlines(True),fromfile='head1/'+name,tofile='head2/'+name));(S/('head2-'+name+'.diff.txt')).write_text(diff);outputs[name]={'beforeSha256':hashlib.sha256(x.encode()).hexdigest(),'afterSha256':hashlib.sha256(y.encode()).hexdigest(),'changed':x!=y}
(S/'head2-artifacts.json').write_text(json.dumps({'javascript':js,'javascriptCount':9,'allJavascriptIdentical':True,'declarations':outputs},indent=2)+'\n')
assert json.loads((S/'head1-build-tests-status.json').read_text())['exit']==0
assert run('head2-build-tests',['npm','run','build:tests','--','--lkg'],after)==0
assert run('head2-replay',['node',D/'adapt.cjs',after])==0
(after/'node_modules').unlink()
