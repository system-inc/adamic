Renamed 300 corpus inputs with git mv to exact upstream .ts names; bytes and hash pins unchanged.
Retained the other 51 headers; added 17 host TS2591 headers: 68 total, 13 refusal and 55 type-error.
Node tsc goldens pass 301/301; corpus audit passes; a-check scans 375 inputs, accepts 201 and lists 174 findings.
Driver output/provenance mutants, 34 host-header mutants and the host source-span mutant are caught.
Host check.py fails on recorded diagnostic provenance; source-audit.cjs passes. No full repository/native tsc gate.

# Stage 3 a-check headers

## Current ruling and changes

The 21:50 @system_adamic ruling supersedes the proposed corpus exemption. Upstream compiler tests are TypeScript data, not authored Adamic programs. All **300** corpus inputs were renamed with `git mv` from input.a to the exact basename in selection.json's upstream `source`. This selection contains 300 `.ts` files and no `.tsx` or `.d.ts` files. Every renamed source is byte-identical both to its existing SHA-256 pin and to the source in the pinned upstream checkout, commit 050880ce59e30b356b686bd3144efe24f875ebc8. BOMs, CRLFs and upstream comments are intact. [corpus-renames.json](corpus-renames.json) lists every old/new path.

selection.json retains its upstream source paths and unchanged byte pins, and now records each local source as `path`. The importer, audit, driver and provenance mutants use those names. The audit also validates that every local path has the exact upstream filename. The shell wrapper needs no change because it delegates to the updated driver. Goldens, expected outputs, baseline files and baseline hashes are unchanged. The two authored tiny negative programs retain their a-check headers; the driver omits those first-line gate comments only when materializing the tiny project, preserving the existing diagnostic line positions. Corpus source bytes are not altered or stripped of gate metadata.

All **51** existing header files are byte-identical to the previous branch head e1fe34ce. Exactly **17** host fixtures whose measured first a-check diagnostic is TS2591 gained `// a-check: type error TS2591` above their existing comments. The other eight host fixtures start with TS2345 or TS2503 and remain findings. No text below the inserted host header changed. No developer-tools exemption or compiler change was made.

## Counts and a-check

The exact pinned Gate.aCheck method at fbac28c62493f27a788edc02a18bc8edb68de5da was rerun after all header mutants were restored. It scans **375** `.a` files under stage3; **zero** corpus inputs are selected. Results: **81 clean, 52 not-yet, 135 refused, 107 type errors**. Clean and not-yet both count as checked. There are **13 refusal headers, 55 type-error headers, 174 findings and zero unclassified files**. The raw scan exits 1 and lists exactly those 174 findings (122 refusals, 52 type errors). The identical gate over the other **201** files exits 0. No failures are suppressed in the gate.

The compact table groups cases/probes under their owning directory. [directories.md](directories.md) gives every exact parent directory. The 300 renamed inputs are counted separately from a-check inputs.

| Directory | Refusal headers | Type-error headers | Findings | Pass without header | Renamed corpus | Unclassified |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| stage3/adapt/48-memoize | 0 | 0 | 4 | 0 | 0 | 0 |
| stage3/adapt/75-optional-widening | 0 | 0 | 2 | 0 | 0 | 0 |
| stage3/census/repro | 13 | 25 | 0 | 16 | 0 | 0 |
| stage3/drivers/parser | 0 | 0 | 5 | 4 | 0 | 0 |
| stage3/drivers/scanner | 0 | 0 | 16 | 14 | 0 | 0 |
| stage3/drivers/tsc | 0 | 2 | 0 | 1 | 300 | 0 |
| stage3/fixtures/assertions | 0 | 0 | 18 | 2 | 0 | 0 |
| stage3/fixtures/cycles | 0 | 0 | 25 | 18 | 0 | 0 |
| stage3/fixtures/enums | 0 | 0 | 3 | 11 | 0 | 0 |
| stage3/fixtures/host | 0 | 17 | 8 | 0 | 0 | 0 |
| stage3/fixtures/namespaces | 0 | 0 | 12 | 0 | 0 | 0 |
| stage3/fixtures/nested-functions | 0 | 0 | 2 | 9 | 0 | 0 |
| stage3/fixtures/objects | 0 | 0 | 16 | 9 | 0 | 0 |
| stage3/fixtures/predicates | 0 | 0 | 10 | 1 | 0 | 0 |
| stage3/fixtures/records | 0 | 4 | 35 | 30 | 0 | 0 |
| stage3/fixtures/runner | 0 | 0 | 0 | 3 | 0 | 0 |
| stage3/fixtures/taste | 0 | 0 | 12 | 14 | 0 | 0 |
| stage3/ledger/checker-259 | 0 | 7 | 6 | 1 | 0 | 0 |
| **Total** | 13 | 55 | 174 | 133 | 300 | 0 |

## Node driver and provenance verification

