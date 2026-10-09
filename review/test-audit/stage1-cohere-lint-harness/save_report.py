import json,pathlib,subprocess,re
p=pathlib.Path('/workspace/adamic/review/test-audit/stage1-cohere-lint-harness')
rows=json.loads((p/'rows.json').read_text());matrix=json.loads((p/'matrix.json').read_text());isol=json.loads((p/'isolation-results.json').read_text())
for m in matrix:
 if m['id'] in ['M1','M2','M3']:m['additional_commands']=[x['command'] for x in isol if x['id']==m['id']]
(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
header='''u109 started at origin/main 16f436a16a8e3b9cd2343f449eb1b0b4577c135c; all 31 requested names exist, grouped into 15 rows.
The clean whole package cooked at 90.186 seconds; the restored bounded run passed in 27.645 seconds.
Verdicts: 9 setup-check, 2 witness, 2 limited untrue, 1 subsumed, 1 cannot-judge; no unique production kills proved.
Three production mutants, nine construction breaks, two weakened checks, and twelve separate probes were recorded; four mutants survived.
Eight rows passed their own empty-entry probe; nproc=5; evidence is on test-audit/stage1-cohere-lint-harness under this directory.
'''
text=header+'\n```json\n'+json.dumps(rows,indent=2)+'\n```\n\nMutants, with every site against the starting commit:\n\n| ID | Kind | Origin file:line | Change | Failed grouped rows |\n| --- | --- | --- | --- | --- |\n'
for m in matrix:
 if m['kind']=='probe':continue
 text+='| '+m['id']+' | '+m['kind']+' | '+m['file']+':'+str(m['line'])+' | '+m['change']+' | '+', '.join(m['failed_rows'])+' |\n'
text+='''
Survivors:

- M1: Node directly evaluating written("~") prints `~` before and `\\u007e` after. See M1-before.log and M1-after.log. This is changed behavior unguarded by the bounded matrix; outside rows are unknown.
- H6: both product tests pass without invoking Go build. The reference construction has oracle executables; both H6 products contain only overlay.json and lack oracle. See artifact-witnesses.json.
- H8: the lowered product row passes with empty.c present and program.c absent. See artifact-witnesses.json. This is a construction gap, not a production survivor.
- H9: the native product row passes after flipping sanitize=true to false. `nm` exits zero for both reference and mutant binaries; __asan_init is present in M1-nm.log and absent in H9-nm.log. The reference uses the original sanitizer construction, although its independent ASCII-bound mutant was active. This proves the sanitizer option changed, without claiming byte-for-byte clean-binary identity.

There are no equivalent-candidate survivors. Probes do not count as mutants, kills for worthiness, uniqueness, or subsumption. H and W mutations apply only to construction or witnessed comparisons. Neither Go cohere nor Go TypeScript oracle source was mutated.

Scope and problems with the brief:

- The brief says 15 rows and lists 31 functions. Sixteen JSX tree wrappers share one checker. The two Go-oracle product wrappers also share jsxGoOracle with different input recipes. Those groupings produce 15 rows. All names remain present. The other product recipes differ, so they remain separate.
- Names have stale implications. TestEmittedJavaScriptMismatch is now a shard-census check; TestSuggestionAlongsideAutomaticFix is now an AST-census check. Their runtime checks moved to numbered leaves outside this named scope. TestCompleteSuggestionSerialization is now a union and planted-disagreement witness, after preparation. The delegate files are recorded in go-helper-inventory.txt.
- The union test asserts declaration and exact coverage facts absent from runtime shards. It remains its own construction row under the explicit distinct-assertions exception to family grouping.
- The whole package contains thousands of compiler agreement leaves and cannot fit the 90-second budget. Its first run stopped in the serial opted-in benchmark, before releasing paused rows. A narrowed run also cooked during build preparation. Those are timeouts, not assertion-red baselines. Split compilation made both remaining preparation consumers pass, and all bounded rows then passed alone and together.
- The benchmark still cooked alone with warm dependencies and ADAMIC_NATIVE_SPLIT=1. Its median and quality verdict are unknown. No second or third timing was attempted after it cooked. Its source compares full finding output before checking count equality, but no completed run supports an empirical oracle-strength verdict.
- Product wrappers can return successfully with no product at all. P2/P3/P4/P5/P7/P8/P9 show empty-entry acceptance. Construction-error catches still justify setup-check where demonstrated; empty acceptance is a separate vacuity finding. The lowered and native product untrue verdicts rest on one valid construction break each, not an exhaustive search.
- Isolation is mixed: its purpose is cold preparation, but its child performs an actual parser agreement check. Its observed production failures are included in the matrix and remove uniqueness from the JSX family. Its setup-check verdict separately rests on H4. The family is subsumed only on M2 and M3; this is a two-mutant hint, not a deletion recommendation. The family is faster.
- The production matrix includes the named parser consumers and isolation. It cannot establish package or repository uniqueness. Other lint-port consumers also import written and parser functions; their results are unknown and require central replay. Construction and witness runs are bounded to the check they exercise, so their one-row failures do not establish package uniqueness either.
- Exact dynamic function reach was not established. Conservative TypeScript declaration and Go delegate inventories are supplied, including functions that may not execute on these fixtures. This does not fulfill an exact all-reached-functions proof; the tested mutation sites and their actual failures are demonstrated.
- Port mutations used the allowed per-mutant rebuild alternative, with only three distinct production mutations and a separate entry probe. Go construction edits were also compiled independently instead of using one switch. This added vet/build overhead; no stale products were reused across mutation caches.
- The first H9 prototype did not compile because its changed input left data unused. It was rejected before any matrix run and replaced with the sanitizer-option flip. H9-rejected-vet.log and H9-rejected-prototype.txt document it; the prototype is excluded from replay diffs and verdicts. Removing probe bodies also required removing newly unused imports. P3-initial-vet.log records that cost.
- Drop-statement mutations H4 and H5 drop the whole preparation statement, including its closure. W2 drops the whole assertion block. Runtime failure line numbers shift under those edits; matrix.json maps them back to origin/main. Logs preserve the literal runtime output.
- P12 panicked on a nil shard-selection function. It was run alone and rerun alone in P12-recheck.log. No later code in that row is claimed to have executed. Coverage's positive ownership and relocation checks survive P10/P11 before the negative empty-corpus check fails, recorded as vacuous_subcases.
- The package's cold-isolation child has a 600-second timeout, while this audit gives its parent only 90 seconds. It passed alone three times under the audit budget. Its initial parallel baseline took 63.52 seconds, but its alone median was 23.457 seconds, so that first shared-run time is not used for the cost verdict.
- Warm tools did not remove npm preparation. npm ci was run in stage3/api. The scoped Node entry uses oracle/node.mjs and Node's built-in type stripping; no additional scoped node_modules dependency was found. All requested rows were attempted with their applicable opt-in; none skipped in the completed bounded runs. The benchmark was enabled and cooked.

Timing and uncovered work:

Toolchain setup was skipped because env.sh worked. npm ci took 0.475 seconds; registry generation took 0.321 seconds; listing took 6.214 seconds. nproc was 5. Three-alone timing commands totaled 190.928 wall seconds. Family costs use one binary invocation for all members, not a sum of per-member rounded PASS times. seconds in rows.json are medians of the test binary's package line, with ADAMIC_NATIVE_SPLIT=1 and warm persistent products. Cold costs are visible in baseline logs.

Whole baseline: 91.847 command wall seconds and 90.186 binary seconds. First bounded baseline: 91.692 wall seconds, cooked. Split preparation retry: 81.322 wall seconds and 79.628 binary seconds, passed. Benchmark alone: 91.675 wall seconds, cooked. Final restored bounded baseline: 29.301 wall seconds and 27.645 package-line seconds, passed.

Initial valid mutation runs totaled 181.523 wall seconds; H9/P1 took 42.866; entry probes P2-P12 took 77.916; the additional isolation replays took 75.156. These overlap neither production timings nor each other. Rejected compile attempts and reading/reporting time are additional. Overall unit work was approximately 25 minutes, within the stage1-port allowance rather than the 20-minute ordinary-unit target.

Native rebuild timing for every production mutation and the port probe is recorded below. Native compiler execution is demonstrated by successful product builds before comparisons. Separate caches are /tmp/u109/cache/<id>; isolation children use their own fresh TempDir caches. The complete commands are in run-results-first.json, run-results.json, isolation-results.json, probe-results.json, timings.json, and matrix.json.
'''
for mid in ['M1','M2','M3','P1']:
 for line in (p/(mid+'.log')).read_text().splitlines():
  try:r=json.loads(line)
  except:continue
  output=r.get('Output','')
  if re.search(r'build jsx-tree-(lowered|native) .* miss ',output):text+='\n- '+mid+': '+output.strip()+'\n'
text+='''
Uncovered: completed benchmark execution and three-run median; other package rows and repo-wide replay; an exact dynamic TypeScript reach proof; stronger semantic construction mutants for product-only rows. No tests or production fixes were committed. Only audit evidence is intended for push. No pull request or main push was opened.
'''
# re imported here for the build-line extraction
(p/'REPORT.md').write_text(text)
