import json
from pathlib import Path
out=Path(__file__).resolve().parent
checks=json.loads((out/'checks.json').read_text())
mutants=json.loads((out/'mutants/results.json').read_text())
times=json.loads((out/'test-times.json').read_text())
admission=json.loads((out/'admission.json').read_text())
counts=json.loads((out/'counts-attribution.json').read_text())
assert len(mutants)==5 and all(m['catchers'] for m in mutants)
assert all(r['exit']==(1 if r['label'].startswith('mutant-') or r['label']=='command-consumers-initial' else 0) for r in checks)
report='''Built: #fzx7q7y phantom string brands, applying only fb3c9989's own change on the exact pinned main.
Commits: source fb3c998910540944150ef08fcc35924139682fcc; base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; delivery SHA in final response.
Commands and outputs: focused compiler, oracle, consumer, guard and vet checks pass; exact argv/exits/seconds in checks.json.
Mutants: all five recorded mutants caught by their named tests; detailed catchers below.
Not covered: full packages/full gate, ahra CLI, Loom submission, WASI or the historical full-mode closure census.

Application and conflicts: git fetch origin, branch compiler/loom-r-string-brand at exact dbdf977f5b8b0979d414ebe6f20e7aadd312e309, then git cherry-pick --no-commit fb3c998910540944150ef08fcc35924139682fcc. No conflicted hunk exists. Git automatically merged expression.go, refusals.go and counts.md. Main's f93a65fb parallelMap lowering for both backends remains intact. Production, docs, tests and fixture +/- lines equal the source commit's own diff (source-delta-comparison.json). There were no manual code resolutions.

Counts: discarded the automatic counts merge in favor of the exact base table, then ran -update-counts exactly once on Linux with house clang 20.1.8 after setup compiled the tree. Only the six rows below were added; no existing row changed or disappeared. Values equal the source's six rows. Main's measurement-product counts format is retained. counts.patch and counts-attribution.json record the complete delta.

Runtime clearance: internal/native/runtime added/deleted lines are identical to source, both empty. There are no runtime lines requiring clearance; runtime-comparison.json contains both arrays.

Validation: all six accepted fixtures agree with Node, native ASan/UBSan, release native and the JavaScript backend, including leak checks. Five refusal fixtures remain under string_brand_refused/, so flow never tries to lower them as root fixtures. Lowering tests cover string/missing representations, layout exclusions, literal refinement and checked undefined removal. Flow and freshness corpus remainders, native and JavaScript consumers, CLI checks, compiler input/census checks, meter and test262 compiler-worker checks pass. Every test command uses -count=1 -timeout 90s -parallel 4, GOMAXPROCS=4 and ADAMIC_GATE_UNCACHED=1. Logs are direct files. Every observed test is below 60 seconds; test-times.tsv/json records each leaf and parent. Go formatting and source whitespace checks pass.

Guards: TestEveryMapAProgramHoldsIsSorted, gob stability/nil SortedMap checks, TestRuntimeSetsAreEachUsed, TestNativeAgreementRowsNameRegisteredPrograms, TestEveryTestIsParallelOrSaysWhy and TestCallTargetReaders all pass. Vet passes on changed packages internal/lower and internal/oracle. No new package, runtime feature set, IR map, product, build-directory literal or machine-path literal was introduced. Existing packages already declare compiler dependencies. All added/touched tests already call t.Parallel first; no test was deleted or split.

One extra CLI check initially failed because its fresh XDG cache hid the required runtime product. command-consumers-initial.log records this failure, which was not a compiler lowering error. Built the sanitized runtime owner and command product, exported ADAMIC_BUILD_CACHE_DIR to the normal shared cache so driver subprocesses retain it, then reran successfully. No source or other package test was modified. A formatting evidence command initially lacked gofmt on PATH; sourcing the printed setup env fixed it, and both formatting/whitespace checks passed.

Setup: exported GOPROXY=https://proxy.golang.org|direct, ran timeout 600 bash cloud/setup.sh and sourced /workspace/adamic-tools/env.sh. Setup succeeded with Go 1.27.1, Node v24.19.0 and clang 20.1.8. nproc=5; cpu.max=400000 100000, a four-CPU quota. Cumulative setup timings: submodules 0.072s, Go 0.074s, Node 0.075s, Markdown 0.182s (dependency step 0.010s), clang 0.389s, shared cache 1.191s, Go build 82.443s, test binaries deferred 82.515s, cache warm 82.517s, done 82.545s. setup.log records every timing line.

Assumption: the explicit pinned base takes precedence over the generic instruction to branch from origin/main. Keep main's current counts format and lowering infrastructure; apply only the requested commit. No cohere code was copied. No PR, ahra command or Loom submission was made. Delivery is one commit pushed only to compiler/loom-r-string-brand without force.

'''
report+=f"Fresh admission comparison: {admission['files']} oracle sources, exactly {len(admission['newly_admitted'])} new admissions, all six string_brand_ fixtures, and {len(admission['lost_admissions'])} lost admissions. Compared fresh CLIs with four workers and a 20-second per-file subprocess limit. The baseline overlays the exact pinned base's four existing hook files; the new proof file has no side effects and is uncalled in the baseline. admission.json records every observation, first-error counts and binary hashes; admission-builds.json records bounded build commands. This is a first-error scan, not the historical full-mode census.\n\n"
report+='Commands and observed outcomes (exact argv in checks.json):\n\n'+'\n'.join(f"- {r['label']}: exit {r['exit']}, {r['seconds']}s." for r in checks)
report+='\n\nRecorded mutants and catchers:\n\n'+'\n'.join('- '+r['topic']+': '+', '.join(c['test']+f" ({c['seconds']}s)" for c in r['catchers'])+'.' for r in mutants)
report+='\n\nAdded/touched test times:\n\n'+'\n'.join(f"- {r['test']}: {r['seconds']}s." for r in times if r['log'] in ['lower-focus.log','oracle-focus.log'] and (r['test'].startswith('TestStringBrand') or r['test']=='TestClockGenericReturnsT01DeclinesBrandedPrimitives'))
report+='\n\nCounts additions:\n\n| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |\n|---|---:|---:|---:|---:|---:|---:|\n'+'\n'.join(counts['added'])+'\n'
(out/'REPORT.md').write_text(report)
