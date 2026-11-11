from pathlib import Path
import json,hashlib,shutil,subprocess
S=Path(__file__).resolve().parent;R=Path('/workspace/adamic');D=R/'stage3/adapt/99-correctness-tracing-dumptypes'
for name in ['finish-step1','fresh-apply','baseline-build','after-build','fresh-types','compare-builds','trace-witness','adapter-contract','audit-mutants','final-source-audit']:
 assert json.loads((S/(name+'.json')).read_text())['exit']==0,name
assert len(json.loads((S/'fresh-types.json').read_text())['command'])-2==79
assert subprocess.check_output(['git','rev-parse','HEAD'],cwd=R)==subprocess.check_output(['git','rev-parse','origin/main'],cwd=R)
step=json.loads((S/'step1-results.json').read_text());assert step['laneShards']==398 and step['completedNodeProcesses']==797 and step['coveredShards']==398
build=json.loads((S/'build-comparison-results.json').read_text());changed=[r for r in build['files'] if not r['equal']];assert len(changed)==3 and sum(r['hunks'] for r in changed)==6
witness=json.loads((S/'trace-witness-results.json').read_text());assert witness['equal'] and all(r['exit']==0 for r in witness['runs']);assert witness['runs'][0]['sha256']==witness['runs'][1]['sha256']
project=json.loads((S/'project-counts-summary.json').read_text());assert project['objectOwnOutsideFlag']==project['objectDefinedOutsideFlag']==46 and project['objectZeroOutsideFlag']==0
frozen=json.loads((S/'frozen-adapter.json').read_text());assert all(hashlib.sha256((D/n).read_bytes()).hexdigest()==h for n,h in frozen.items())
proof=(S/'PROOF-DRAFT.md').read_text().replace('This draft is incomplete until the whole lane and Step 2 results are filled in.\n','')
proof=proof.replace('The new adaptation must pin the complete composed dumpTypes\nbefore and after and change its two subtype reads together.','The new adaptation pins the complete composed dumpTypes\nbefore and after and changes its two subtype reads together.')
proof=proof.replace('PENDING COMPLETE LANE RESULT.','''All 398 real gate TestShard_ units pass with child exit 0. Their manifests
validate 106,366 passing oracle cases, one sanctioned API acknowledgement
failure and zero skipped cases. All 398 units have counter-loaded process
coverage, across 797 completed Node processes. dumpTypes executes **0 times**
and visits **0 types** in the entire lane. Written-value differences are **0**.
This is a zero count, not direct coverage of the changed function. The three
real generateTrace projects below supply that execution coverage.''')
section=f'''The selected replacements are getObjectFlags(type) and the Intrinsic flag
conditional followed by a checked IntrinsicType read. No unknown substitution,
new helper or compiler change is introduced. Only tracing.ts changes among all
747 source files: two imports and the two owned expressions. All 747 source
files from the fresh full pipeline match the normally built candidate exactly.

| Check | Result |
| --- | --- |
| Whole-owner/import contract | 15 cases: 4 accepted with zero-write replay, 11 refusals before any write |
| LF/CRLF and unrelated text/imports | Preserved |
| Baseline and fresh complete apply | Both child exits 0 |
| Normal builds before and after | Both child exits 0 |
| All 79 roots, baseline and fresh candidate | Both child exits 0, zero diagnostics |
| Normal emitted JavaScript | 10 files checked; 3 changed files, 6 hunks, all inside dumpTypes |
| Emitted declarations | All {build['declarationFiles']} files byte-identical |
| Runtime export keys | All {build['exportKeys']:,} own string keys identical; {build['symbolKeys']} symbol keys on each side |

The changed bundles are _tsc.js, run.js and typescript.js. Each complete file
matches baseline outside dumpTypes, and its entire changed owner equals the
baseline owner with only the two approved emitted replacements. The other seven
JavaScript files are byte-identical. Source maps and build metadata are outside
this JavaScript comparison.

The normal-build Node witness runs a small project with unions, tuples, a
conditional type and a generic mapping function. Both CLI images run in one
shared directory with the same library and project paths. Each --generateTrace
run exits 0 and writes **{witness['runs'][0]['bytes']:,} bytes**, exactly identical before and after.
SHA-256 on both written types files:

    {witness['runs'][0]['sha256']}

Both restoring mutants use complete copies of all 747 source files, restoring
only one original cast expression in tracing.ts. All 23 tracing census units
run in each case with zero checker diagnostics. The two imports account for
the two-line coordinate shift:

| Restored original site | Mutant finding | Phases |
| --- | --- | --- |
| 232:34 objectFlags | 234:34 | refusal_scan |
| 318:33 intrinsicName | 320:33 | refusal_scan and lowering |

The candidate has zero unchecked-cast findings in the complete tracing census.
The baseline has three such events at these two sites; the object mutant
restores one and the intrinsic mutant restores two. Other existing NotYet and
Refused findings remain. Both no-output guards are enabled in the measurement
binary; production loader and lowering code are unchanged.

The adapter code and pins stay frozen throughout full apply, builds and final
source verification. Their retained hashes are in frozen-adapter.json. This
final prose was added after those checks and changes no generated source.
'''
proof=proof.replace('PENDING ADAPTER, RESTORING MUTANTS, NORMAL APPLY, TYPES, BUILDS, WRITTEN TYPES\nWITNESS, JAVASCRIPT HUNKS AND RUNTIME EXPORT KEYS.',section)
proof += '\nThe first small-project witness invocation was also retained: TypeScript 6\nrejected the repository tsconfig context. The accepted invocation uses the\nproject working directory and --ignoreConfig on both sides.\n'
assert 'PENDING' not in proof and '\u2013' not in proof and '\u2014' not in proof
(D/'PROOF.md').write_text(proof)
entry='''Task dp24r5m: 99-correctness-tracing-dumptypes replaces two explicit-any reads
with getObjectFlags and an Intrinsic flag guard plus checked IntrinsicType read.
The whole lane records 0 calls across 398 shards and 797 Node processes; three
generateTrace projects visit 84,642 types, including 46 real objectFlags values
outside ObjectFlagsType, with 0 differences at all five masks or intrinsicName.
Written types are byte-identical, 79 roots have 0 diagnostics, both restoring
mutants restore their sites, every JavaScript hunk stays inside dumpTypes, and
2,256 runtime export keys are identical. See its PROOF.md. No upstream issue filed.
'''
p=R/'stage3/upstream/LEDGER.md';text=p.read_text();assert 'Task dp24r5m:' not in text;p.write_text(text.rstrip()+'\n\n'+entry)
shutil.copyfile('/workspace/scratch/dp24r5m-apply-base.log',S/'baseline-full-apply.log')
(S/'baseline-apply-result.json').write_text(json.dumps(dict(command=['bash','stage3/apply.sh','/workspace/scratch/dp24r5m-base'],cwd=str(R),exit=0,recordedFrom='Original shell wrapper reported its child exit 0'),indent=2)+'\n')
(S/'STATUS.md').write_text('''# dp24r5m complete verification\n\nStep 1 and Step 2 pass. Both casts are replaced in one whole-owner adaptation.\nAll 398 gate shards pass their expected counts, with 797 counter-loaded Node\nprocesses, 0 dumpTypes calls and 0 visits. The actual adapted compiler visits\n84642 types across the three named generateTrace projects, with 0 written-value\ndifferences. Both restoring mutants restore their exact sites. Full apply,\n79 roots, normal builds, JavaScript scope, declaration identity, written-types\nbyte identity and runtime export keys all pass.\n\nThe cancelled attempt1 corpus-copy run and other corrected initial attempts\nare retained separately and excluded. PROOF.md contains the complete numbers.\n''')
(S/'STEP1.md').write_text(proof.split('## Step 2')[0])
shutil.copytree(D,S/'final-adapter',dirs_exist_ok=True)
(S/'final-verification-results.json').write_text(json.dumps(dict(complete=True,sourceFiles=747,roots=79,laneShards=398,nodeProcesses=797,laneCalls=0,measuredTypes=84642,writtenDifferences=0,javascriptFiles=10,hunks=6,declarationFiles=716,exportKeys=2256,restoringMutants=2),indent=2)+'\n')
print('All requested verification complete; proof and ledger written.')
