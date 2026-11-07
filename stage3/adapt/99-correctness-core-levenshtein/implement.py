import json,hashlib,shutil,tarfile,io
from pathlib import Path
s=Path('/workspace/scratch/levenshtein');repo=Path('/workspace/adamic');out=repo/'stage3/adapt/99-correctness-core-levenshtein';count=json.loads((s/'COUNT.json').read_text());reconciled=json.loads((s/'reconciled.json').read_text())
assert count['complete'] and reconciled['complete'] and count['differences']==0 and count['calls']>0
assert not (out/'owners.json').exists(),'already implemented'
oldReadme=(out/'README.md').read_text();argument='1. The two allocations'+oldReadme.split('1. The two allocations',1)[1].split('NaN and positive Infinity',1)[0]
# Keep the first stopping report intact under a clearly historical archive.
archive=s/'initial-report.tar.gz';files=sorted(p for p in out.iterdir() if p.is_file());assert len(files)==13
with tarfile.open(archive,'w:gz') as tar:
 for p in files:tar.add(p,arcname=p.name)
 manifest=''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.name+'\n' for p in files).encode();info=tarfile.TarInfo('SHA256SUMS');info.size=len(manifest);tar.addfile(info,io.BytesIO(manifest))
shutil.rmtree(out);out.mkdir();shutil.copyfile(archive,out/archive.name);(out/'initial-report.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+archive.name+'\n')
pins=json.loads((s/'measured-owners.json').read_text());pin={'file':'core.ts','owner':'levenshteinWithMax',**{k:v.replace('\r\n','\n') for k,v in pins.items()}}
for k in ['before','after']:pin[k+'Sha256']=hashlib.sha256(pin[k].encode()).hexdigest()
assert pin['before'].count('s1[i - 1]!')==pin['after'].count('s1[i - 1]!')==1
assert pin['before'].count('s2[j - 1]!')==pin['after'].count('s2[j - 1]!')==1
(out/'owners.json').write_text(json.dumps([pin],indent=2)+'\n');shutil.copyfile(s/'adapter-template.cjs',out/'adapt.cjs')
(s/'adapter-frozen.json').write_text(json.dumps({'adaptSha256':hashlib.sha256((out/'adapt.cjs').read_bytes()).hexdigest(),'ownersSha256':hashlib.sha256((out/'owners.json').read_bytes()).hexdigest(),'beforeOwnerSha256':pin['beforeSha256'],'afterOwnerSha256':pin['afterSha256'],'countCompleteBeforeImplementation':True,'calls':count['calls'],'differences':0},indent=2)+'\n')
for name in ['instrument.cjs','preload.cjs','measure-prepare.py','count.py','control.py','reconcile.py','witness.cjs','replay.py','compare.cjs','contract.cjs','prepare-census.sh','run-census.py','census.py','checked-sites.cjs','implement.py','adapter-template.cjs']:
 shutil.copyfile(s/name,out/name)
for name in ['COUNT.json','reconciled.json','checked-read-sites.json']:shutil.copyfile(s/name,out/name)
for archive in sorted(s.glob('count-*.tar.gz')):
 shutil.copyfile(archive,out/archive.name);(out/(archive.name.removesuffix('.tar.gz')+'.sha256')).write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+archive.name+'\n')
rows=json.loads((s/'checked-read-sites.json').read_text());table='| Site | Read index | Why in range |\n| --- | --- | --- |\n'
for r in rows:
 proof='0 <= j - 1 <= n - 1, because 1 <= minJ <= j <= maxJ <= n' if r['index']=='j - 1' else '1 <= j <= n, because minJ <= j <= maxJ <= n and minJ >= 1' if r['index']=='j' else 'n = s2.length >= 0, and the row has length n + 1'
 table+='| '+r['where']+' '+r['site']+' (`'+r['expression']+'`) | `'+r['index']+'` | '+proof+' |\n'
