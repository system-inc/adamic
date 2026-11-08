import sys
sys.dont_write_bytecode=True
from run import *
import difflib,collections,hashlib
D=ROOT/'stage3/adapt/99-correctness-core-sys-timer-slot';out=D/'generator-fix';out.mkdir(exist_ok=True)
assert run('head1-api',['git','show','HEAD:stage3/lane/sanctioned-api.json'])==0
committed=(S/'head1-api.log').read_text();old=(S/'head1-regenerated.json').read_text();new=(ROOT/'stage3/lane/sanctioned-api.json').read_text();a=json.loads(old);b=json.loads(new);assert set(a)==set(b)
changes=[k for k in a if a[k]!=b[k]];assert set(changes)=={'normalized_changes','owners','proofs','diff_lines'}
added={}
for field in ['normalized_changes','owners','proofs']:
 aa=collections.Counter(json.dumps(v,sort_keys=True) for v in a[field]);bb=collections.Counter(json.dumps(v,sort_keys=True) for v in b[field]);assert not aa-bb
 added[field]=[json.loads(v) for v,n in (bb-aa).items() for _ in range(n)]
assert len(added['normalized_changes'])==2 and all('System' in v['declaration'] and any(x in v['declaration'] for x in ['setTimeout','clearTimeout']) for v in added['normalized_changes'])
assert len(added['owners'])==2 and all(v['path'][0]=='System' for v in added['owners'])
assert len(added['proofs'])==1 and added['proofs'][0]['file']=='stage3/adapt/99-correctness-core-sys-timer-slot/api-sanctions.json'
for prefix,field in [('-','before'),('+','after')]:
 aa=collections.Counter(line for line in a['diff_lines'] if line.startswith(prefix) and not line.startswith(prefix*3));bb=collections.Counter(line for line in b['diff_lines'] if line.startswith(prefix) and not line.startswith(prefix*3));assert not aa-bb
 wanted=collections.Counter(prefix+'        '+r[field]+'\n' for r in json.loads((D/'api-sanctions.json').read_text()));assert bb-aa==wanted,(bb-aa,wanted)
diff=''.join(difflib.unified_diff(committed.splitlines(True),new.splitlines(True),fromfile='head1/stage3/lane/sanctioned-api.json',tofile='head2/stage3/lane/sanctioned-api.json'))
(out/'sanctioned-api.diff.txt').write_text(diff)
result={'status':'pass','head1':'5ba70a9486a9d297f42a6a29fd9b2388caaf093d','keys':list(b),'beforeLines':len(committed.splitlines()),'afterLines':len(new.splitlines()),'changedFields':changes,'unchangedExistingRowsAgainstRegeneratedHead1':True,'preexistingHead1ProofDrift':json.loads((S/'head1-generator-drift.json').read_text()),'added':added,'countsUnchanged':True,'baselineDiffsUnchanged':True,'diffAddsOnlyTwoTimerBeforeAfterLines':True,'sha256':hashlib.sha256(new.encode()).hexdigest()}
(out/'delta.json').write_text(json.dumps(result,indent=2)+'\n')
p=D/'README.md';text=p.read_text();text+='\n## House API generator correction\n\nThe earlier 915a67cf receipt used the raw cjs helper output. The corrected head\nuses the documented writer, which removes original/expected and retains the\npretty house schema with baseline_diffs and diff_lines:\n\n```sh\nNODE_PATH="$PWD/stage3/api/node_modules" \\\n  python3 stage3/lane/generate-api.py "$ADAMIC_TYPESCRIPT_SOURCE" \\\n  > /tmp/lane-generate.log 2>&1\n```\n\nThe following complete diff is against committed Head 1. In addition to the\ntwo timer members and their owner/proof rows, it includes two pre-existing\nproof-hash refreshes. An isolated regeneration on Head 1 reproduces exactly\nthose two hash refreshes, with no timer adaptation present. Their inclusion\nrequires resolution of the requested timer-only constraint. Against regenerated\nHead 1, existing normalized rows, owners and proofs are unchanged; additions\nbelong only to System timers. Counts and baseline_diffs stay unchanged. The patch-line delta contains\nonly the two before/after timer declarations. delta.json records these checks.\n\n```diff\n'+diff+'```\n';p.write_text(text)
p=D/'regenerate-head2.py';text=p.read_text().replace("['node','stage3/lane/generate-api.cjs','/workspace/typescript-source']","['python3','stage3/lane/generate-api.py','/workspace/typescript-source']");text=text.replace("shutil.copyfile(S/'head2-generate-api.log',ROOT/'stage3/lane/sanctioned-api.json')", "# generate-api.py writes its default sanctioned-api.json output.");p.write_text(text)
print(json.dumps(result,indent=2))
