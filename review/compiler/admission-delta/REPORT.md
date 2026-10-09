Built the admission-delta classifier, provenance verification, deterministic sampling, and three-runtime gate for task #936mrbr.
Classifier commit 43de6ebc was pushed first; final code is 4b4f5f66, with corpus dependency 5d044eca explicitly merged.
Focused tests pass; equal revisions report admitted: 0, verdict: pass; the lowering mutant reports admitted: 1, verdict: fail, exit 1.
Five gate mutants fail their specific checks; a real admission-plus-numeric-lowering mutant prints 41 on both backends against Node's 42.
Not covered: the full corpus execution or full gate; real cold compiler revision builds were not rerun as command tests.

This lands the compiler testing tool toward the admission-delta unit. The brief names task #936mrbr without a numbered roadmap step, so none is inferred.

The corpus generator at head produced witnesses=2, fixtures=801, gaps=107, review=6, fuzz=0. Its generator blob, manifest blob, and program blobs are in manifest.json and the result documents. The independent lowering probe lives one directory deeper than the generator's review patterns, preserving the requested corpus counts. Empty corpora are represented in JSON.

Commands and observations:

- `export GOPROXY='https://proxy.golang.org|direct'; timeout 300 bash cloud/setup.sh`: the initial cold-cache setup reached submodules-ready at 50.614 s, then hit the hard limit, exit 124. Cold build and initial guard attempts also hit their limits; their empty logs did not establish success.
- `timeout 180 bash cloud/setup.sh`: exit 0. Go ready 0.043 s, Node ready 0.036 s, submodules ready 0.186 s, markdown ready 0.194 s, clang ready 0.454 s, Go build ready 46.290 s, build cache warm 46.437 s, done 46.472 s. Environment sourced from `/workspace/adamic-tools/env.sh`. `nproc`: 5, cgroup quota: 4 CPUs.
- `timeout 60 go test ./cmd/adamic-admission-delta -count=1 -v -timeout 60s`: exit 0, package 0.295 s. Tests exercise fake compiler command contracts, including revision preparation, and the real Node process for the command-level mismatch test.
- `timeout 90 go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 60s`: exit 0, as recorded in call-targets.log.
- `timeout 180 python3 review/compiler/admission-delta/run-mutants.py`: exit 0, every individual Go test command exited 1 for its mutant. Each command has `-timeout 60s`, with a 90 s outer bound. Sources are restored in finally.
- `timeout 10 python3 cloud/admission-corpus/manifest.py --sha HEAD`: exit 0; counts above. Full pinned manifest is retained.
- `timeout 60 /tmp/adamic-admission-delta --base HEAD --head HEAD --base-binary /tmp/admission-adamic --head-binary /tmp/admission-adamic --corpus internal/flow/testdata --manifest-generator cloud/admission-corpus/manifest.py --json`: exit 0, admitted: 0, verdict: pass, two real flow corpus programs. Result source revision: 4b4f5f666457537727cf2c96279476363826984b.
- `timeout 60 /tmp/adamic-admission-delta --base HEAD --head a682028f87db906bccd2fa95b861b12b3069a44b --base-binary /tmp/admission-adamic --head-binary /tmp/admission-mutant-adamic --manifest review/compiler/admission-delta/mutant-manifest.json --manifest-generator cloud/admission-corpus/manifest.py --budget 0.1 --json`: exit 1, admitted: 1, sampled witness count 1, omitted 0, agree false. Node exit 0/stdout `42\n`; JavaScript exit 0/stdout `41\n`; native release exit 0/stdout `41\n`. This disagreement comes from lowering, without compiler warnings or linker failure.
- Integration's prescribed fetch and lane-check command: exit 0, gofmt and tools on three Go files, t.Parallel on one test package, vet one package. Exact output is in lane-final.log.

Mutants and catchers:

- crash-as-refusal: TestClassification directly checks that crashes/timeouts are compiler errors.
- ignore-output: TestOutputsAgree and TestNewAdmissionMismatch reject ignored bytes and incorrect success exits.
- omit-witness: TestSampling and TestNewAdmissionMismatch reject dropping witnesses under the budget.
- ignore-blob: TestManifestMismatch rejects an altered program blob.
- dirty-input: TestHeadInputIsolation rejects using working-tree source instead of the head snapshot.
- unsupported-view plus wrong numeric lowering: the real compiler's dictionary read refusal is removed, and numeric literal 42 lowers as 41. The source Node oracle catches the two backends agreeing on that wrong answer. The complete patch is unsupported-view.patch; apply it to 3a22a8e3 to reproduce the isolated head's source changes. No mutant compiler sources are Go files under review.

Each new test leaf:

- TestCommandHelper: 0.00 s
- TestClassification: 0.00 s
- TestOutputsAgree: 0.00 s
- TestSampling: 0.00 s
- TestNewlyRefusedIsInformational: 0.05 s
- TestEmptyCorpus: 0.05 s
- TestHeadInputIsolation: 0.09 s
- TestCompilerTimeoutIsError: 0.10 s
- TestManifestMismatch: 0.05 s
- TestNewAdmissionMismatch: 0.13 s
- TestBuildRevisions: 0.25 s

Assumptions and limits:

- The explicit instruction to merge 5d044eca is followed as the corpus dependency exception to the general worker-branch merge restriction. Current origin/main was fetched and was already an ancestor of the final delivery code.
- Supplied binaries require their corresponding explicit source revisions; the tool cannot prove a caller's binary came from that revision. The real proof uses separately built binaries and records the replayable head mutation.
- Budget selection reserves five per-command timeouts per program. Compiler preparation and admission classification are outside that runtime budget. Witnesses can exceed the budget. A sampled result returns verdict sampled and reports omissions; it does not claim a complete proof.
- The revision-build test uses a tiny compiler to verify checkout and build mechanics. The real compiler proof uses prebuilt binaries. No full corpus execution, whole package gate, sanitizer run, or integration push-main modification was performed.
- No oracle fixture was added, so counts.md has no new fixture row to refresh. The planted probe is review evidence.

All logs, results, manifests, and mutant patches are under review/compiler/admission-delta/. Pushes target compiler/admission-delta only; no pull request is opened.
