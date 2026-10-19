import collections,json,pathlib,shutil
root=pathlib.Path(__file__).resolve().parent
source=pathlib.Path('/tmp/adamic-recut4-evidence')
for file in source.iterdir():shutil.copyfile(file,root/file.name)
shutil.copyfile('/tmp/adamic-recut4-validation.log',root/'validation.log')
rows=json.loads((root/'validation.json').read_text())
assert collections.Counter(row['mode'] for row in rows)=={'trace':5,'build':21,'read':21}
assert all(row['exit']==0 and row['seconds']<60 for row in rows)
assert (root/'ir-readers.exit').read_text().strip()=='0'
assert (root/'vet.exit').read_text().strip()=='0'
assert (root/'dependency-census.exit').read_text().strip()=='0'
with (root/'REPORT.md').open('a') as report:
 report.write('\nFinal fresh verification: all 21 cmd/adamic top-level tests were selected individually in both build and read-only modes, 42 passing invocations. Each invocation has timeout 60 and -test.timeout=55s; request/header setup is included. No whole-package test command was used. Largest cmd invocation: %.3f s. Timings for every test are in validation.json and build-/read- logs. Restored and new runtime tests:\n\n| Test | Build s | Read-only s |\n|---|---:|---:|\n'%max(row['seconds'] for row in rows if row['mode']!='trace'))
 for row in rows:
  if row['mode']=='build' and 'Runtime' in row['test']:
   read=next(r for r in rows if r['mode']=='read' and r['test']==row['test'])
   report.write(f'| {row["test"]} | {row["seconds"]:.3f} | {read["seconds"]:.3f} |\n')
 report.write('\nOrdered real tracing used fresh products, oracle owners first. All five trace invocations passed below 60 seconds with zero refusals:\n\n| Product | Total traced s | Result |\n|---|---:|---|\n')
 for row in rows:
  if row['mode']=='trace':report.write(f'| {row["test"]} | {row["seconds"]:.3f} | '+('four settled' if 'NativeRuntimes' in row['test'] else 'one settled' if 'Counted' in row['test'] else 'found, zero settled')+' |\n')
 report.write('\ncheck-traces.py passed: nine observed runtime keys have one builder each; shared cmd finders read exactly the oracle settlement keys. trace-ownership.log includes their full keys. The six incidental owner-group archives remain oracle-only before and after this change: ')
 keys=[]
 for variant,base in [('Sanitized','3ac8c29ccf42'),('WASI','6e08d061bac9')]:
  data=[json.loads(line) for line in (root/f'trace-TestProduct_NativeRuntimes{variant}0.jsonl').read_text().splitlines()]
  keys += [r['nameKey'][:12] for r in data if r.get('outcome')=='settled' and not r['nameKey'].startswith(base)]
 report.write(', '.join(keys)+'.\n')
 report.write('''
All ten independently applied mutants failed for the intended test assertion or parser refusal, not a compiler warning. Source files were restored in finally, and the final binary and vet were run on restored sources. Patches are noncompilable .patch evidence.

| Mutant | Catcher / observed failure |
|---|---|
| second-owner | TestCommandRuntimeArchivesHaveOneBuilder: two producer packages [internal/oracle cmd/adamic] for the actual base sanitized name key |
| second-builder | TestSharedCommandRuntimeCannotBuildItsOwnersArchive: child builds successfully instead of failing for the missing oracle owner |
| missing-wait | TestCommandRuntimeProductsWaitForOwners: missing sanitized owner wait |
| missing-key-part | TestCommandRuntimeProductKeyHoldsEveryPart: adamic-runtime-v6 absent from identity |
| undeclared-sanitized | unchanged TestExplainChecksOutput/overload_some_empty/build: no runtime product declared |
| ignored-directory | TestCommandRuntimeCensusWalksSourceTree: ignored .cache/decoy.go is incorrectly parsed |
| fixture-directory | TestCommandRuntimeCensusWalksSourceTree: testdata/decoy.go is incorrectly parsed |
| submodule-directory | TestCommandRuntimeCensusWalksSourceTree: cohere/decoy.go is incorrectly parsed |
| missing-producer | TestCommandRuntimeCensusWalksSourceTree: ordinary producer absent from actual package list |
| git-required | real TestCommandRuntimeArchivesHaveOneBuilder with no Git PATH: exec: "git": executable file not found in $PATH |

Commands and outputs (all test output goes directly to logs):

* timeout 60 git fetch origin; git checkout -b compiler/loom-cmd-archives-4 origin/main, base 062ba8b7ec31ca855248511f80614d99c3a915a5. A final precommit fetch still names that tip. Read CLAUDE.md, docs/front-door.md, root .gitignore/.gitmodules, the restored test files and current runtimeproducts/Find/BuiltElsewhere. Binding language/memory docs and product order are unchanged from the version already read for -3. Only the same two product waits are added to main's file.
* export GOPROXY='https://proxy.golang.org|direct'; timeout 240 bash cloud/setup.sh --wasi-sdk; source /workspace/adamic-tools/env.sh. Passed, no workaround. nproc 5; cpu.max 400000 100000, four CPU share. Timing lines: Node 0.050 s, Go 0.058 s, submodules 0.155 s, markdown 0.206 s, clang 0.491 s, WASI 0.787 s, shared cache 3.571 s, go build 32.709 s, cache warm 32.946 s, done 33.043 s. Node 24.19.0, Go 1.27.1, clang 20.1.8. Full setup.log retained.
* timeout 120 go test -c -o /tmp/adamic-recut4.test ./cmd/adamic (initial compile timeout 180); timeout 120 go test -c -o /tmp/oracle4.test ./internal/oracle; timeout 90 go build -o /tmp/traced4 ./internal/buildcache/cmd/traced. Passed. Cmd binary rebuilt after mutants on final restored sources. Go compilation/toolchain installation is outside measured test leaves; per-test setup is included.
* timeout 60 /tmp/adamic-recut4.test -test.run='^TestCommandRuntime(ArchivesHaveOneBuilder|CensusWalksSourceTree)$' -test.v -test.timeout=55s from cmd/adamic: both pass, 5.96 and 0.00 s. Full-tree scan 1,354 Go files. timeout 90 python3 no-git.py: real full-tree ownership test passes with shutil.which git equal to None, 5.94 s test time. no-git.json preserves invocation time and PATH.
* timeout 450 python3 mutants.py: five existing mutants caught. Each selected go test ./cmd/adamic -run SELECTED -count=1 -timeout 60s has outer timeout 90. timeout 480 python3 walk-mutants.py: four traversal mutants caught via selected go test -timeout 55s / timeout 90; the Git-regression mutant is compiled under timeout 120 and run as a binary under timeout 60 / -test.timeout=55s with no Git in PATH. All expected failures and restores succeeded.
* timeout 540 python3 validate.py: the five traced products listed above, then every cmd Test function separately in build and ADAMIC_BUILD_CACHE=read modes. Product trace command is timeout 60 /tmp/traced4 -- BINARY -test.timeout=55s -test.v -test.run=^NAME$, with cwd internal/oracle or cmd/adamic, ADAMIC_ORACLE_WASI=1 and ADAMIC_BUILD_STORE=off. All live trace reports/logs were outside the checkout until settlement completed; no unstable-tree retry was needed. Every 21 + 21 test invocation passed. python3 check-traces.py passed its owner-key and one-builder assertions.
* timeout 90 go test ./internal/compilercensus -run '^TestDependencies(FollowTheMarkers|CoverTheDeclaredMap)$' -count=1 -timeout 60s: passed, 4.921 s, exit 0. This covers main's compiler-dependency census and cmd consumers. No new package needs declaration.
* timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s: passed, 39.908 s, exit 0. timeout 90 go vet ./cmd/adamic on final sources: passed without diagnostics, exit 0. gofmt, source diff checks and main product-order preservation checks passed.

No test leaf exceeded 60 seconds. No full packages or gate were run. No PR or Loom submission was made; the user submits the pushed branch.
''')
print('PASS: 47 validation invocations, all below 60 seconds; report completed')