The upstream checkout was verified at the pinned commit. A fresh scratch tree was built with the current registered stage3 adaptations using `STAGE3_CACHE=/tmp/stage3-ruling/cache bash stage3/apply.sh /tmp/stage3-ruling/adapted`, then `npm ci --prefix /tmp/stage3-ruling/adapted --ignore-scripts --no-audit --no-fund` and `npx hereby tsc --no-typecheck` inside that tree. This targeted Node bundle is a runtime oracle; it does not claim the full adapted compiler's typechecking or native compilation succeeds. Build/apply logs are retained.

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/drivers/tsc/audit.py > /tmp/stage3-ruling/corpus-audit.log 2>&1
TSC_RESULTS=/tmp/stage3-ruling/node-goldens stage3/drivers/tsc/run.sh node /tmp/stage3-ruling/adapted/built/local/tsc.js > /tmp/stage3-ruling/node-goldens.log 2>&1
python3 stage3/drivers/tsc/mutants.py node /tmp/stage3-ruling/adapted/built/local/tsc.js > /tmp/stage3-ruling/driver-mutants.log 2>&1
```

The provenance audit exits **0**: `PASS provenance, 300 cases, 60 clean, baseline summaries, headers, golden bytes and exits`. The ordinary Node driver exits **0**, with **301/301** projects passing in **53.234s**, comparing every stdout byte, stderr byte and exit code against the unchanged goldens. The tiny project passes too, proving its preserved headers do not move materialized diagnostic positions. [logs/node-goldens-report.json](logs/node-goldens-report.json) records the command and count.

The driver mutant suite exits 0 and catches three independent golden mutants (diagnostic column, stderr, exit) and nine provenance mutants (source hash, baseline hash, baseline summary, expected stderr, expected exit, case population, header options, new source-path guard, baseline codes). Its Node-forwarder smoke proves native-slot plumbing only. No actual native tsc run is claimed.

All 17 new host headers were separately mutated to TS999999 and removed, for **34/34** catches by the unchanged gate. Each was restored in a finally block. [host-header-mutants.json](host-header-mutants.json) records every exact catcher. Earlier header-mutant logs remain historical evidence; no corpus header remains in the current tree.

## Host source audits and provenance failures

Both audit programs and status.json were left unchanged. The actual source-span audit was run before and after the headers:

```sh
NODE_PATH=/workspace/adamic/stage3/api/node_modules node stage3/fixtures/host/source-audit.cjs /tmp/stage3-ruling/upstream > /tmp/stage3-ruling/host-source-audit-after.log 2>&1
python3 stage3/fixtures/host/check.py --mutants --logs /tmp/stage3-ruling/host-check-after > /tmp/stage3-ruling/host-check-after.log 2>&1
```

source-audit.cjs exits **0** both times: `pass: 124 upstream function and method spans retain their tokens`; its SHA256-to-SHA1 mutant is caught. It compares parsed syntax without comments, so the new first line does not break its spans. Original upstream `From ...` provenance comments are unchanged below each header.

check.py exits **1** before and after this change at `AssertionError: 01_readFile_utf8.a stage0: exact observation differs`. That fixture has no new header. The recorded diagnostic lacks the final newline and `exit status 1` emitted by the current `go run` invocation; the before-header diff proves this existing failure. The stop prevents check.py from reaching later fixtures and its --mutants phase, so that phase is not claimed to have passed.

The 17 new headers additionally move source diagnostic lines by one, which breaks check.py's exact recorded stage0 provenance if it reaches those fixtures. For example, 06_fileExists.a's first TS2591 moves from **10:24 to 11:24**; its node:fs import moves from **8:22 to 9:22**. The affected files are listed below. A supplemental run applied the same exact recorded comparisons to all 25 files without modifying the audit: all **25 Node goldens match**, but all 25 current stage0 records differ, including the existing go-run trailer mismatch and the added host line shifts. [host-audit-details.json](host-audit-details.json) retains each complete diagnostic diff. These are reported failures, not refreshed expectations or weakened checks.

| New host header | First diagnostic line before | After |
| --- | --- | --- |
| stage3/fixtures/host/06_fileExists.a | 10:24 | 10:24 |
| stage3/fixtures/host/07_directoryExists.a | 10:24 | 10:24 |
| stage3/fixtures/host/08_getDirectories.a | 18:22 | 19:22 |
| stage3/fixtures/host/09_realpath.a | 6:22 | 10:21 |
| stage3/fixtures/host/10_getModifiedTime.a | 12:41 | 10:21 |
| stage3/fixtures/host/11_setModifiedTime.a | 4:22 | 5:22 |
| stage3/fixtures/host/12_deleteFile.a | 4:22 | 5:22 |
| stage3/fixtures/host/13_createDirectory.a | 10:24 | 10:22 |
| stage3/fixtures/host/14_getCurrentDirectory.a | 5:22 | 6:22 |
| stage3/fixtures/host/15_getExecutingFilePath.a | 5:22 | 6:22 |
| stage3/fixtures/host/17_write.a | 4:22 | 5:22 |
| stage3/fixtures/host/18_exit_0.a | 5:22 | 6:22 |
| stage3/fixtures/host/19_exit_1.a | 5:22 | 6:22 |
| stage3/fixtures/host/20_exit_2.a | 5:22 | 6:22 |
| stage3/fixtures/host/21_createHash.a | 11:26 | 10:21 |
| stage3/fixtures/host/22_createHash_fallback.a | 5:22 | 6:22 |
| stage3/fixtures/host/24_useCaseSensitiveFileNames.a | 11:22 | 12:22 |

## Scope, setup and history

Fetched current main and created the requested branch. Read CLAUDE.md and its prerequisite documents, then the fast-gate implementation at fbac28c6 before the scan. Setup used `export GOPROXY='https://proxy.golang.org|direct'`, `bash cloud/setup.sh > /tmp/stage3-a-setup.log 2>&1`, and `source /workspace/adamic-tools/env.sh`. Setup succeeded without a workaround. `nproc=5`, CPU quota four, Go 1.27.1, Node 24.19.0, clang 20.1.8. Complete timing lines are in [logs/setup.log](logs/setup.log): Node 0.021s; Go 0.032s; clang 0.228s; markdown install step 1.660s, ready 1.732s; submodules 317.574s; Go build 479.572s; tests deferred 479.676s; cache warm 479.678s; done 479.714s.

This ruling is delivered as a new descendant commit after e1fe34ce, without rebasing or rewriting history. Changes are restricted to the tsc driver, host first lines, and this report/evidence. The corpus golden streams and host audits are unchanged. Earlier scans and the initial corpus-header mistake are retained in Git history and old logs; they are not the current findings or an exemption request. No full repository gate, full upstream compiler suite, native tsc implementation run, or full host check.py mutant pass is claimed. `git diff --check` passes.

## Reproduce a-check

Save the harness below to /tmp/stage3-a-check/run.py. Fetch devtools/fast-gate if the pinned object is unavailable. Run from the repository root after sourcing setup's env.sh:

```sh
mkdir -p /tmp/stage3-a-check /tmp/stage3-ruling
git show fbac28c62493f27a788edc02a18bc8edb68de5da:cloud/fast-gate/run.py > /tmp/stage3-a-check/gate.py
python3 /tmp/stage3-a-check/run.py ruling-final > /tmp/stage3-ruling/acheck-final.log 2>&1
# Whole stage3: exit 1, exactly the 174 listed findings; no corpus inputs.
python3 - <<'PYCODE'
import json
from pathlib import Path
rows = json.loads(Path('stage3/a-check-headers/outcomes.json').read_text())
Path('/tmp/stage3-ruling/accepted-paths.json').write_text(json.dumps(
    [p for p, r in rows.items() if r['classification'] in ('header', 'checked')]))
