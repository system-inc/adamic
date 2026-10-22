import json,re,subprocess,collections
from pathlib import Path
out=Path(__file__).resolve().parent
root=out.parents[2]
base=(out/'base.txt').read_text().strip()
source='a4b12a2c463e05ee612c064caa5d8db1c8486cf8'
results=[json.loads(l) for l in (out/'results.jsonl').read_text().splitlines()]
mutants=json.loads((out/'mutants-results.json.txt').read_text())
scan=json.loads((out/'after.json.txt').read_text())
counts=json.loads((out/'counts-changes.json.txt').read_text())
events=[]
for file in out.glob('*-seconds.json.txt'):
 if file.name!='touched-test-seconds.json.txt':events.extend(json.loads(file.read_text()))
owned={}
paths=subprocess.check_output(['git','diff','--name-only',base,'HEAD','--','*_test.go'],cwd=root,text=True).splitlines()
for path in paths:
 pkg='github.com/system-inc/adamic/'+str(Path(path).parent)
 for name in re.findall(r'^func (Test\w+)\(', (root/path).read_text(),re.M):owned[(pkg,name)]=path
owned_events=[e for e in events if (e['package'],e['test'].split('/')[0]) in owned]
(out/'touched-test-seconds.json.txt').write_text(json.dumps(owned_events,indent=2)+'\n')
failed=[r for r in results if r['exit']]
overbudget=[e for e in events if e.get('seconds',0)>=60]
lines=[
'Built: re-cut Slice 3 as one commit on the newest main fetched on resumption, retaining both refusal checks.',
f'Commit: delivery branch compiler/loom-r-slice-3-ac-m2; base {base}; full delivery SHA is in the final response.',
'Commands and outputs: exact commands, attempts and test timings are listed below and in results.jsonl.',
f'Mutants: {sum(r["caught"] for r in mutants)}/{len(mutants)} production mutants caught; removed-header and wrong-reason mutants rejected.',
'Not certified: the blanket 60-second grain requirement fails on the existing TestProduct_GrainMarkdownFormatter (97.60s) and TestProduct_MarkdownblocksTextFormatter (130.04s) cold builds. Neither was rerun. No ahra, Loom submission, full repository gate or macOS execution; skipped cases remain skips.',
'',
'## Base, conflicts and assumptions',
'',
f'Source: {source}, applying only its own parent-to-commit change. Initial main was 561aeb4b851ca0820f39440a09dc5fb3fd27ba40. The resume fetch returned {base}; that newer main is the delivery base. Assumption: hold this fetched main snapshot throughout validation rather than continually chasing main. Its eep presence checks, tests and witnesses are retained unchanged. Every changed program is byte-identical to the source commit.',
'',
'One refusal-pass hunk conflicted: main called knownLoweringGap(node), while Slice 3 called functionCallRefusal(node) at the same position. Retained both error-return blocks, with main\'s known-gap check first. Also retained main\'s literal-callable branch selection and the source\'s module-level merged-field and Function-annotation checks. refusal-resolution.patch records the final resolution. No newer main line was reverted by the resolution.',
'',
'counts.md conflicted. Started from the final main table and regenerated it on Linux with house clang 20.1.8, never hand-merged. The new-main automatic merge retains its exception, expression, emission and runtime-header additions together with Slice 3. Private unreferenced validation snapshots make corpus inputs clean against HEAD; they are not delivery commits. The final branch contains exactly one commit above the final main base.',
'',
'Completed IR, JavaScript, lower and flow runs on the earlier base are preserved separately under earlier-base-561aeb4b and are not credited as final-base coverage. Its partial native product run was interrupted to incorporate the newly fetched main. No completed over-budget leaf was rerun. The requested complete recursive validation takes priority over the older 30-minute first-green workflow.',
'',
'## Runtime clearance',
'',
'Extra changed lines in internal/native/runtime/*.c or *.h beyond the source commit\'s own delta: NONE. Added and removed lines match by file in adamic.h, closure.c and union.c. runtime-clearance.json.txt and the two runtime patches retain the exact comparison. Main\'s newer header declarations and host-runtime include remain intact.',
'',
'## Counts',
'',
'Final-base command: timeout 840 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -json -parallel=4 -timeout=13m -args -update-counts. Complete output: counts-update.jsonl. Regenerated once on the final base, after a separate successful regeneration on the superseded base. All 33 changed rows belong to the source change and exactly equal its recorded values: 22 additions, 11 existing rows. No unrelated row changed.',
'',
'| Fixture | Before (alloc/free/retain/release/peak/region) | After |',
'|---|---|---|']
for row in counts:lines.append('| '+row['fixture']+' | '+('/'.join(row['before']) if row['before'] else 'new')+' | '+('/'.join(row['after']) if row['after'] else 'removed')+' |')
lines += ['', '## Tests, reader guard and vet', '',
'Go runs use GOMAXPROCS=4 under the four-CPU cgroup quota; the recursive suite and mutant drivers also pin four-CPU affinity. Product owners run first with -parallel=1. Recursive consumer suites use -skip ^TestProduct_ -parallel=4 -p 1, so the product leaves are covered by their preceding owner runs and are not rerun. Child oracles packages have their own preceding product runs. Each command has an outer hard limit; tests also have a Go timeout. Output goes directly to logs. Results include setup and every failed attempt, if any.', '',
'| Command label | Exit | Wall seconds | Failed test events |', '|---|---:|---:|---:|']
for row in results:lines.append(f'| {row["name"]} | {row["exit"]} | {row["seconds"]} | {len(row["failed"])} |')
lines += ['',f'Final-base recorded test events: {dict(collections.Counter(e["action"] for e in events))}. Maximum in source-touched test files: {max((e["seconds"] for e in owned_events),default=0):.3f}s. touched-test-seconds.json.txt lists every such test and its seconds. All timing events at or above 60s: '+json.dumps(overbudget)+'.', '',
'Exact commands:', '', '```text']
for row in results:
 limit=1440 if row['name']=='production-mutants' else 240 if row['name']=='assignment-proof-driver' else 120 if row['name'].startswith('reader-guard') else 180 if row['name'].startswith('vet') else 840
 lines.append('timeout '+str(limit)+' '+' '.join(row['command']))