counts='| Calls / class flag | Count |\n| --- | ---: |\n| Total calls | '+str(count['calls'])+' |\n| Return differences | 0 |\n'+''.join('| '+k+' | '+str(v)+' |\n' for k,v in count['classes'].items())
(out/'README.md').write_text('''# levenshteinWithMax: filled number rows and checked indexed reads

JavaScript-changing established-shape adaptation. TypeScript 6.0.3,
050880ce59e30b356b686bd3144efe24f875ebc8. Base main f0271c6a7; the existing
816190260 stopping report is preserved in initial-report.tar.gz. The user then
explicitly authorized the six non-null assertions and continuation on this branch.
Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts.

Both row allocations become new Array<number>(s2.length + 1).fill(0), and six
indexed reads gain !. No other owner text changes. The complete exact owner
texts and hashes are pinned in owners.json. Missing, duplicate, nested,
drifted and half-adapted owners refuse before any write. Replay writes nothing
and preserves bytes and mtime. LF/CRLF and all foreign source text survive.

30-indexed-reads already supplies s1[i - 1]! and s2[j - 1]! (sites.json entries
at lines 966 and 978). They are present in the before-pin and remain untouched.
The six new assertions use 30's rule class: checked indexed reads, erased in
emitted JavaScript and checked loudly by the native runtime. 30 could not add
these while the row elements were any. The user confirmed that its two
expression/occurrence pins do not reserve the complete owner.

## Each checked read is in range

Let n = s2.length. Both rows are filled number arrays of length n + 1 and
retain that representation across swaps. Whenever the middle loop executes,
minJ >= 1 and maxJ <= n; j increases from minJ through maxJ. The j - 1
predecessor is therefore nonnegative. The final index n is always in the row.

'''+table+'''
The first five assertions are inside that middle loop. The last is after all
rows (including the empty-s1 case). Filling makes every in-range slot present;
the per-row assignment proof below additionally shows that a new zero cannot
influence the returned distance. These assertions resolve the seven checker
diagnostics in the original initializer-only proposal. They add no emitted
JavaScript beyond the two fill(0) calls.

## Why the initial zeros cannot affect a returned value

'''+argument+'''
## Measurement before implementation

The complete count finished and was independently reconciled before adapt.cjs
was replaced with an implementation. Scratch instrumentation runs the exact
original and accepted owner bodies on every invocation, compares Object.is
return equality, throws on a difference, and returns the original result.
No compiler or lane-runner source is changed. In the accepted union, all 398
unchanged lane shards run once in 20 bounded pieces with their normal identity, baseline, API and
count guards. Raw arguments and both returns are retained. Input source and
built-file manifests remain frozen. Piece archives have SHA256SUMS and outer
SHA-256 files. The initial duplicate-node_modules setup failure recorded no
calls and is retained separately, excluded from this accepted union. A second
partial count was stopped after observed memory contention (2,611 calls, zero
differences), and the contended full control had three test timeouts. Both are
retained as excluded attempts. The accepted full control and count run in
isolation with all normal guards; excluded calls do not inflate the totals.

'''+counts+'''
Class flags overlap and are not a partition. The separate unchanged full
stage3/lane/run.sh control runs from the pinned main worktree, reusing its
verified baseline product. Reconciliation checks identical test identities,
counts and raw baseline diff for that full control and the 398-shard union.
Node witnesses cover every requested edge class and ordinary finite distance;
they are excluded from the real-call count.

Full apply, full paired replay, 79-root types, JavaScript, public declarations,
complete core.ts census and initializer-only mutant results are in PROOF.md.
The paired replay requires adapter 30's outcomes to remain identical. Its
known addRange decline refusal on a fully adapted baseline is preexisting and
must match on both sides; successful replay is separately checked at its own
post-apply checkpoint. A refusal caused by these edits is a stop.
''')
(out/'PROOF.md').write_text('# Verification checkpoint\n\nThe complete differential count passes. The amended full apply and subsequent proofs are in progress.\n')
ledger=repo/'stage3/upstream/LEDGER.md'
with ledger.open('a') as f:f.write('\n- levenshteinWithMax filled number rows and six checked indexed reads: Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts. Whole-lane differential count: '+str(count['calls'])+' calls, 0 counterexamples. Six assertions use adaptation 30\'s checked-indexed-read rule class, authorized by the user. See stage3/adapt/99-correctness-core-levenshtein/COUNT.json.\n')
print('IMPLEMENTED after reconciled zero-difference count:',count['calls'],flush=True)