PYCODE
python3 /tmp/stage3-a-check/run.py ruling-accepted /tmp/stage3-ruling/accepted-paths.json > /tmp/stage3-ruling/acheck-accepted.log 2>&1
# Exit 0, 201 files.
```

```python
import importlib.util, json, pathlib, subprocess, sys, time, collections
spec=importlib.util.spec_from_file_location('fastgate','/tmp/stage3-a-check/gate.py'); mod=importlib.util.module_from_spec(spec); spec.loader.exec_module(mod)
root=pathlib.Path.cwd(); label=sys.argv[1]
class Process:
 def __init__(self,p,path): self.p=p; self.path=path
 def communicate(self):
  out,err=self.p.communicate(); self.returncode=self.p.returncode
  diagnostics[self.path]={'exit':self.returncode,'stderr':err,'stdout_bytes':len(out.encode())}
  return out,err
class Harness:
 def __init__(self):
  self.arguments=type('Args',(),{'tree':str(root)})(); self.result={};self.steps={};self.exits={};self.failure=None
 def step(self,name,cmd):
  with open('/tmp/stage3-a-check/'+label+'-build.log','w') as log: code=subprocess.run(cmd,cwd=root,stdout=log,stderr=subprocess.STDOUT).returncode
  assert code==0,code
  return True
 def spawn(self,cmd,stdout,stderr): return Process(subprocess.Popen(cmd,cwd=root,stdout=stdout,stderr=stderr,text=True),cmd[-1])
 def fail(self,name,detail): self.failure=detail
paths=json.loads(pathlib.Path(sys.argv[2]).read_text()) if len(sys.argv)>2 else sorted(str(p.relative_to(root)) for p in (root/'stage3').rglob('*.a')); diagnostics={}; h=Harness(); mod.Gate.aCheck(h,paths)
for path,row in h.result['a_check'].items(): row.update(diagnostics[path])
pathlib.Path('/tmp/stage3-a-check/'+label+'.json').write_text(json.dumps(h.result['a_check'],indent=2)+'\n')
print(label,len(paths),dict(collections.Counter(r['outcome'] for r in h.result['a_check'].values())),h.exits,h.steps)
print(h.failure or 'PASS')

