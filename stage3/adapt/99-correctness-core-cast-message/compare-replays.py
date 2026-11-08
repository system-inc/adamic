from pathlib import Path
import json,difflib,hashlib,re
s=Path('/workspace/scratch/core-finish');raw=s/'raw'
def load(side):
 rows=json.loads((s/('replay-raw-'+side)/'records.json').read_text())
 for row in rows:row['stdout_stderr']=row['stdout_stderr'].replace('/workspace/adamic','<repo>')
 return {r['adapter']:r for r in rows}
a=load('baseline');b=load('head1');assert set(a)==set(b) and len(a)==97
rawDifferences=[k for k in a if a[k]!=b[k]]
# 49 writes each already-adapted file back unchanged. Preserve its raw hashes,
# prove the exact inherited input difference, then compare its identity delta.
left=a['49-honest-optionals'];right=b['49-honest-optionals'];assert left['exit']==right['exit']==0 and left['stdout_stderr']==right['stdout_stderr'];assert len(left['writes'])==len(right['writes'])==4
pins=json.loads((Path('/workspace/adamic')/'stage3/adapt/99-correctness-core-cast-message/owners.json').read_text());old=(s/'baseline/src/compiler/core.ts').read_bytes();new=(s/'head1/src/compiler/core.ts').read_bytes();assert old.replace(pins['removedPhrase'].encode(),b'')==new
for l,r in zip(left['writes'],right['writes']):
 assert l['file']==r['file'];assert l['before']==l['after'] and r['before']==r['after']
 if l!=r:
  assert l['file']=='src/compiler/core.ts';assert l['before']==hashlib.sha256(old).hexdigest();assert r['before']==hashlib.sha256(new).hexdigest()
(raw/'head1-replay-noop-write-proof.json').write_text(json.dumps({'adapter':'49-honest-optionals','sameExitOutputPathsAndWriteCount':True,'bothWriteIdentityBytes':True,'differentRawInputHashes':{'baseline':hashlib.sha256(old).hexdigest(),'head1':hashlib.sha256(new).hexdigest()},'onlyInheritedPayloadDifference':'exact cast function-name phrase deletion','noHashOrRawJournalRewritten':True},indent=2)+'\n')
diff=[k for k in rawDifferences if k!='49-honest-optionals'];assert diff==['99-correctness-core-cast-message'],diff;assert a[diff[0]]['exit']==b[diff[0]]['exit']==0;assert len(a[diff[0]]['writes'])==1 and not b[diff[0]]['writes'];first={'passed':True,'adapters':97,'differentAdapters':diff,'rawJournalDifferentAdapters':rawDifferences,'identityRewriteWithInheritedInputHashDifference':'49-honest-optionals, independently proved in head1-replay-noop-write-proof.json','identicalOtherAdapterTransformations':96,'beforeEditCalls':1,'afterWriteCalls':0,'normalization':'disposable source-tree and repository roots only; no exits, reasons, writes or data normalized'};(raw/'head1-paired-replay.json').write_text(json.dumps(first,indent=2)+'\n');print('head1 paired replay passed',flush=True)
if (s/'replay-raw-head2/records.json').exists():
 c=load('head2');assert len(c)==98 and set(c)-set(b)=={'99-correctness-core-getfunctionname-move'};amended={'30-indexed-reads','41-explicit-any-remaining','45-regex-captures','99-correctness-core-debug-messages'};different=[k for k in b if b[k]!=c[k]];assert set(different)<=amended,different
 for k in amended:
  assert b[k]['exit']==c[k]['exit'],(k,b[k]['exit'],c[k]['exit']);assert b[k]['writes']==c[k]['writes'],k
  if k=='41-explicit-any-remaining':assert 'missing or duplicate reviewed site: 31' in b[k]['stdout_stderr'] and 'missing or duplicate reviewed site: 31' in c[k]['stdout_stderr'];assert b[k]['stdout_stderr'].count('41-explicit-any-remaining/adapt.cjs:16')==2;assert b[k]['stdout_stderr'].replace('41-explicit-any-remaining/adapt.cjs:16','41-explicit-any-remaining/adapt.cjs:18')==c[k]['stdout_stderr']
  else:assert b[k]['stdout_stderr']==c[k]['stdout_stderr'],k
 assert c['99-correctness-core-getfunctionname-move']['exit']==0 and not c['99-correctness-core-getfunctionname-move']['writes'];result={'passed':True,'adapters':98,'differentExistingAdapters':different,'allowedRecognitionAmendments':sorted(amended),'allExistingExitsAndWritesIdentical':True,'unamendedAdapterOutcomesIdentical':True,'moveReplayWrites':0,'amended41SameRule31Refusal':'only amended source stack line numbers differ','normalization':first['normalization']};(raw/'head2-paired-replay.json').write_text(json.dumps(result,indent=2)+'\n');(raw/'head2-replay-difference.json').write_text(json.dumps({k:{'before':b[k],'after':c[k]} for k in different},indent=2)+'\n');print('head2 paired replay passed',flush=True)
