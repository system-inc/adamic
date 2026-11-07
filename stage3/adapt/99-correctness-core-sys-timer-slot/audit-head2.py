from pathlib import Path
import json,hashlib,subprocess
R=Path('/workspace/adamic');D=R/'stage3/adapt/99-correctness-core-sys-timer-slot'
S=Path('/workspace/scratch/sys-timer-stack')
for name in json.loads((D/'artifacts.json').read_text())['javascript']:assert (S/'main/built/local'/name).read_bytes()==(S/'head1/built/local'/name).read_bytes()==(S/'head2/built/local'/name).read_bytes()
p=json.loads((D/'proof-results.json').read_text());assert p['status']=='pass' and p['unexpectedNewSites']==[]
for n in ['typescript.d.ts','typescript.internal.d.ts']:
 text=(D/('head2-'+n+'.diff.txt')).read_text();minus=[l[1:].strip() for l in text.splitlines() if l.startswith('-') and not l.startswith('---')];plus=[l[1:].strip() for l in text.splitlines() if l.startswith('+') and not l.startswith('+++')];rows=json.loads((D/'api-sanctions.json').read_text());assert minus==[r['before'] for r in rows];assert plus==[r['after'] for r in rows]
assert (D/'head2-tsserverlibrary.d.ts.diff.txt').read_text()==''
assert hashlib.sha256((D/'evidence.tar.gz').read_bytes()).hexdigest()==(D/'evidence.sha256').read_text().split()[0]
scope=json.loads((D/'source-scope.json').read_text());assert scope['sourceFiles']==746
for label,key in [('head1','before'),('head2','after')]:
 actual={str(f.relative_to(S/label)):hashlib.sha256(f.read_bytes()).hexdigest() for f in (S/label/'src').rglob('*') if f.is_file()};assert actual==scope[key]
for name in ['README.md','PROOF.md']:
 text=(D/name).read_text();assert '\u2013' not in text and '\u2014' not in text
c=json.loads((D/'census-comparison.json').read_text());assert c['head2']['pass'] and len(c['head2']['new'])==8 and c['head2']['mutantRestores384'];assert any(s['where']=='src/compiler/sys.ts:1519:36' for s in c['head2']['new'])
print(json.dumps({'status':'pass','declarationMembers':2,'sourceFiles':746,'unexpectedSites':0,'evidenceArchiveVerified':True},indent=2))
