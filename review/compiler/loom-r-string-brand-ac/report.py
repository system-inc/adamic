import json,subprocess
from pathlib import Path
out=Path(__file__).resolve().parent
checks=json.loads((out/'checks.json').read_text());status=json.loads((out/'statuses.json').read_text());events=[];mutants=[]
assert all(r['exit']==(1 if r['label'].startswith('mutant-') else 0) for r in checks)
for p in out.glob('*.log'):
 rows=[]
 for line in p.read_text().splitlines():
  try:r=json.loads(line)
  except ValueError:continue
  if isinstance(r,dict) and r.get('Action') in ('pass','fail') and 'Test' in r:
   e={'log':p.name,'test':r['Test'],'seconds':r.get('Elapsed',0),'result':r['Action']};events.append(e);rows.append(e)
 if p.name.startswith('mutant-'):mutants.append({'topic':p.stem.removeprefix('mutant-'),'catchers':[e for e in rows if e['result']=='fail']})
assert len(mutants)==5 and all(m['catchers'] for m in mutants)
assert all(e['seconds']<60 for e in events)
(out/'test-times.json').write_text(json.dumps(events,indent=2)+'\n')
(out/'test-times.tsv').write_text('log\ttest\tseconds\tresult\n'+''.join(f"{r['log']}\t{r['test']}\t{r['seconds']}\t{r['result']}\n" for r in events))
(out/'lowering-mutants.json').write_text(json.dumps(mutants,indent=2)+'\n')
report='''Built: phantom string brands from 2194037a plus five exact first-line a-check refusal headers.
Commits: source 2194037a6242a32e20f2f526fb9e9bf59d14b4c5; base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; delivery SHA in final response.
Commands and outputs: all 12 changed .a files compile/classify as expected; focused tests, counts reads, consumers, guards and vet pass.
Mutants: removed header and wrong reason each fail unchanged Gate.aCheck; all five original lowering mutants are caught again.
Not covered: full repository a-check scan, full packages/gate, WASI, ahra or Loom submission.

Branch compiler/loom-r-string-brand-ac starts from the exact requested base, with source 2194037a applied using git cherry-pick --no-commit. No conflicts. origin/main fetched to a newer revision, but the explicit pinned base takes precedence over the generic current-main instruction. This is one commit containing the original change plus the five headers and new evidence. Compiler Go, runtime C, tests and all program bodies are unchanged from 2194037a. delta.json records the exact five-file change beyond the source, excluding this evidence directory.

Read stage3/a-check-headers/README.md whole and followed its scan/harness contract. Fetched fbac28c62493f27a788edc02a18bc8edb68de5da and saved its exact cloud/fast-gate/run.py in pinned-gate.py. Gate.aCheck is unchanged. The harness only reuses the separately built branch compiler instead of rebuilding it, and substitutes that compiler binary path in spawn, with a 20-second subprocess deadline. Compiler build: timeout 180 go build -o /tmp/string-brand-ac-adamic ./cmd/adamic, exit 0. Header scan and mutant driver: timeout 240 python3 review/compiler/loom-r-string-brand-ac/headers.py, exit 0. This ran real branch compiler invocations for all initial, header, final and mutant scans; diagnostics were not fabricated or replayed. gate-provenance.json records the source revision and hash.

Status totals: checked 7; NotYet 0; refused 5; type error 0. Checked here means adamic c exits zero, not that a runtime execution was tested. In particular, the review void-view probe checks cleanly and takes no header. statuses.json includes every full diagnostic, exact added header and original body SHA-256. All bodies were verified byte for byte against 2194037a after the final scan. The initial unheadered predicate scan failed for the five negative fixtures, as expected; header and restored-final scans pass all 12.

Position audit: searched complete fixture paths and string_brand/void-view-probe references in *_test.go and testdata, and searched Markdown reports for affected path:line pins. internal/oracle/string_brand_test.go is the reader; it checks named member/cast refusals, not diagnostic positions. No test or report expectation requires changing. Historical diagnostic logs remain historical evidence. path-readers.log, report-position-search.log and position-audit.json record this audit.

Only the five refused oracle fixtures gain headers. They remain under string_brand_refused/, excluded from flow's root fixture lowering. No clean fixture or review program changes. Header comments shift refusal diagnostics by one line; final-scan.json records the actual diagnostics with headers in place.

Counts are byte-identical to 2194037a; no regeneration was performed. The original six measured rows remain intact, and focused recorded-count checks and the table check pass. counts-unchanged.json records the hash. No runtime line, feature set, IR map, product, package, dependency declaration, build-directory literal or machine-path literal is introduced by this header pass. All guard tests passed: TestEveryMapAProgramHoldsIsSorted, gob stability/nil SortedMap checks, TestRuntimeSetsAreEachUsed, TestNativeAgreementRowsNameRegisteredPrograms, TestEveryTestIsParallelOrSaysWhy and TestCallTargetReaders (-timeout 90s). Vet passes on the change's internal/lower and internal/oracle packages. No test was removed or split.

Reran all string-brand lower/oracle tests, six accepted fixtures against Node, native ASan/UBSan, release native and JavaScript backend, and leak checks; five named refusal checks; flow/fresh corpus remainders; native consumers; counts reads; guards and vet. Every go test command has -count=1 -timeout 90s -parallel 4, GOMAXPROCS=4 and ADAMIC_GATE_UNCACHED=1. All output is direct to logs; the driver is bounded by timeout 900 and each subprocess by 240 seconds. All observed test leaves and parents remain under 60 seconds. The counts-read command's wall time includes several independent measurement products and does not represent one test leaf. test-times.tsv/json lists every observed test. Source whitespace check passes.

Setup used GOPROXY=https://proxy.golang.org|direct and timeout 600 bash cloud/setup.sh; sourced /workspace/adamic-tools/env.sh. House clang 20.1.8, Go 1.27.1, Node v24.19.0. nproc reports 5, cpu.max=400000 100000 (four CPUs). Cumulative timings: submodules 0.055s, Go 0.060s, Node 0.089s, Markdown 0.205s (dependency step 0.025s), clang 0.401s, shared cache 0.798s, Go build 18.005s, test binaries deferred 18.075s, build cache warm 18.076s, done 18.101s. Setup succeeded without workaround; setup.log contains all timing lines.

Assumptions: use the explicit pinned base; include every .a changed by the source, including review/; use the documented pinned predicate; copy the full measured refusal reason including its fix text into the header. No cohere code was copied. No PR or Loom submission was made. Push only the new own branch, without force.

Per-file status table:

| Path | Status | Header added |
|---|---|---|
'''
report+='\n'.join(f"| {r['path']} | {r['status']} | {'yes' if r['header'] else 'no'} |" for r in status['files'])+'\n'
report+='\nExact headers added:\n\n'
for r in status['files']:
 if r['header']:report+=f"{r['path']}:\n\n```\n{r['header']}\n```\n\n"
report+='Header mutants:\n\n'
for m in json.loads((out/'header-mutants.json').read_text()):report+=f"- {m['mutation']} on {m['file']}: exit {m['exit']}, caught by {m['caught_by']}.\n"
report+='\nOriginal lowering mutants rerun:\n\n'+'\n'.join('- '+m['topic']+': '+', '.join(r['test']+f" ({r['seconds']}s)" for r in m['catchers'])+'.' for m in mutants)
report+='\n\nCommands and outcomes (exact argv in checks.json):\n\n'+'\n'.join(f"- {r['label']}: exit {r['exit']}, {r['seconds']}s." for r in checks)
report+='\n\nAdded/touched test timings inherited from the source change:\n\n'+'\n'.join(f"- {r['test']}: {r['seconds']}s." for r in events if r['log'] in ('lower-focus.log','oracle-focus.log') and (r['test'].startswith('TestStringBrand') or r['test']=='TestClockGenericReturnsT01DeclinesBrandedPrimitives'))
report+=f"\n\nMaximum observed individual test: {max(r['seconds'] for r in events)}s.\n"
(out/'REPORT.md').write_text(report)
print('report generated; maximum individual test seconds:',max(r['seconds'] for r in events))
