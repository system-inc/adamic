import json, subprocess
from pathlib import Path
out=Path(__file__).parent
base=subprocess.check_output(["git","rev-parse","origin/main"],text=True).strip()
source=subprocess.check_output(["git","rev-parse","origin/compiler/loom-slice-3-2"],text=True).strip()
checks=[json.loads(line) for line in (out/"checks.jsonl").read_text().splitlines()]
mutants=json.loads((out/"mutants-results.json").read_text())
counts=json.loads((out/"counts-attribution.json").read_text())
leaves=[]
for path in out.glob("*.log"):
 for line in path.read_text().splitlines():
  if not line.startswith("{"):continue
  try:event=json.loads(line)
  except ValueError:continue
  if event.get("Test") and event.get("Action")=="pass":leaves.append(dict(log=path.name,package=event.get("Package"),test=event["Test"],seconds=event.get("Elapsed",0)))
assert all(row["seconds"]<60 for row in leaves)
(out/"leaf-seconds.json").write_text(json.dumps(leaves,indent=2)+"\n")
text=f"""Built: slice 3 re-cut as one commit, retaining assignment proofs, optional-read exceptions, call spread, JSON shorthand, frontend rulings and TSGo caching.
Commits: base {base}; source {source}; delivery SHA is in the final response and git log.
Commands and outputs: focused fixtures, seven stage-1 gap packages, all guards, readers and 14-package vet pass; counts add 22 rows and change 11 owned existing rows.
Mutants: all 14 production mutants and eight embedded mutants were caught; each result and catcher is below.
Not covered: full package suites, the full gate, WASI, macOS execution and real checker ABI execution. No PR or Loom submission was made.

## Main and the conflict

Exactly one textual conflict occurred, in internal/native/tsgo.go. Main's ae1d1c703eef648048dd94e7d53581dba8862b99 added a withCompilerToken wrapper around the checker link, enforcing ADAMIC_CC_JOBSERVER. The source selects compilerName(options) and caches the TSGo runtime through splitTSGoRuntime. Resolved once by retaining both the selected compiler and token wrapper. The source's separate Darwin compile also takes a token, preserving main's every-clang rule for that new path. Relative inputs, runtime caching, feature flags, prefix maps and stable Darwin object metadata remain. Linux TSGo cache, placeless archive and jobserver guards pass. The cache tests link inert ABI stubs; they do not exercise the real checker ABI.

Current main's counts API moved measurement into keyed products (1f07ce2d, followed by ac8ffaac and 9f8ce955) and removed the digest column again. Conservative assumption: use the fetched tip's counts-only schema and current preparation/measurement API. Assignment registrations return the new tuple. The six frontend-mode registrations prepare C once and return a measurement closure; the physical TypeScript copy's path is passed through main's countPortable, so a temporary directory cannot enter the digest.

The first oracle guard invocation did not build because the assignment registrations still returned only a string. No test ran or mutant credit was claimed. Converted those registrations and ran the actual guard successfully. All source artifacts remain .a; the source's existing temporary .ts frontend adapter is retained. Stock TypeScript 6.0.3 dependencies were installed with timeout 120 npm ci --prefix stage3/api --ignore-scripts (npm.log).

## Counts and runtime

Took main's complete counts.md, then regenerated exactly once on Linux with house clang 20.1.8. The 22 additions and 11 existing changes have no removals. counts-attribution.json records every full row. An isolated Go overlay reverting only internal/lower/narrowed.go, internal/lower/exceptions.go, internal/native/reuse.go and internal/flow/build.go to main reproduces all 11 old main rows exactly (counts-counterfactual.log, 2.862 seconds in the test process). This was a measurement proof, with no second regeneration. Exact command: GOMAXPROCS=4 timeout 180 go test ./internal/oracle -overlay=review/compiler/loom-slice-3-3/counts-counterfactual/overlay.json -run ^TestSlice3CountsAttribution$ -count=1 -json -timeout=90s -parallel=4; exit 0. New rows and their numeric changes agree with the source's recorded ownership.

Runtime +/- lines are identical to the source: adamic.h +1, closure.c +1, union.c +3, no removals or differences (runtime-comparison.json).

| Existing row | Main A/F/R/L/P/G | Re-cut A/F/R/L/P/G |
|---|---|---|
"""
for key,value in counts["changed"].items():
 def numbers(line):return "/".join(line.split(" | ")[1:]).removesuffix(" |")
 text+="| "+key+" | "+numbers(value["main"])+" | "+numbers(value["recut"])+" |\n"
text+="""
## Guards and grain

TestEveryMapAProgramHoldsIsSorted, TestRuntimeSetsAreEachUsed, TestNativeAgreementRowsNameRegisteredPrograms, TestEveryTestIsParallelOrSaysWhy and TestCallTargetReaders all pass. No new compiler-consuming Go package or runtime feature set was introduced, so the dependency registry and runtime witness declarations needed no additions. The source's existing serial markdown helper explanation satisfies the parallel rule. No test was deleted or split; every completed affected leaf is below 60 seconds. leaf-seconds.json lists every selected leaf's seconds. Tests include setup in their clock. Runtime products were prepared one group at a time.

The TSGo cold/warm release and sanitized builds agree with source Node, retain identical binary bytes, and do not replace their runtime archive. The emitted binaries name no build directory. The native jobserver tests pass. Changed-package vet includes internal/javascript and all seven changed stage-1 packages, 14 packages total. No code was copied from cohere.

Setup completed in 51.649 seconds. Every timing line is retained in setup.log. Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc=5 and cgroup cpu.max=400000 100000. All builds and tests sourced the setup environment with GOPROXY=https://proxy.golang.org|direct; suites used GOMAXPROCS=4 and four-CPU affinity. Counterfactual measurement used GOMAXPROCS=4 on the same four-CPU quota. No whole package suite or full gate was run; setup compiled all repository packages.

## Commands

| Check | Exit | Wall seconds | Command |
|---|---:|---:|---|
"""
for row in checks:text+="| "+row["name"]+" | "+str(row["exit"])+" | "+str(row["seconds"])+" | `"+" ".join(row["command"]).replace("|","\\|")+"` |\n"
text+="\n## Production mutants\n\n| Mutant | Caught | Seconds | Catching tests |\n|---|---|---:|---|\n"
for row in mutants:
 log=(out/"mutants"/row["name"]/"test.txt").read_text()
 import re
 failed=re.findall(r"--- FAIL: (Test[^ ]+)",log)
 text+="| "+row["name"]+" | "+str(row["caught"])+" | "+str(round(row["seconds"],3))+" | "+", ".join(failed)+" |\n"
text+="""
Every production mutant exited 1 at a behavior assertion; build failures and timeouts are excluded. Exact commands and raw failures are in mutants-results.json and mutants/*/test.txt. Mutant Go sources use .go.txt.

The eight embedded mutants are TestAssignmentProofScannerKeywordMutant, TestAssignmentProofIfNarrowingMutant, TestAssignmentProofReturnDefinedMutant, TestAssignmentProofChainedMutant, TestAssignmentProofCompoundMutant, TestAssignmentProofWiderTargetMutant, TestCallSpreadEvaluatedTwiceMutant and TestJSONStringifyShorthandWrongBindingMutant. Each successfully builds/runs altered valid source or IR and requires a disagreement with independently captured Node bytes. Assignment witnesses also pass stock TypeScript with zero diagnostics; semantic execution uses release, JavaScript, ASan/UBSan and leak checks. TSGo's feature omission negative control is an include-order compile contract, not a semantic mutant credit.
"""
(out/"REPORT.md").write_text(text)
