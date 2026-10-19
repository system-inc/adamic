import hashlib,json,re,subprocess
from pathlib import Path
p=Path(__file__).parent
before=json.loads((p/'before.json').read_text());after=json.loads((p/'after.json').read_text());headers=json.loads((p/'headers.json').read_text());bodies=json.loads((p/'body-verification.json').read_text());mutants=json.loads((p/'mutants.json').read_text());checks=[json.loads(l) for l in (p/'checks.jsonl').read_text().splitlines()]
assert len(after['rows'])==22 and after['exit']==0 and all(c['exit']==0 for c in checks)
seconds=[]
for f in p.glob('*-seconds.json'):
 if f.name=='test-seconds.json':continue
 for r in json.loads(f.read_text()):seconds.append(dict(check=f.name.removesuffix('-seconds.json'),**r))
for line in (p/'counts-stable.log').read_text().splitlines():
 r=json.loads(line)
 if r.get('Test') and r['Action']=='pass':seconds.append(dict(check='counts-stable',test=r['Test'],seconds=r.get('Elapsed',0),action='pass'))
assert all(r['action']=='pass' and r['seconds']<60 for r in seconds)
(p/'test-seconds.json').write_text(json.dumps(seconds,indent=2)+'\n')
lines=['Built: a-check header pass for all 22 Adamic programs carried by compiler slice 3, preserving every program body byte.',
'Commits: fixed base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; source 0a8d7ed8f798ff96915bb24892550dd1d77dfab8; delivery SHA is in the final response and git log.',
'Commands and outputs: branch-own adamic c reports 16 checked, zero NotYet, six refused, zero type errors; the unchanged pinned a-check predicate accepts all 22 after the header edits.',
'Mutants: removed-header and wrong-reason on function_call.a both remain genuinely refused and each makes the unchanged predicate exit 1.',
'Not covered: a whole-tree scan, full packages/gate, unrelated production mutants, macOS/WASI and real checker ABI execution; original slice evidence remains under review/compiler/loom-r-slice-3.',
'', '## Branch and scope', '',
'Fetched origin, created compiler/loom-r-slice-3-ac from exactly dbdf977f5b, and applied git cherry-pick --no-commit 0a8d7ed8f798ff96915bb24892550dd1d77dfab8. The delivery is one commit on that fixed base, not an extra commit on the earlier branch. origin/main advanced during fetch; the requested base stays fixed.',
'', 'The changed-program list is git diff --name-only dbdf977f5b 0a8d7ed8 -- \'*.a\'. It contains 22 paths, including every location in the tree: there are no .a additions or modifications under review/. paths.txt records the exact list. New mutant source captures use .a.txt, not executable .a files. No new program was introduced by this header pass.',
'', 'Read CLAUDE.md, docs/front-door.md, and stage3/a-check-headers/README.md whole. The language and memory documents had been read during the original slice work in this conversation. The README refers to Gate.aCheck at fbac28c62493f27a788edc02a18bc8edb68de5da. Its object was initially unavailable locally; fetching origin devtools/fast-gate made it available. gate.py is byte-identical to that exact pinned cloud/fast-gate/run.py, verified in predicate-source.json.',
'', 'scan.py invokes that unchanged Gate.aCheck. Its adapter redirects only the compiler build output into the local evidence tools directory, supplies hard subprocess limits, and captures each real compiler observation. It does not change classification or header matching. Every row is compiled in place with this branch\'s own cmd/adamic via adamic c, exactly the frontend/emission operation a-check uses. Initial scan: four predicate findings among the six refusals. Final scan: exit 0, all 22 accepted.',
'', '## Headers and body proof', '',
'Conservative reading of "clean programs take no header": remove the six redundant a-check pass comments, rather than carrying them forward. Replace all six refusal headers with the complete measured refusal reason and fix, including the two shorter headers that already matched as substrings. No type-error header is needed. Updated the assignment bucket README\'s claim that every file begins with a pass header.',
'', 'No header was prepended to a previously headerless refused program: six existing refusal headers were corrected/canonicalized. Ten other .a files are byte-identical to 0a8d7ed8. For all 22 files, stripping only an existing first-line a-check comment gives a byte-identical body before and after (body-verification.json). Headers were restored after both mutants and this proof was rerun.',
'', '| File | New first-line header / action | Line shift |', '|---|---|---:|']
for h in headers:lines.append('| '+h['path']+' | '+(h['header'] or 'Removed redundant pass header')+' | '+str(h['line_shift'])+' |')
lines+=['', 'Position audit: position-references.txt records the *_test.go/testdata path search, and report-position-references.txt records the search for assignment filenames followed by pinned local line numbers in reports and metadata. No expectation needed updating. The six refusal replacements occupy the same one line, so the pinned 2:22, 2:23, 2:10, 3:16 and merged-field 2:1 diagnostics remain valid. The six assignment witnesses shift up one line, but no test or report pins their local diagnostic positions. Their upstream scanner.ts:1815-1828 provenance is unchanged. No _test.go expectation or program body was edited.',
'', '## Per-file compiler status', '',
'Counts: checked 16; NotYet 0; refused 6; type error 0; unclassified 0. Refused rows are expected and all are accepted by the header predicate. after.json retains full diagnostics, exit codes, first-line expectations, emitted-C hashes and per-file seconds.',
'', '| File | Status | Reason/code | Seconds |', '|---|---|---|---:|']
for name,r in after['rows'].items(): lines.append('| '+name+' | '+r['status']+' | '+(r['reason'] or r['code'] or '')+' | '+str(r['seconds'])+' |')
lines+=['', '## Header mutants', '',
'Both experiments use internal/oracle/testdata/refusal_rulings/function_call.a and the real branch compiler, not a fabricated or replayed diagnostic. Each rebuilds/reuses the branch compiler through the unchanged pinned predicate. Program bodies stay identical. A finally block restores the correct header. Full results are in removed-header.json, wrong-reason.json and mutants.json.',
'', '| Mutant | Compiler observation | Predicate result / catcher |', '|---|---|---|',
'| Remove first-line header | refused, reason unchanged; location moves to line 1 | exit 1: expected checked, observed refused |',
'| Replace with refused intentionally wrong reason 999999 | refused, reason unchanged | exit 1: expected reason is absent from actual diagnostic |',
'', 'Exact reproduction commands after sourcing the setup environment:', '', '```sh',
"timeout 180 python3 review/compiler/loom-r-slice-3-ac/scan.py after",
"timeout 180 python3 review/compiler/loom-r-slice-3-ac/scan.py mutants", '```',
'', '## Tests, reader guard, vet and counts', '',
'Header-affected checks pass: Function and merged-field oracle tests, all six assignment goldens and six assignment source mutants, the assignment fixture directory, and the fixture directory registry. These include source Node, JavaScript backend, native release, ASan/UBSan and leak checks where applicable. The tests pin actual diagnostic positions as well as meanings. TestCallTargetReaders passes with -timeout=90s. Vet passes for all 14 Go packages carried by the original compiler slice.',
'', 'All test runs are focused, not whole packages. Output goes directly to .log files. The runner sets GOMAXPROCS=4, four-CPU affinity and ADAMIC_GATE_UNCACHED=1. No test was over 60 seconds; no test was split, removed or rerun for exceeding its budget. Exact commands and wall times follow; test-seconds.json gives every test/subtest time.',
'', '| Check | Exit | Wall seconds | Command |', '|---|---:|---:|---|']
for c in checks:lines.append('| '+c['name']+' | '+str(c['exit'])+' | '+str(c['seconds'])+' | `'+ ' '.join(c['command']).replace('|','\\|')+'` |')
lines+=['', 'Counts: internal/oracle/counts.md is byte-identical to 0a8d7ed8. No regeneration was run. An isolated Go test overlay measures all 14 registered ruling/assignment rows directly and compares every numeric field to the existing table; all pass (package 2.691 seconds). The overlay is a .go.txt evidence file, not a new Go package or permanent test.',
'', '```sh', "GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1 timeout 180 go test ./internal/oracle -overlay=review/compiler/loom-r-slice-3-ac/counts-overlay/overlay.json -run '^TestHeaderCountsStayRecorded$' -count=1 -json -timeout=90s -parallel=4", '```',
'', 'Every recorded top-level test time:', '', '| Check | Test | Seconds |', '|---|---|---:|']
for r in seconds:
 if '/' not in r['test']:lines.append('| '+r['check']+' | '+r['test']+' | '+str(r['seconds'])+' |')
