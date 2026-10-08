# react/no-this-in-sfc

Branch: `lint-rules/react-no-this-in-sfc`, from `origin/area/stage1-lint`
`6bf7bcec73a04b85900c5a69a257b3cf34b35a40`, which contains origin/main
`45487a809f89885a3fc651cd590e7dabf31362dc`. The all-origin duplicate check
found no existing descriptor before this port. Only this rule directory changes.

The oracle adapter returns the unmodified Go `react.NoThisInSfc` at cohere pin
`7945d102a6c18dd36adf9114a758ce646e8b2359`. Go takes no options, so its adapter
returns nil and no witness needs an options sidecar. Shared ES6 and strict ES5
React helpers are imported; no shared predicate is reimplemented.

## Matched behavior

The selected gate passed: 95 unique captured source/file/options cases for this
rule, including 52 JSX cases, plus three owned witnesses and the inherited
comparisons. Source Node, emitted JavaScript and ASan/UBSan native match Go's
findings, descriptions, ranges, fixes and suggestions. This rule supplies no
fixes or suggestions. Captured inputs are in
[evidence/upstream-cases.json.gz](evidence/upstream-cases.json.gz).

Go's deliberately narrow behavior is retained: JSX is unnecessary, only ASCII
uppercase component names count, anonymous function expressions do not take their
variable's name, and memo/forwardRef arrows are not recognized. Parentheses around
`this` break the immediate member-parent anchor. Go reports `a[this]` even though
`this` is the argument, not the member receiver; see upstream
`TestNoThisInSfcAcceptsAnElementAccessArgument` and our `boundaries.ts.txt` witness.
Disable directives are not processed at this rule layer, matching upstream
`TestNoThisInSfcReportsUnderADisableDirective`. Findings point at `this` alone.

Witnesses:

- [boundaries.ts.txt](testdata/boundaries.ts.txt): names, parentheses, wrappers,
  element-access arguments and nested arrows.
- [contexts.ts.txt](testdata/contexts.ts.txt): React versus plain class ancestry,
  strict versus loose ES5 factories, accessor values versus keys, and nested
  functions/static blocks.
- [keyword.tsx.txt](testdata/keyword.tsx.txt): JSX, computed access, callback and
  spread findings, including multiple findings in one component.

The mutant negates the shared ASCII component-name predicate. It
successfully compiles to emitted JavaScript, runs on Node and emitted JavaScript,
and disagrees with live Go. [evidence/mutant.log:5](evidence/mutant.log) records the
first caught comparison. No compiler or sanitizer error is credited as a kill.

## Commands and evidence

All commands run from the repository root after sourcing
`/workspace/adamic-tools/env.sh`; GOPROXY is `https://proxy.golang.org|direct`.

```sh
bash cloud/setup.sh > /tmp/s13-react-no-this-in-sfc-setup.log 2>&1
bash cloud/setup.sh --wasi-sdk > /tmp/s13-react-no-this-in-sfc-wasi-setup.log 2>&1
go run ./cmd/lint-registry > /tmp/s13-react-no-this-in-sfc-registry.log 2>&1
go test -json ./stage1/cohere/lint -run 'TestRulesAgree|TestOwnedWitnesses|TestMutants/react-no-this-in-sfc-component-name-inverted' -count=1 -timeout 40m > /tmp/s13-react-no-this-in-sfc-selected-final.jsonl 2>&1
```

Selected package PASS, 240.698s; owned witnesses PASS, 80.07s; mutant subtest
PASS, 8.54s. The selected gate includes all registered owned witnesses and the
inherited captured corpus, as the existing harness does not split those two
aggregate tests by rule. See [evidence/selected-summary.json](evidence/selected-summary.json)
and the complete [selected log](evidence/selected.jsonl.gz).

Setup took 179.704s (Go 0.085s, Node 0.082s, clang 0.428s, markdown 1.067s,
submodules 7.350s, build/cache 179.677s); the WASI-enabled rerun took 78.508s.
`nproc` is 5, cgroup CPU quota is four CPUs. Both setup logs are preserved.

The complete lint package gate runs once with the clean TypeScript v6.0.3 checkout
at `050880ce59e30b356b686bd3144efe24f875ebc8`, WASI SDK 27, and one fresh directory
shared by ADAMIC_LINT_PROFILE_DIR and ADAMIC_LINT_PROFILE_SNAPSHOTS. Benchmarks are
also enabled to avoid their input-dependent skips. The exact environment is in
[evidence/whole-inputs.json](evidence/whole-inputs.json).

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/borrow-chains-typescript \
ADAMIC_LINT_PROFILE_DIR=/tmp/s13-react-no-this-in-sfc-profile-final.UvGlDk \
ADAMIC_LINT_PROFILE_SNAPSHOTS=/tmp/s13-react-no-this-in-sfc-profile-final.UvGlDk \
ADAMIC_LINT_BENCH=1 \
WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot \
go test -json ./stage1/cohere/lint -count=1 -timeout 40m > /tmp/s13-react-no-this-in-sfc-whole-final.jsonl 2>&1
```

An initial whole-package attempt was stopped during serial fix-engine tests
after review found an exact shared ASCII predicate that the first version had
duplicated. The final rule imports that helper and its local mutant negates the
call. The final selected gate is rerun, followed by one completed full package
gate on that source. The stopped attempt is retained as
[evidence/whole-before-helper.jsonl.gz](evidence/whole-before-helper.jsonl.gz),
not credited as a full pass. The completed whole lint package gate PASSed (exit 0): **152 pass, 0 fail,
1 skip**, counting every terminal test event including subtests. Top-level counts
are **46 pass, 0 fail, 1 skip**. Package time is 1085.787s; measured wall time is
1087.641s. `nproc` is 5. Load before is 1.06/1.52/1.89; after is 5.39/5.39/4.12.

The single skip is `TestCheckerBridgeRefusalPending`, whose existing test says it
awaits `codex/tsgo-errors-as-values`: TSGoError is absent from this compiler's
prelude. All corpus, profile and benchmark inputs were supplied; no input-dependent
test skipped. This dependency belongs to another compiler unit and is not changed
here. See [evidence/whole-summary.json](evidence/whole-summary.json) and the complete
[evidence/whole.jsonl.gz](evidence/whole.jsonl.gz). Captured-rule, owned-witness,
compiler/stage1 corpus, shards, profile snapshots and all mutant tests passed. No new Go tests or shared
source changes were needed. No language gap or missing helper was encountered.
