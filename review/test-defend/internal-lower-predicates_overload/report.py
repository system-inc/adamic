from pathlib import Path
import json,gzip,shutil,subprocess,hashlib
s=Path('/tmp/def-pred');d=Path('review/test-defend/internal-lower-predicates_overload');d.mkdir(parents=True,exist_ok=True)
r=json.loads((s/'D1-result.json').read_text());base=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip()
row=dict(test='TestPredicateBodiesAreProven',package='internal/lower',prior_verdict='subsumed',prior_subsumed_by=['TestUnprovenPredicateReturnsAreRefused'],subsumed_by=[],defense='defended',unique_mutant=r['mutant']+' '+r['file_line'],attempts=[{k:r[k] for k in ['mutant','file_line','change','rows_failed']}],evidence=r['command']+' > /tmp/def-pred/D1.log 2>&1; '+r['errors'][0],uniqueness_scope='Whole-package completed run: 274 executed top-level tests, 2 opt-in rows skipped; one unsupported subcase skipped.',rows_passed_evidence='D1-result.json: rows_passed lists all 273 passing top-level tests.')
(d/'rows.json').write_text(json.dumps([row],indent=2)+'\n')
for p in s.iterdir():
 if not p.is_file():continue
 if p.suffix=='.log':
  with gzip.open(d/(p.name+'.gz'),'wb') as f:f.write(p.read_bytes())
 else:shutil.copy2(p,d/p.name)
(d/'REPORT.md').write_text('''TestPredicateBodiesAreProven defended by one unique production constant mutant.
Whole-package replay completed; 273 other executed top-level tests passed.
Standalone diff, complete logs, coverage differences and passing rows are saved here.

Start: '''+base+'''. Audit start: 7709c91213f476eba7e3dbfbf4ee65e6988038cb. Requested test and subsumer still exist in internal/lower/predicates_test.go. The current list has 37 added tests relative to the audit, listed in new-tests.json; every one was included in the whole-package baseline and mutant run. Two top-level rows skipped, leaving 274 executed rows.

CODE UNDER TEST: Adamic's Go predicate admission and body proof in internal/lower, reached through lowerSource's Lower invocation. In particular predicateFlowProof.literal recognizes a global undefined identifier as an independently trusted narrowing operand. TypeScript's checker, tests, fixture inputs and expected diagnostics were not mutated.
ORACLE: self-written admission expectations for eleven positive predicate/assertion bodies. The target checks only err == nil from lowerSource. Its subsumer asserts handwritten refusal type, reason and rewrite for 26 invalid bodies. No Node or independent external authority computes either expectation.

Coverage: both rows ran individually with go test -json -count=1 -timeout 90s -coverpkg=./internal/lower -coverprofile, with stdout/stderr directed to the saved .cover.log.gz files. There are 860 covered blocks exclusive to the target versus its named subsumer, listed in exclusive.json. One useful exclusive block is predicates.go:311. The target includes both a function declaration and an arrow predicate with x !== undefined; the subsumer includes neither input. The target also accepts reversed literal comparisons and other valid bodies, but no additional mutation was needed after the first unique kill.

D1 is a menu constant change at predicates.go:311: node.Text() == "undefined" becomes node.Text() == "null", leaving the checker symbol-identity test intact. This breaks trusted recognition of undefined without changing the negative admission oracle. The complete package run fails only TestPredicateBodiesAreProven. Both undefined-body inputs report predicates_test.go:67, return expression is not a trusted check on x. TestUnprovenPredicateReturnsAreRefused passes. D1-result.json lists every one of the 273 other passing top-level tests and the exact command. No panic abort, timeout or incomplete column occurred. Every diff applies to the starting origin/main commit and passes go vet ./internal/lower/.

Limits, ambiguities and costs:
- /tmp is an 8.8 GB filesystem, so 15 GB free cannot be achieved. Initial df showed 5.9 GB free there and 17 GB free in /workspace. Removing the completed earlier unit's /tmp/def-args scratch and per-mutant caches increased /tmp free space to 6.0 GB. Nothing under the repository or tools was deleted. No ENOSPC baseline failure occurred.
- Whole-package uniqueness is observed over executed tests. TestOriginalCycleLedger and TestOptionalWideningCensus skipped because they require project inventory configuration; TestMixedUnionContractGraph has an unsupported views-v3 array subcase. Their potential kills are unknown. Logs preserve all skip reasons. This is not unrestricted uniqueness across optional project inventories or the repository.
- The audit combined a bounded panic-aborted column with complete columns. This defense's single mutation column completed the entire current package, including added tests.
- The fixed audit menu tested proof unsoundness broadly. This aimed defense tests loss of valid undefined-check admission, a different semantic direction from rejecting invalid bodies.
- The target's name mentions proof, but its assertions inspect only successful admission, not returned IR or proof metadata. The audit recorded that its empty-Lower probe passes. That probe was not rerun here and is not counted as a mutant. This defense supports keeping the row for valid-body acceptance, not treating it as independent proof of emitted behavior.
- No cost/performance threshold is promised by this row. Twins and cost-specific mutation requirements do not apply.
- Warm tools worked; setup was skipped, nproc 5. npm ci in stage3/api completed before the baseline. Test output always went to files.
- A first attempt to save coverage differences used the read-only sandbox and failed without changing a file; it was rerun with the authorized write permission.
- Go command time includes compilation and package execution. D1 command wall time was 80.908 seconds; its test binary reported 65.453 seconds. The clean baseline binary reported 78.054 seconds. Coverage runs overlapped the baseline, so that baseline timing is not a standalone performance measurement. No exclusive compiler-build time was isolated.
- No other package suite, additional mutants, new empty-answer probe, opt-in project inventory setup or repository-wide replay was run. The successful first mutation satisfies the up-to-three rule without additional attempts.

Restoration: production source was restored by the runner's finally block. Standalone git apply --check, restored go vet and the clean target-plus-subsumer rerun passed. Complete evidence is pushed on test-defend/internal-lower-predicates_overload; no test edit, main push or PR.
''')
(d/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.name+'\n' for p in sorted(d.iterdir()) if p.is_file() and p.name!='SHA256SUMS'))