sys.exit(h.exits["a-check"])
```

## Every remaining finding

These 174 `.a` files have no expected-error header. Full current/baseline diagnostics, classification reasons and evidence paths are in [outcomes.json](outcomes.json). Renamed upstream TypeScript inputs are recorded as data and are no longer findings in a-check.

| File | Outcome | Current first diagnostic |
| --- | --- | --- |
| stage3/adapt/48-memoize/a-explicit.a | refused | adamic: /workspace/adamic/stage3/adapt/48-memoize/a-explicit.a:1:28: Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; write the fun |
| stage3/adapt/48-memoize/a.a | refused | adamic: /workspace/adamic/stage3/adapt/48-memoize/a.a:4:13: Adamic 0.1 refuses a value as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/adapt/48-memoize/b-explicit.a | refused | adamic: /workspace/adamic/stage3/adapt/48-memoize/b-explicit.a:3:9: Adamic 0.1 refuses 'pending', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; write the funct |
| stage3/adapt/48-memoize/b.a | refused | adamic: /workspace/adamic/stage3/adapt/48-memoize/b.a:5:13: Adamic 0.1 refuses a value as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/adapt/75-optional-widening/coverage/reduced-union.a | refused | adamic: /workspace/adamic/stage3/adapt/75-optional-widening/coverage/reduced-union.a:3:31: Adamic 0.1 refuses optional property id in { id?: number; } absent from structural source never, which can hide fields; declare id on the source type, or build a fresh object with known fields (adamic/no-optio |
| stage3/adapt/75-optional-widening/probes/read-write.a | type error | stage3/adapt/75-optional-widening/probes/read-write.a:11:13: error TS2345: Argument of type 'number &#124; undefined' is not assignable to parameter of type 'string'. |
| stage3/drivers/parser/main.a | type error | stage3/drivers/parser/main.a:3:21: error TS2591: Cannot find name 'node:fs'. Do you need to install type definitions for node? Try &#96;npm i --save-dev @types/node&#96; and then add 'node' to the types field in your tsconfig. |
| stage3/drivers/parser/native-node-binding.a | type error | stage3/drivers/parser/native-node-binding.a:3:20: error TS2591: Cannot find name 'require'. Do you need to install type definitions for node? Try &#96;npm i --save-dev @types/node&#96; and then add 'node' to the types field in your tsconfig. |
| stage3/drivers/parser/native-type-only-map-like.a | refused | adamic: /workspace/adamic/stage3/drivers/parser/native-type-only-map-like.a:4:5: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added |
| stage3/drivers/parser/range-cases.a | type error | stage3/drivers/parser/range-cases.a:13:24: error TS7006: Parameter 'child' implicitly has an 'any' type. |
| stage3/drivers/parser/range-check.a | refused | adamic: /workspace/adamic/stage3/drivers/parser/range-check.a:8:13: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| stage3/drivers/scanner/main.a | type error | stage3/drivers/scanner/main.a:13:25: error TS7006: Parameter 'message' implicitly has an 'any' type. |
| stage3/drivers/scanner/probes/assert-failure.a | type error | stage3/drivers/scanner/probes/assert-failure.a:2:31: error TS2307: Cannot find module './adapted/src/compiler/scanner.ts' or its corresponding type declarations. |
| stage3/drivers/scanner/probes/assert-unknown.a | refused | adamic: /workspace/adamic/stage3/drivers/scanner/probes/assert-unknown.a:1:39: Adamic 0.1 refuses a type predicate whose return is not proven (asserts cond needs a boolean parameter); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-p |
| stage3/drivers/scanner/probes/capture-stack-marker-control.a | type error | stage3/drivers/scanner/probes/capture-stack-marker-control.a:2:21: error TS2591: Cannot find name 'node:util'. Do you need to install type definitions for node? Try &#96;npm i --save-dev @types/node&#96; and then add 'node' to the types field in your tsconfig. |
| stage3/drivers/scanner/probes/capture-stack-marker.a | type error | stage3/drivers/scanner/probes/capture-stack-marker.a:2:21: error TS2591: Cannot find name 'node:util'. Do you need to install type definitions for node? Try &#96;npm i --save-dev @types/node&#96; and then add 'node' to the types field in your tsconfig. |
| stage3/drivers/scanner/probes/discovery-captured-literal.a | refused | adamic: /workspace/adamic/stage3/drivers/scanner/probes/discovery-captured-literal.a:2:24: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| stage3/drivers/scanner/probes/discovery-explicit-undefined-method.a | refused | adamic: /workspace/adamic/stage3/drivers/scanner/probes/discovery-explicit-undefined-method.a:8:63: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| stage3/drivers/scanner/probes/discovery-optional-method-trace.a | refused | adamic: /workspace/adamic/stage3/drivers/scanner/probes/discovery-optional-method-trace.a:9:63: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| stage3/drivers/scanner/probes/discovery-optional-method.a | refused | adamic: /workspace/adamic/stage3/drivers/scanner/probes/discovery-optional-method.a:8:63: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| stage3/drivers/scanner/probes/discovery-property.a | refused | adamic: /workspace/adamic/stage3/drivers/scanner/probes/discovery-property.a:1:64: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| stage3/drivers/scanner/probes/error-constructor-value.a | refused | adamic: /workspace/adamic/stage3/drivers/scanner/probes/error-constructor-value.a:2:6: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adami |
| stage3/drivers/scanner/probes/index-signature-type-only.a | refused | adamic: /workspace/adamic/stage3/drivers/scanner/probes/index-signature-type-only.a:3:5: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added |
| stage3/drivers/scanner/probes/scanner-captured-token-value.a | refused | adamic: /workspace/adamic/stage3/drivers/scanner/probes/scanner-captured-token-value.a:2:19: Adamic 0.1 refuses a definite assignment assertion !; remove ! and initialize it where it is declared or in the constructor, or type it T &#124; undefined |
| stage3/drivers/scanner/probes/scanner-lazy-text.a | refused | adamic: /workspace/adamic/stage3/drivers/scanner/probes/scanner-lazy-text.a:3:16: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| stage3/drivers/scanner/probes/scanner-state.a | type error | stage3/drivers/scanner/probes/scanner-state.a:12:28: error TS2554: Expected 1 arguments, but got 2. |
| stage3/drivers/scanner/probes/uninitialized-non-null.a | refused | adamic: /workspace/adamic/stage3/drivers/scanner/probes/uninitialized-non-null.a:2:20: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| stage3/fixtures/assertions/01_map_call.a | refused | adamic: /workspace/adamic/stage3/fixtures/assertions/01_map_call.a:7:16: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| stage3/fixtures/assertions/02_code_point_call.a | refused | adamic: /workspace/adamic/stage3/fixtures/assertions/02_code_point_call.a:7:12: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| stage3/fixtures/assertions/03_exports_field.a | refused | adamic: /workspace/adamic/stage3/fixtures/assertions/03_exports_field.a:9:16: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| stage3/fixtures/assertions/04_field_then_call.a | refused | adamic: /workspace/adamic/stage3/fixtures/assertions/04_field_then_call.a:11:35: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| stage3/fixtures/assertions/05_regex_call_index.a | refused | adamic: /workspace/adamic/stage3/fixtures/assertions/05_regex_call_index.a:8:21: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| stage3/fixtures/assertions/06_memoize_clear.a | refused | adamic: /workspace/adamic/stage3/fixtures/assertions/06_memoize_clear.a:10:24: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| stage3/fixtures/assertions/07_indexed_operation.a | refused | adamic: /workspace/adamic/stage3/fixtures/assertions/07_indexed_operation.a:40:24: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| stage3/fixtures/assertions/08_optional_start.a | refused | adamic: /workspace/adamic/stage3/fixtures/assertions/08_optional_start.a:11:52: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| stage3/fixtures/assertions/09_interface_kind.a | refused | adamic: /workspace/adamic/stage3/fixtures/assertions/09_interface_kind.a:15:10: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-un |
| stage3/fixtures/assertions/11_parenthesized_kind.a | refused | adamic: /workspace/adamic/stage3/fixtures/assertions/11_parenthesized_kind.a:15:14: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/n |
| stage3/fixtures/assertions/12_union_target.a | refused | adamic: /workspace/adamic/stage3/fixtures/assertions/12_union_target.a:19:13: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unch |
| stage3/fixtures/assertions/13_flag_downcast.a | refused | adamic: /workspace/adamic/stage3/fixtures/assertions/13_flag_downcast.a:11:52: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unc |
| stage3/fixtures/assertions/14_structural_cache.a | refused | adamic: /workspace/adamic/stage3/fixtures/assertions/14_structural_cache.a:11:17: Adamic 0.1 refuses an unproven relation from Type to IterableOrIteratorType: optional field iterationTypes has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object w |
| stage3/fixtures/assertions/15_mutable_view.a | refused | adamic: /workspace/adamic/stage3/fixtures/assertions/15_mutable_view.a:9:6: Adamic 0.1 refuses a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the typ |
| stage3/fixtures/assertions/16_unknown_chain.a | refused | adamic: /workspace/adamic/stage3/fixtures/assertions/16_unknown_chain.a:6:50: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unch |
| stage3/fixtures/assertions/17_brand_upcast.a | refused | adamic: /workspace/adamic/stage3/fixtures/assertions/17_brand_upcast.a:14:70: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unch |
| stage3/fixtures/assertions/19_identifier_kind.a | refused | adamic: /workspace/adamic/stage3/fixtures/assertions/19_identifier_kind.a:15:24: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-u |
| stage3/fixtures/assertions/20_type_flag_mask.a | refused | adamic: /workspace/adamic/stage3/fixtures/assertions/20_type_flag_mask.a:11:54: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-un |
| stage3/fixtures/cycles/01_call_time/_namespaces/ts.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/01_call_time/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/01_call_time/core.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/01_call_time/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/01_call_time/main.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/01_call_time/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/01_call_time/path.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/01_call_time/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/02_reverse_barrel/_namespaces/ts.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/02_reverse_barrel/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/02_reverse_barrel/core.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/02_reverse_barrel/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/02_reverse_barrel/main.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/02_reverse_barrel/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/02_reverse_barrel/path.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/02_reverse_barrel/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/03_extensions/_namespaces/ts.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/03_extensions/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/03_extensions/core.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/03_extensions/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/03_extensions/main.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/03_extensions/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/03_extensions/path.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/03_extensions/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/04_directory_callback/_namespaces/ts.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/04_directory_callback/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/04_directory_callback/core.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/04_directory_callback/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/04_directory_callback/main.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/04_directory_callback/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/04_directory_callback/path.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/04_directory_callback/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/07_namespace_barrel/_namespaces/ts.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/07_namespace_barrel/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/07_namespace_barrel/_namespaces/ts.path.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/07_namespace_barrel/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/07_namespace_barrel/core.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/07_namespace_barrel/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/07_namespace_barrel/main.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/07_namespace_barrel/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/07_namespace_barrel/path.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/07_namespace_barrel/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/08_named_reexports/_namespaces/ts.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/08_named_reexports/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/08_named_reexports/core.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/08_named_reexports/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/08_named_reexports/main.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/08_named_reexports/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/cycles/08_named_reexports/path.a | refused | adamic: /workspace/adamic/stage3/fixtures/cycles/08_named_reexports/core.a:10:9: Adamic 0.1 refuses a boolean &#124; undefined as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/enums/06_set_node_flags.a | refused | adamic: /workspace/adamic/stage3/fixtures/enums/06_set_node_flags.a:79:10: Adamic 0.1 refuses a value of type T seen as Mutable<T>, which can write T["flags"] where NodeFlags is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([.. |
| stage3/fixtures/enums/08_debug_format_enum.a | type error | stage3/fixtures/enums/08_debug_format_enum.a:16:42: error TS2532: Object is possibly 'undefined'. |
| stage3/fixtures/enums/11_format_syntax_kind.a | type error | stage3/fixtures/enums/11_format_syntax_kind.a:14:42: error TS2532: Object is possibly 'undefined'. |
| stage3/fixtures/host/01_readFile_utf8.a | type error | stage3/fixtures/host/01_readFile_utf8.a:47:21: error TS2345: Argument of type 'string &#124; undefined' is not assignable to parameter of type 'string'. |
| stage3/fixtures/host/02_readFile_utf16le.a | type error | stage3/fixtures/host/02_readFile_utf16le.a:47:21: error TS2345: Argument of type 'string &#124; undefined' is not assignable to parameter of type 'string'. |
| stage3/fixtures/host/03_readFile_utf16be.a | type error | stage3/fixtures/host/03_readFile_utf16be.a:47:21: error TS2345: Argument of type 'string &#124; undefined' is not assignable to parameter of type 'string'. |
| stage3/fixtures/host/04_readFile_missing.a | type error | stage3/fixtures/host/04_readFile_missing.a:46:17: error TS2345: Argument of type 'string &#124; undefined' is not assignable to parameter of type 'string'. |
| stage3/fixtures/host/05_writeFile.a | type error | stage3/fixtures/host/05_writeFile.a:37:35: error TS2503: Cannot find namespace 'NodeJS'. |
| stage3/fixtures/host/16_getEnvironmentVariable.a | type error | stage3/fixtures/host/16_getEnvironmentVariable.a:18:17: error TS2345: Argument of type 'string &#124; undefined' is not assignable to parameter of type 'string'. |
| stage3/fixtures/host/23_newLine.a | type error | stage3/fixtures/host/23_newLine.a:13:17: error TS2345: Argument of type 'string &#124; undefined' is not assignable to parameter of type 'string'. |
| stage3/fixtures/host/25_readDirectory.a | type error | stage3/fixtures/host/25_readDirectory.a:1001:23: error TS2345: Argument of type 'T &#124; undefined' is not assignable to parameter of type 'T'. |
| stage3/fixtures/namespaces/01_builder_state.a | refused | adamic: /workspace/adamic/stage3/fixtures/namespaces/01_builder_state.a:10:1: Adamic 0.1 refuses a namespace; use a module: a file of its own, with named exports |
| stage3/fixtures/namespaces/02_jsx_names.a | refused | adamic: /workspace/adamic/stage3/fixtures/namespaces/02_jsx_names.a:6:1: Adamic 0.1 refuses a namespace; use a module: a file of its own, with named exports |
| stage3/fixtures/namespaces/03_react_names.a | refused | adamic: /workspace/adamic/stage3/fixtures/namespaces/03_react_names.a:6:1: Adamic 0.1 refuses a namespace; use a module: a file of its own, with named exports |
| stage3/fixtures/namespaces/04_debug_state.a | refused | adamic: /workspace/adamic/stage3/fixtures/namespaces/04_debug_state.a:13:1: Adamic 0.1 refuses a namespace; use a module: a file of its own, with named exports |
| stage3/fixtures/namespaces/05_debug_log_merge.a | refused | adamic: /workspace/adamic/stage3/fixtures/namespaces/05_debug_log_merge.a:18:1: Adamic 0.1 refuses a namespace; use a module: a file of its own, with named exports |
| stage3/fixtures/namespaces/06_binary_expression_state.a | refused | adamic: /workspace/adamic/stage3/fixtures/namespaces/06_binary_expression_state.a:11:1: Adamic 0.1 refuses a namespace; use a module: a file of its own, with named exports |
| stage3/fixtures/namespaces/07_parser_singleton.a | refused | adamic: /workspace/adamic/stage3/fixtures/namespaces/07_parser_singleton.a:6:1: Adamic 0.1 refuses a namespace; use a module: a file of its own, with named exports |
| stage3/fixtures/namespaces/08_parser_jsdoc_nested.a | refused | adamic: /workspace/adamic/stage3/fixtures/namespaces/08_parser_jsdoc_nested.a:11:1: Adamic 0.1 refuses a namespace; use a module: a file of its own, with named exports |
| stage3/fixtures/namespaces/09_incremental_parser.a | refused | adamic: /workspace/adamic/stage3/fixtures/namespaces/09_incremental_parser.a:7:1: Adamic 0.1 refuses a namespace; use a module: a file of its own, with named exports |
| stage3/fixtures/namespaces/10_tracing_escape.a | refused | adamic: /workspace/adamic/stage3/fixtures/namespaces/10_tracing_escape.a:7:1: Adamic 0.1 refuses a namespace; use a module: a file of its own, with named exports |
| stage3/fixtures/namespaces/11_status_type_only.a | refused | adamic: /workspace/adamic/stage3/fixtures/namespaces/11_status_type_only.a:36:1: Adamic 0.1 refuses a namespace; use a module: a file of its own, with named exports |
| stage3/fixtures/namespaces/12_builder_release_cache.a | refused | adamic: /workspace/adamic/stage3/fixtures/namespaces/12_builder_release_cache.a:10:1: Adamic 0.1 refuses a namespace; use a module: a file of its own, with named exports |
| stage3/fixtures/nested-functions/01_scanner_frame.a | refused | adamic: /workspace/adamic/stage3/fixtures/nested-functions/01_scanner_frame.a:49:12: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| stage3/fixtures/nested-functions/09_checker_constituent_recursion.a | refused | adamic: /workspace/adamic/stage3/fixtures/nested-functions/09_checker_constituent_recursion.a:15:94: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| stage3/fixtures/objects/01_reference_spreads.a | refused | adamic: /workspace/adamic/stage3/fixtures/objects/01_reference_spreads.a:7:111: Adamic 0.1 refuses optional property preserve in FileReference absent from structural source { resolutionMode: number; }, which can hide fields; declare preserve on the source type, or build a fresh object with known fie |
| stage3/fixtures/objects/03_resolution_cache_spreads.a | refused | adamic: /workspace/adamic/stage3/fixtures/objects/03_resolution_cache_spreads.a:13:9: Adamic 0.1 refuses a spread after the first field; spread once, first: { ...source, field: value } (adamic/single-spread) |
| stage3/fixtures/objects/04_trace_metadata.a | refused | adamic: /workspace/adamic/stage3/fixtures/objects/04_trace_metadata.a:13:67: Adamic 0.1 refuses a spread after the first field; spread once, first: { ...source, field: value } (adamic/single-spread) |
| stage3/fixtures/objects/05_polling_array_metadata.a | refused | adamic: /workspace/adamic/stage3/fixtures/objects/05_polling_array_metadata.a:12:23: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/ |
| stage3/fixtures/objects/08_loop_state.a | refused | adamic: /workspace/adamic/stage3/fixtures/objects/08_loop_state.a:13:17: Adamic 0.1 refuses a string as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/objects/10_build_options.a | refused | adamic: /workspace/adamic/stage3/fixtures/objects/10_build_options.a:14:54: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added |
| stage3/fixtures/objects/14_comment_pending.a | refused | adamic: /workspace/adamic/stage3/fixtures/objects/14_comment_pending.a:10:19: Adamic 0.1 refuses a definite assignment assertion !; remove ! and initialize it where it is declared or in the constructor, or type it T &#124; undefined |
| stage3/fixtures/objects/15_accessor_absence.a | refused | adamic: /workspace/adamic/stage3/fixtures/objects/15_accessor_absence.a:17:22: Adamic 0.1 refuses a definite assignment assertion !; remove ! and initialize it where it is declared or in the constructor, or type it T &#124; undefined |
| stage3/fixtures/objects/19_identifier_multimap.a | refused | adamic: /workspace/adamic/stage3/fixtures/objects/19_identifier_multimap.a:20:13: Adamic 0.1 refuses a value as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/objects/20_queue_optional_call.a | refused | adamic: /workspace/adamic/stage3/fixtures/objects/20_queue_optional_call.a:4:41: Adamic 0.1 refuses a value of type T[] seen as (T &#124; undefined)[], which can write T &#124; undefined where T is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the  |
| stage3/fixtures/objects/22_nested_optional_calls.a | refused | adamic: /workspace/adamic/stage3/fixtures/objects/22_nested_optional_calls.a:16:33: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/n |
| stage3/fixtures/objects/25_map_generator.a | refused | adamic: /workspace/adamic/stage3/fixtures/objects/25_map_generator.a:3:1: Adamic 0.1 refuses a generator function; use an explicit iterator object; suspended frames need ownership and cancellation rules before generators can be compiled without a collector (docs/user-iterators.md) |
| stage3/fixtures/objects/26_diagnostic_in.a | refused | adamic: /workspace/adamic/stage3/fixtures/objects/26_diagnostic_in.a:21:53: Adamic 0.1 refuses in; an object's shape is known; use a discriminant, or a Map |
| stage3/fixtures/objects/27_delete_substitution.a | refused | adamic: /workspace/adamic/stage3/fixtures/objects/27_delete_substitution.a:12:23: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added |
| stage3/fixtures/objects/28_debugger.a | refused | adamic: /workspace/adamic/stage3/fixtures/objects/28_debugger.a:7:9: Adamic 0.1 refuses debugger; remove it |
| stage3/fixtures/objects/30_host_optional_method.a | refused | adamic: /workspace/adamic/stage3/fixtures/objects/30_host_optional_method.a:8:73: Adamic 0.1 refuses optional property getCompilerHost in ResolutionCacheHost absent from structural source { name: string; }, which can hide fields; declare getCompilerHost on the source type, or build a fresh object wi |
| stage3/fixtures/predicates/01_identifier.a | refused | adamic: /workspace/adamic/stage3/fixtures/predicates/01_identifier.a:23:5: Adamic 0.1 refuses a type predicate whose return is not proven (return expression is not a trusted check on node); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no- |
| stage3/fixtures/predicates/02_module_name.a | refused | adamic: /workspace/adamic/stage3/fixtures/predicates/02_module_name.a:23:5: Adamic 0.1 refuses a type predicate whose return is not proven (return expression is not a trusted check on node); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no |
| stage3/fixtures/predicates/03_void_zero.a | refused | adamic: /workspace/adamic/stage3/fixtures/predicates/03_void_zero.a:23:5: Adamic 0.1 refuses a type predicate whose return is not proven (return expression is not a trusted check on node); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-t |
| stage3/fixtures/predicates/04_literal_or.a | refused | adamic: /workspace/adamic/stage3/fixtures/predicates/04_literal_or.a:24:5: Adamic 0.1 refuses a type predicate whose return is not proven (return expression is not a trusted check on node); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no- |
| stage3/fixtures/predicates/05_kind_switch.a | refused | adamic: /workspace/adamic/stage3/fixtures/predicates/05_kind_switch.a:21:5: Adamic 0.1 refuses a type predicate whose return is not proven (return paths through KindSwitchStatement are not verified); inline the check where you use it, or return a discriminant comparison on the unmodified parameter ( |
| stage3/fixtures/predicates/06_signed_numeric.a | refused | adamic: /workspace/adamic/stage3/fixtures/predicates/06_signed_numeric.a:23:5: Adamic 0.1 refuses a type predicate whose return is not proven (return expression is not a trusted check on node); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic |
| stage3/fixtures/predicates/08_flags.a | refused | adamic: /workspace/adamic/stage3/fixtures/predicates/08_flags.a:9:5: Adamic 0.1 refuses a type predicate whose return is not proven (return expression is not a trusted check on symbol); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type |
| stage3/fixtures/predicates/09_assert_defined.a | refused | adamic: /workspace/adamic/stage3/fixtures/predicates/09_assert_defined.a:8:127: Adamic 0.1 refuses a type predicate whose return is not proven (normal return has not narrowed value to NonNullable<T>); inline the check where you use it, or return a discriminant comparison on the unmodified parameter  |
| stage3/fixtures/predicates/10_nullish.a | refused | adamic: /workspace/adamic/stage3/fixtures/predicates/10_nullish.a:6:5: Adamic 0.1 refuses a type predicate whose return is not proven (the body's true narrowing does not match null &#124; undefined); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adami |
| stage3/fixtures/predicates/11_empty_assert.a | refused | adamic: /workspace/adamic/stage3/fixtures/predicates/11_empty_assert.a:6:42: Adamic 0.1 refuses a type predicate whose return is not proven (there is no body proving this parameter); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-pr |
| stage3/fixtures/records/01_has_property.a | refused | adamic: /workspace/adamic/stage3/fixtures/records/01_has_property.a:5:5: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added |
| stage3/fixtures/records/02_get_property.a | refused | adamic: /workspace/adamic/stage3/fixtures/records/02_get_property.a:5:5: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added |
| stage3/fixtures/records/03_own_keys.a | refused | adamic: /workspace/adamic/stage3/fixtures/records/03_own_keys.a:5:5: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added |
| stage3/fixtures/records/04_own_values.a | type error | stage3/fixtures/records/04_own_values.a:12:25: error TS2345: Argument of type 'T &#124; undefined' is not assignable to parameter of type 'T'. |
| stage3/fixtures/records/05_integer_order.a | refused | adamic: /workspace/adamic/stage3/fixtures/records/05_integer_order.a:5:5: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added |
| stage3/fixtures/records/06_delete_readd.a | refused | adamic: /workspace/adamic/stage3/fixtures/records/06_delete_readd.a:6:5: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added |
| stage3/fixtures/records/07_optional_view.a | refused | adamic: /workspace/adamic/stage3/fixtures/records/07_optional_view.a:5:5: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added |
| stage3/fixtures/records/08_option_boolean.a | type error | stage3/fixtures/records/08_option_boolean.a:66:79: error TS2345: Argument of type 'string &#124; undefined' is not assignable to parameter of type 'string'. |
| stage3/fixtures/records/09_option_undefined.a | type error | stage3/fixtures/records/09_option_undefined.a:66:79: error TS2345: Argument of type 'string &#124; undefined' is not assignable to parameter of type 'string'. |
| stage3/fixtures/records/10_option_wrong_type.a | type error | stage3/fixtures/records/10_option_wrong_type.a:66:79: error TS2345: Argument of type 'string &#124; undefined' is not assignable to parameter of type 'string'. |
| stage3/fixtures/records/11_option_string.a | type error | stage3/fixtures/records/11_option_string.a:66:79: error TS2345: Argument of type 'string &#124; undefined' is not assignable to parameter of type 'string'. |
| stage3/fixtures/records/12_nested_entries.a | refused | adamic: /workspace/adamic/stage3/fixtures/records/12_nested_entries.a:7:5: Adamic 0.1 refuses a value of type Map<string, Map<string, string[]> &#124; Map<string, never[]> &#124; Map<string, string[] &#124; never[]>> seen as ScriptTargetFeatures, which can write string where never is read; make the wider type read |
| stage3/fixtures/records/13_strict_option.a | refused | adamic: /workspace/adamic/stage3/fixtures/records/13_strict_option.a:4:72: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added |
| stage3/fixtures/records/14_group_by.a | type error | stage3/fixtures/records/14_group_by.a:10:24: error TS2345: Argument of type 'T &#124; undefined' is not assignable to parameter of type 'T'. |
| stage3/fixtures/records/15_inherited_read.a | type error | stage3/fixtures/records/15_inherited_read.a:10:24: error TS2345: Argument of type 'T &#124; undefined' is not assignable to parameter of type 'T'. |
| stage3/fixtures/records/16_built_strings.a | refused | adamic: /workspace/adamic/stage3/fixtures/records/16_built_strings.a:5:5: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added |
| stage3/fixtures/records/17_compare_missing_object.a | type error | stage3/fixtures/records/17_compare_missing_object.a:27:17: error TS2345: Argument of type 'boolean' is not assignable to parameter of type 'string'. |
| stage3/fixtures/records/18_compare_missing_scalar.a | type error | stage3/fixtures/records/18_compare_missing_scalar.a:27:17: error TS2345: Argument of type 'boolean' is not assignable to parameter of type 'string'. |
| stage3/fixtures/records/19_compare_empty_objects.a | type error | stage3/fixtures/records/19_compare_empty_objects.a:27:17: error TS2345: Argument of type 'boolean' is not assignable to parameter of type 'string'. |
| stage3/fixtures/records/input-fixtures/05_paths-own_constructor/main.a | type error | stage3/fixtures/records/input-fixtures/05_paths-own_constructor/main.a:1:23: error TS2307: Cannot find module 'constructor' or its corresponding type declarations. |
| stage3/fixtures/records/input-fixtures/06_paths-own_toString/main.a | type error | stage3/fixtures/records/input-fixtures/06_paths-own_toString/main.a:1:23: error TS2307: Cannot find module 'toString' or its corresponding type declarations. |
| stage3/fixtures/records/input-fixtures/07_paths-own_hasOwnProperty/main.a | type error | stage3/fixtures/records/input-fixtures/07_paths-own_hasOwnProperty/main.a:1:23: error TS2307: Cannot find module 'hasOwnProperty' or its corresponding type declarations. |
| stage3/fixtures/records/input-fixtures/08_paths-own___proto__/main.a | type error | stage3/fixtures/records/input-fixtures/08_paths-own___proto__/main.a:1:23: error TS2307: Cannot find module '__proto__' or its corresponding type declarations. |
| stage3/fixtures/records/input-fixtures/09_exports-missing_constructor/main.a | type error | stage3/fixtures/records/input-fixtures/09_exports-missing_constructor/main.a:1:23: error TS2307: Cannot find module 'pkg' or its corresponding type declarations. |
| stage3/fixtures/records/input-fixtures/10_exports-missing_toString/main.a | type error | stage3/fixtures/records/input-fixtures/10_exports-missing_toString/main.a:1:23: error TS2307: Cannot find module 'pkg' or its corresponding type declarations. |
| stage3/fixtures/records/input-fixtures/11_exports-missing_hasOwnProperty/main.a | type error | stage3/fixtures/records/input-fixtures/11_exports-missing_hasOwnProperty/main.a:1:23: error TS2307: Cannot find module 'pkg' or its corresponding type declarations. |
| stage3/fixtures/records/input-fixtures/12_exports-missing___proto__/main.a | type error | stage3/fixtures/records/input-fixtures/12_exports-missing___proto__/main.a:1:23: error TS2307: Cannot find module 'pkg' or its corresponding type declarations. |
| stage3/fixtures/records/input-fixtures/13_exports-own_constructor/main.a | type error | stage3/fixtures/records/input-fixtures/13_exports-own_constructor/main.a:1:23: error TS2307: Cannot find module 'pkg' or its corresponding type declarations. |
| stage3/fixtures/records/input-fixtures/14_exports-own_toString/main.a | type error | stage3/fixtures/records/input-fixtures/14_exports-own_toString/main.a:1:23: error TS2307: Cannot find module 'pkg' or its corresponding type declarations. |
| stage3/fixtures/records/input-fixtures/15_exports-own_hasOwnProperty/main.a | type error | stage3/fixtures/records/input-fixtures/15_exports-own_hasOwnProperty/main.a:1:23: error TS2307: Cannot find module 'pkg' or its corresponding type declarations. |
| stage3/fixtures/records/input-fixtures/16_exports-own___proto__/main.a | type error | stage3/fixtures/records/input-fixtures/16_exports-own___proto__/main.a:1:23: error TS2307: Cannot find module 'pkg' or its corresponding type declarations. |
| stage3/fixtures/records/input-fixtures/17_typesVersions-own_constructor/main.a | type error | stage3/fixtures/records/input-fixtures/17_typesVersions-own_constructor/main.a:1:23: error TS2307: Cannot find module 'pkg/constructor' or its corresponding type declarations. |
| stage3/fixtures/records/input-fixtures/18_typesVersions-own_toString/main.a | type error | stage3/fixtures/records/input-fixtures/18_typesVersions-own_toString/main.a:1:23: error TS2307: Cannot find module 'pkg/toString' or its corresponding type declarations. |
| stage3/fixtures/records/input-fixtures/19_typesVersions-own_hasOwnProperty/main.a | type error | stage3/fixtures/records/input-fixtures/19_typesVersions-own_hasOwnProperty/main.a:1:23: error TS2307: Cannot find module 'pkg/hasOwnProperty' or its corresponding type declarations. |
| stage3/fixtures/records/input-fixtures/20_typesVersions-own___proto__/main.a | type error | stage3/fixtures/records/input-fixtures/20_typesVersions-own___proto__/main.a:1:23: error TS2307: Cannot find module 'pkg/__proto__' or its corresponding type declarations. |
| stage3/fixtures/taste/04_jsx_runtime.a | refused | adamic: /workspace/adamic/stage3/fixtures/taste/04_jsx_runtime.a:6:12: Adamic 0.1 refuses a string as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/taste/06_localized_message.a | refused | adamic: /workspace/adamic/stage3/fixtures/taste/06_localized_message.a:4:36: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added |
| stage3/fixtures/taste/08_for_each.a | type error | stage3/fixtures/taste/08_for_each.a:7:37: error TS2345: Argument of type 'T &#124; undefined' is not assignable to parameter of type 'T'. |
| stage3/fixtures/taste/09_literal_cache.a | refused | adamic: /workspace/adamic/stage3/fixtures/taste/09_literal_cache.a:10:122: Adamic 0.1 refuses the comma operator; write each expression as its own statement |
| stage3/fixtures/taste/10_global_import_meta.a | refused | adamic: /workspace/adamic/stage3/fixtures/taste/10_global_import_meta.a:11:45: Adamic 0.1 refuses &#124;&#124;=; write the if |
| stage3/fixtures/taste/12_symbol_links.a | refused | adamic: /workspace/adamic/stage3/fixtures/taste/12_symbol_links.a:10:13: Adamic 0.1 refuses a number as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/taste/13_void_callback.a | refused | adamic: /workspace/adamic/stage3/fixtures/taste/13_void_callback.a:5:62: Adamic 0.1 refuses the void operator; evaluate the expression as a statement |
| stage3/fixtures/taste/16_scan_exclamation.a | refused | adamic: /workspace/adamic/stage3/fixtures/taste/16_scan_exclamation.a:14:44: Adamic 0.1 refuses the comma operator; write each expression as its own statement |
| stage3/fixtures/taste/17_binder_flow.a | refused | adamic: /workspace/adamic/stage3/fixtures/taste/17_binder_flow.a:36:24: Adamic 0.1 refuses &#124;&#124;=; write the if |
| stage3/fixtures/taste/21_truthy_loops.a | refused | adamic: /workspace/adamic/stage3/fixtures/taste/21_truthy_loops.a:31:13: Adamic 0.1 refuses a number as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/fixtures/taste/22_assignment_once.a | refused | adamic: /workspace/adamic/stage3/fixtures/taste/22_assignment_once.a:20:22: Adamic 0.1 refuses &&=; write the if |
| stage3/fixtures/taste/24_proportional_conditions.a | refused | adamic: /workspace/adamic/stage3/fixtures/taste/24_proportional_conditions.a:34:16: Adamic 0.1 refuses a string as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| stage3/ledger/checker-259/witnesses/callback.a | type error | stage3/ledger/checker-259/witnesses/callback.a:2:53: error TS18048: 'item' is possibly 'undefined'. |
| stage3/ledger/checker-259/witnesses/entries.a | type error | stage3/ledger/checker-259/witnesses/entries.a:2:12: error TS2488: Type '[string, number] &#124; undefined' must have a '[Symbol.iterator]()' method that returns an iterator. |
| stage3/ledger/checker-259/witnesses/iterable.a | type error | stage3/ledger/checker-259/witnesses/iterable.a:2:7: error TS2322: Type '(string &#124; undefined)[]' is not assignable to type 'string[]'. |
| stage3/ledger/checker-259/witnesses/nested-entries.a | type error | stage3/ledger/checker-259/witnesses/nested-entries.a:3:12: error TS2488: Type '[string, Map<string, { name: string; }>] &#124; undefined' must have a '[Symbol.iterator]()' method that returns an iterator. |
| stage3/ledger/checker-259/witnesses/set-copy.a | type error | stage3/ledger/checker-259/witnesses/set-copy.a:2:7: error TS2322: Type 'Set<string &#124; undefined>' is not assignable to type 'Set<string>'. |
| stage3/ledger/checker-259/witnesses/set-shape.a | type error | stage3/ledger/checker-259/witnesses/set-shape.a:2:7: error TS2740: Type 'Omit<Set<string>, "difference" &#124; "intersection" &#124; "isDisjointFrom" &#124; "isSubsetOf" &#124; "isSupersetOf" &#124; "symmetricDifference" &#124; "union">' is missing the following properties from type 'Set<string>': union, intersection, difference |

Unclassified: **0**. No corpus exemption is pending.
