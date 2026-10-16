Built admission-delta with its compiler dependency declaration and runtime input-isolation proof.
One source-squash commit on base d79302e50c022c997c00f7c68ddb7475f9ee65fc, branch compiler/loom-admission-delta-3.
Named tool checks PASS 0.937s; TestCallTargetReaders PASS 3.152s; vet PASS with no output.
All five source gate mutants caught, including dirty-input by TestHeadInputIsolation's runtime assertion; all restored.
Not covered: real compiler corpus admission run, full packages/full gate, later repair-script mutants.

Source squash: 865eaef99efad5d65d5061e7c389b44ab6119da3.
Queued origin/resupply/0pfxen75 was checked first at
76a6382c7ea715a2a0bd4b13af3ff879b94c0563; it lacks this package declaration.
The new entry is under packages, naming internal/load, internal/lower, internal/native,
internal/javascript, internal/ir, internal/flow and cmd/adamic. This is conservative:
cachedCompiler builds the selected revision's cmd/adamic with a lowering adapter;
main exercises lowering and both backends. It does more than run supplied executables.

TestHeadInputIsolation keeps the equal-admission lowering checks and now forces one
new admission using the fixture's refusing base compiler. The working-tree input is
changed while head still contains console.log('42'). The head compiler reads its input
before accepting, emitting JavaScript, or building native. The test requires one sampled
new admission, one Node agreement, a passing verdict and stdout 42 from all three runs.
This proves runtime comparison happened, rather than merely checking a successful command.
No new oracle fixture is registered, so counts.md has no new row to regenerate.

The old dirty-input mutant initially survived because the test accepted both revisions
and never selected runtime comparisons. With the extension, the unchanged mutant fails
at the runtime-input assertion: Admitted=1, SamplingSize=1, NodeAgreements=0, Verdict=fail.
This is an assertion kill, not a build-warning or sanitizer claim. Its exact patch and
failure log are committed. The runner restores each mutation in finally.

Commands ran at repository root after source /workspace/adamic-tools/env.sh, with output
redirected directly to the saved logs, never piped:
- timeout 180 go test ./cmd/adamic-admission-delta -run '^(TestClassification|TestOutputsAgree|TestSampling|TestDiffCoverage|TestDiffSamplingIsMandatory|TestShardPartitionKeepsDiff|TestShardCommandScope|TestNegative.*|TestCommandHelper|TestNewAdmissionMismatch|TestHeadInputIsolation|TestManifestMismatch|TestNewlyRefusedIsInformational|TestEmptyCorpus|TestCompilerTimeoutIsError|TestBuildRevisions|TestNodeDeclarationsArePinned|TestCompilerCacheIncludesSHA|TestParallelLoweringOnly)$' -count=1 -v -timeout 90s > review/compiler/loom-admission-delta-3/focus-fixed.log 2>&1
  PASS 0.937s. The edited TestHeadInputIsolation passed in 0.35s (isolated run 0.22s).
- timeout 600 python3 review/compiler/loom-admission-delta-3/run-mutants.py > review/compiler/loom-admission-delta-3/mutants-fixed.log 2>&1
  exit 0. Each child go test uses -count=1 -timeout 60s and a 90s subprocess timeout.
  crash-as-refusal: TestClassification; ignore-output: TestOutputsAgree and TestNewAdmissionMismatch;
  omit-witness: TestSampling; ignore-blob: TestManifestMismatch;
  dirty-input: TestHeadInputIsolation. Every mutant exited 1, all intended assertions failed.
- timeout 180 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s > review/compiler/loom-admission-delta-3/readers-fixed.log 2>&1
  PASS 3.152s.
- timeout 180 go vet ./cmd/adamic-admission-delta > review/compiler/loom-admission-delta-3/vet-fixed.log 2>&1
  exit 0, no output.
- git diff --check -- cmd cloud: exit 0, no output. Exact evidence patch context is retained.

Toolchain: export GOPROXY='https://proxy.golang.org|direct'; timeout 600 bash cloud/setup.sh.
Setup exit 0: Node 0.044s, Go 0.054s, Markdown dependencies 0.113s, submodules 0.176s,
clang 0.309s, shared cache 0.692s, go build 53.849s, cache warm 53.947s, done 53.992s.
nproc=5, cpu.max=400000 100000; clang 20.1.8, Go 1.27.1, Node 24.19.0.

Source test leaf seconds from the final focused run:
- TestCommandHelper: 0.00s
- TestNegativeExactBackendStop: 0.00s
- TestNegativeContentHash: 0.00s
- TestNegativeEditedWitnessMandatory: 0.00s
- TestDiffSamplingIsMandatory: 0.00s
- TestOutputsAgree: 0.00s
- TestSampling: 0.00s
- TestClassification: 0.00s
- TestEmptyCorpus: 0.08s
- TestCompilerTimeoutIsError: 0.10s
- TestDiffCoverage: 0.17s
- TestNewlyRefusedIsInformational: 0.11s
- TestManifestMismatch: 0.10s
- TestNegativeAlreadyAcceptedRepairMutant: 0.28s
- TestNewAdmissionMismatch: 0.25s
- TestNegativeValidProgramCannotBeListed: 0.00s
- TestNegativeTypeCorrectMutant: 0.27s
- TestNegativeRulingMetadata: 0.00s
- TestNegativeEmptyListPrinted: 0.21s
- TestParallelLoweringOnly: 0.00s
- TestCompilerCacheIncludesSHA: 0.00s
- TestShardPartitionKeepsDiff: 0.00s
- TestNodeDeclarationsArePinned: 0.00s
- TestHeadInputIsolation: 0.35s
- TestShardCommandScope: 0.09s
- TestNegativeCommandAccepted: 0.17s
- TestBuildRevisions: 0.93s
