from pathlib import Path
import json,re,collections,hashlib,difflib
out=Path('/workspace/scratch/99u-final31-reference-oracle');tree=Path('/workspace/scratch/99u-final31-main')
log=(out/'tests.log').read_text();counts={name:int((re.findall(r'^\s*(\d+) '+name+r'\b',log,re.M)or['0'])[-1]) for name in ['passing','failing','pending']}
assert counts=={'passing':106366,'failing':1,'pending':0};assert 'Completed runtests-parallel with errors' in log and 'Process exited with code: 1' in log
raw=collections.Counter()
for line in Path('/workspace/scratch/99u-final31-reference-identities.jsonl').read_text().splitlines():
 r=json.loads(line);raw[(r['event'],r['title'])]+=1
assert sum(raw.values())==212734 and all(n%2==0 for n in raw.values())
assert sum(n for (e,t),n in raw.items() if e=='pass')==212732
assert [(t,n) for (e,t),n in raw.items() if e=='fail']==[('unittests:: Public APIs for typescript.d.ts should be acknowledged when they change',2)]
diffs=[]
with (out/'baseline.diff').open('w') as target:
 for p in sorted((tree/'tests/baselines/local').rglob('*')):
  if not p.is_file():continue
  rel=p.relative_to(tree/'tests/baselines/local');deleted=p.name.endswith('.delete');refname=Path(str(rel)[:-7]) if deleted else rel;ref=tree/'tests/baselines/reference'/refname;old=ref.read_bytes() if ref.exists() else b'';new=b'' if deleted else p.read_bytes()
  if old!=new or deleted or not ref.exists():
   diffs.append(str(rel));target.writelines(difflib.unified_diff(old.decode('utf8',errors='backslashreplace').splitlines(True),new.decode('utf8',errors='backslashreplace').splitlines(True),fromfile='reference/'+str(refname),tofile='local/'+str(rel)))
assert diffs==['api/typescript.d.ts'],diffs
recovered={'base':'31d7a244808fa8092643926d21ab9d310a54cce3','tree':str(tree),'counts':counts,'baseline_diffs':diffs,'observedRawIdentityEvents':sum(raw.values()),'allIdentityMultiplicitiesEven':True,'testProcessExitObservedInLog':1,'wrapperReportWasMissingAfterDisconnect':True,'wrapperExitAndWallTimeNotClaimed':True,'recovery':'The unmodified oracle completed its upstream test process before the executor disconnected. Verify the completed log, every raw identity event and every baseline file without inventing missing wrapper timing or exit metadata.','logSha256':hashlib.sha256(log.encode()).hexdigest()}
(out/'recovered-report.json').write_text(json.dumps(recovered,indent=2)+'\n');print(json.dumps(recovered,indent=2))