lines+=['', 'Compiler Go, runtime lines, dependency declarations and the original counts table are unchanged relative to 0a8d7ed8. No new runtime feature set or compiler-consuming package was added. The original slice\'s own implementation validation and 14 production mutant results are retained in its earlier report; this header pass reran the header-affected checks and its two new predicate mutants.',
'', '## Setup and delivery', '',
'Used the cloud runtime skill and checked the network policy. Toolchain command: export GOPROXY=\'https://proxy.golang.org|direct\'; ADAMIC_GOCACHE_OFF=1 timeout 600 bash cloud/setup.sh, then source the env file it printed. Local-cache mode is the supported fallback established after the shared cache stalled during the original unit. This setup succeeded with no new failure. nproc=5; quota is four CPUs. The full timing lines are:', '', '```text']
lines += [l for l in (p/'setup.log').read_text().splitlines() if l.startswith('setup:')]
lines+=['```', '', 'Go 1.27.1, clang 20.1.8, Node 24.19.0. No ahra command, PR or Loom submission. Push only compiler/loom-r-slice-3-ac, never main or the earlier branch. The unrelated earlier BLOCKED.md remains untracked. The tools binary and Python cache are local artifacts excluded from the commit.']
(p/'REPORT.md').write_text('\n'.join(lines)+'\n')
print('report complete; tests',len(seconds),'maximum seconds',max(r['seconds'] for r in seconds))