lines += ['```', '', '## Production mutants', '',
'Every overlay source is .go.txt; each negative run must fail an actual test and excludes Go/C build errors and test timeouts as catchers. The source catalog supplies all 14 mutations. Embedded assignment, call-spread and JSON-binding source mutants are also covered by the oracle suite.', '',
'| Mutant | Catcher | Caught | Seconds |', '|---|---|---|---:|']
for row in mutants:
 command=row['command'];catcher=command[command.index('-run')+1]
 lines.append(f'| {row["name"]} | {catcher} | {row["caught"]} | {row["seconds"]:.3f} |')
lines += ['', '## A-check', '',
'All 22 source-added or source-modified .a files were compiled in place using the final branch compiler. The unchanged pinned Gate.aCheck predicate passes all 22: 16 checked, six refused, zero NotYet, zero checker errors. No header or program body needed changing on this re-cut. program-byte-check.json.txt records the source-byte comparison. Both removed-header and wrong-reason mutants on refusal_rulings/function_call.a stay genuinely refused and make the predicate exit 1; mutants.json.txt records the catches.', '',
'| File | Status | Reason/code |', '|---|---|---|']
for path,row in scan['rows'].items():lines.append('| '+path+' | '+row['status']+' | '+(row.get('reason') or row.get('code') or '').replace('|','\\|').replace('\n',' ')+' |')
lines += ['', '## Toolchain and evidence', '',
'GOPROXY=https://proxy.golang.org|direct was exported before bash cloud/setup.sh, then /workspace/adamic-tools/env.sh was sourced. nproc=5; cgroup quota 400000/100000 is four CPUs. Go 1.27.1, Node 24.19.0, Linux amd64, clang 20.1.8. setup.log retains all output. Setup timing lines:', '', '```text']
lines.extend(l for l in (out/'setup.log').read_text().splitlines() if l.startswith('setup:'))
lines += ['```', '',
'Existing Go cache data was reused in the task-local cache. Only byte-verified identical export archives were hard-linked to bound disk use; product names and bytes were preserved. Cache records and deduplication logs are retained. Reclaimed 3,259,256,352 bytes of stale pre-unit Adamic compiler archives, identified by their Go compilation-unit package symbols; current-unit archives and Cohere/toolchain dependencies were retained. The complete eviction inventory is saved. Product output and result caches use task-local scratch; no machine path was added to products or tests. Historical source review artifacts remain unchanged, including their pre-existing patch/log whitespace. New evidence uses .json.txt so it does not alter the JSON corpus census. No code was copied from cohere.', '',
'The pinned JSON corpus was provisioned through its unchanged TestProduct_JSONPinnedCorpus in an isolated worktree at the same validation snapshot. This avoids matching historical /workspace/adamic strings against the current builder path; no guard or corpus byte was changed. The ordinary checkout then reads the identical product (0.27s). json-relocated-product.log and json-portable-product-read.log record both runs. The temporary worktree referenced the existing cohere submodule and was removed after the successful read. npm ci --prefix stage3/api --ignore-scripts provisioned the exact lockfile dependencies for the corpus and stock TypeScript 6.0.3 proof driver.', '',
'The original commit message is preserved, including its historical WIP note. This re-cut independently reproduced the existing markdown producer grain failure: TestProduct_GrainMarkdownFormatter took 97.60s and TestProduct_MarkdownblocksTextFormatter took 130.04s while cold-building their standalone Cohere commands. Neither was rerun and no unrelated producer was weakened or edited. Deliver the requested branch as reviewable code, but hold it from Loom until the owner splits or prepares that producer within the budget. Functional pass is not claimed to satisfy this timing requirement.', '',
'The first native recursive consumer run failed only on toolchain setup: GCC 14.2.0 lacked sanitizer/allocator_interface.h, and the unguarded attributed-binding WASI leaf had no WASI_SYSROOT. No leaf exceeded 60 seconds in that attempt. The GCC ASan library exports both required allocator functions; the task-local GCC wrapper supplies only the house LLVM header and adds its SHA256 to the compiler identity. gcc-header-workaround.json.txt and the successful ABI probe record it. No sudo is available. cloud/setup.sh --wasi-sdk provisioned SDK 27 with clang 20.1.8 and the supported sysroot. setup-wasi-sdk.log retains its full timings. The repaired-toolchain native rerun is separately recorded rather than hiding the first attempt.', '',
'All evidence belongs to review/compiler/loom-r-slice-3-ac-m2/. No PR, ahra CLI, Loom submission, main push or force push was performed.']
(out/'REPORT.md').write_text('\n'.join(lines)+'\n')
print('Report written; failed commands:',len(failed),'timing events >=60:',len(overbudget))
