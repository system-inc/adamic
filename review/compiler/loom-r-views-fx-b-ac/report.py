from pathlib import Path
import gzip,hashlib,json,subprocess
root=Path.cwd();out=root/'review/compiler/loom-r-views-fx-b-ac'
statuses=json.loads((out/'statuses.json').read_text())
checks={n:json.loads((out/(n+'.json')).read_text()) for n in ['focused','fixtures','readers','vet']}
assert all(x['exit']==0 for x in checks.values())
assert len([t for t in checks['fixtures']['tests'] if '/' in t['Test']])==14
assert all(x['status']=='checked' and x['header_added'] is None for x in statuses)
assert len(statuses)==17
mutants=json.loads((out/'header-mutants.json').read_text());assert len(mutants)==2 and all(x['caught'] for x in mutants)
assert json.loads((out/'final.json').read_text())['exits']['a-check']==0
base=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip()
assert base=='dbdf977f5b8b0979d414ebe6f20e7aadd312e309'
extra=subprocess.check_output(['git','diff','--name-only','eebc4d4ac6421925ca313bfc5eec332e17d8603b'],text=True);assert not extra,extra
leaves=[dict(check=n,**t) for n,x in checks.items() for t in x['tests']]
(out/'leaf-seconds.json').write_text(json.dumps(leaves,indent=2)+'\n')
proof={'gate_commit':'fbac28c62493f27a788edc02a18bc8edb68de5da','gate_file':'cloud/fast-gate/run.py','gate_sha256':hashlib.sha256((out/'pinned-gate.py').read_bytes()).hexdigest(),'compiler_sha256':hashlib.sha256(Path('/workspace/adamic-binaries/adamic-acheck').read_bytes()).hexdigest(),'compiler_source':'eebc4d4ac6421925ca313bfc5eec332e17d8603b unchanged','base':base,'statuses':{'checked':17,'NotYet':0,'refused':0,'type error':0},'headers_added':[]}
assert (out/'pinned-gate.py').read_bytes()==subprocess.check_output(['git','show',proof['gate_commit']+':'+proof['gate_file']])
(out/'provenance.json').write_text(json.dumps(proof,indent=2)+'\n')
text=f'''Audited all 17 changed Adamic files, including review witnesses: all checked, so no a-check headers are required.
Commit: one delivery commit on compiler/loom-r-views-fx-b-ac, SHA in final response; base {base}; carries eebc4d4ac6421925ca313bfc5eec332e17d8603b unchanged.
Commands and outputs: real compiler a-check scan PASS, focused tests PASS, 14 oracle fixtures PASS, 52 reader packages PASS, reader guard PASS, vet PASS.
Mutants: removed-header and wrong-reason mutants both rejected by the unchanged pinned Gate.aCheck predicate on an existing refusal fixture.
Not covered: full-tree a-check, full package/repository gates, Loom submission, or rerunning the six compiler mutants already proved in the carried commit.

Fetched origin, then created the requested branch at exactly dbdf977f5b8b0979d414ebe6f20e7aadd312e309 and ran `git cherry-pick --no-commit eebc4d4ac6421925ca313bfc5eec332e17d8603b`. No conflicts. Origin/main had moved to 99c369e0, but the explicitly requested pinned base was retained. The original compiler implementation, test split, fixtures, runtime files and counts are unchanged. The only additions beyond the carried change are this audit's evidence under review/compiler/loom-r-views-fx-b-ac/.

Read CLAUDE.md, docs/front-door.md and the complete stage3/a-check-headers/README.md. Its exact historical reproduction harness uses Gate.aCheck from fbac28c62493f27a788edc02a18bc8edb68de5da:cloud/fast-gate/run.py. Fetched that object, retained its byte-identical implementation as pinned-gate.py, and invoked its actual method without changing the predicate. provenance.json records the gate and compiler binary hashes. scan.py adds hard timeouts to the documented harness, records real stderr and exit codes, and preserves every body hash. It does not run other gate phases.

The method builds this branch's own compiler with `go build -o /workspace/adamic-binaries/adamic-acheck ./cmd/adamic`, then runs `adamic-acheck c <path>` on every added or modified .a selected by `git diff --name-only --diff-filter=AM dbdf977f5b -- '*.a'`. All 17 invocations exited 0 and emitted C. No result was a NotYet stop, refusal or checker error. Thus even the review witnesses need no header on this exact compiler. A checked compilation is an observation about this front end; it is not a claim that an unchecked source program's runtime behavior is valid.

| Status | Files |
|---|---:|
| checked (compiler exit 0) | 17 |
| NotYet | 0 |
| refused | 0 |
| type error | 0 |

Per-file status:

| Path | Status | Header added |
|---|---|---|
'''
for x in statuses:text+=f"| {x['path']} | {x['status']} | none |\n"
text+='''
Headers added: none. All 17 complete source files, not merely their bodies, are byte-identical to eebc4d4ac6421925ca313bfc5eec332e17d8603b. statuses.json records SHA-256 before/after values. No diagnostic line moved, so no test, golden, or report expectation required a positional update. The review-witness basename audit over *_test.go and testdata returned no matches; positional-references.json records the exact command and output. Previous review reports and diagnostic captures remain historical evidence for their original runs.

Conservative assumption: because none of the changed files requires a header, do not invent one merely to create a mutant witness. Instead use the existing stage3/fixtures/entries/06_symbol_key.a as a temporary negative control. Read it whole and compile it with the same compiler. The existing first-line header is `// a-check: refused Object enumeration with symbol keys`; the actual diagnostic is `Adamic 0.1 refuses Object enumeration with symbol keys; use own string-keyed data properties`. Its baseline passes the unchanged predicate.

As README.md prescribes, the mutant runs use that captured real compiler diagnostic with the unchanged Gate.aCheck. Removing the header makes its actual refused outcome disagree with expected checked. Replacing its reason with `deliberate wrong reason mutant` preserves the refused classification but fails the required reason match. Both predicate runs returned a-check exit 1, with their specific failures in removed-header.json, wrong-reason.json and header-mutants.json. Each mutation was restored in finally. The entire control file is byte-identical to main and is not changed in the commit. The final real-compiler scan over all 17 owned files passes again, exit 0; final.json records every invocation and first diagnostic. These are two header-predicate catches, not new compiler-behavior mutants.

The initial scan script successfully compiled and classified all 17 programs and passed the gate, then stopped because it tried to select an error case from an empty set. The completed runner selects the existing refusal control instead. scan-final.log and final.json are the completed results; the initial scan.log remains as evidence.

Counts are byte-identical to the carried commit; no regeneration was needed or performed in this pass. TestCountsAreRecordedTable passed. The carried commit's 14 added and six modified count rows remain in the combined diff against main. No runtime C line, IR map, package, compiler consumer declaration or build product key changed in this pass. No cohere code was copied.

All test output went directly to files under this review directory. checks.py imposes 240-second subprocess bounds and each Go test uses -timeout 90s; the shell adds timeout 260. The scan's builds have 120-second limits and each compiler invocation has a 30-second limit, with a 600-second outer bound. No top-level test was added or edited beyond the carried commit. Every measured focused test finished under 60 seconds. Exact commands, exits and each leaf's seconds are in the per-command JSON and leaf-seconds.json. raw-logs.json.gz retains all new raw log output.

| Command record | Result | Wall seconds | Recorded test results |
|---|---|---:|---:|
'''
for n,x in checks.items():text+=f"| {n}.json | PASS | {x['seconds']} | {len(x['tests'])} |\n"
text+='''
Focused selection: TestFX6*, TestScalarUnionView*, TestPredicateBodyProof, TestRuntimeSetsAreEachUsed, TestCountsAreRecordedTable, TestObjectMetadataCacheIsPerEmission and TestViewMixedUnionSelectsNumber. The oracle fixture selector is `^TestNativeAgreesWithNode_[0-9]+$/internal/lower/testdata/(fx6|scalar_union_views)`: exactly 14 actual fixture leaves ran, comparing source Node, emitted JavaScript, release native and ASan/UBSan native, with leak checks for successful programs. Checked failures retain their explicit exit-70 expectations. The prior six compiler-mutant proofs remain unchanged in review/compiler/loom-r-views-fx-b/; this pass reran the two requested header mutants.

The same 52 reverse-reader packages listed in the carried reader-packages.json were compiled with `go test <52 packages> -run '^TestCallTargetReaders$' -count=1 -json -timeout 90s`; the imports are unchanged. Packages with no matching test reported no tests to run. TestCallTargetReaders passed in 3.52 seconds. `go vet ./internal/lower ./internal/native ./internal/javascript ./internal/oracle` passed without diagnostics.

Setup passed: GOPROXY='https://proxy.golang.org|direct'; timeout 600 bash cloud/setup.sh; sourced /workspace/adamic-tools/env.sh. Go 1.27.1, Node 24.19.0, house clang 20.1.8. nproc=5 with cpu.max=400000 100000 (four CPUs), verification GOMAXPROCS=4. Timing lines: Go ready 0.080s; Node 0.091s; submodules 0.093s; markdown ready 0.226s (installed-byte validation 0.011s); clang 0.456s; shared cache 0.955s; go build 17.593s; test binaries deferred 17.690s; build cache warm 17.692s; done 17.724s. The managed environment first needed startup; no setup workaround was needed once it was ready.

No PR, ahra command, Loom submission, main push or force push was made. The new branch is one commit on the requested pinned main base.
'''
(out/'REPORT.md').write_text(text)
logs={f.name:f.read_text(errors='replace') for f in sorted(out.glob('*.log'))}
with gzip.open(out/'raw-logs.json.gz','wt') as f:json.dump(logs,f)
print('report ready; checked 17, NotYet 0, refused 0, type error 0; no headers needed')
