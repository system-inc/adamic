Built: pushed next-three ownership claim and reproducible JSX blocker probes; no new rule port.
Commits: prior work through 068060e; next-three claim 69ec7cb5 pushed before probes; evidence commit is named in the final response.
Commands and outputs: Go rule tests, byte-exact parser refusal probes, owned vet and filtered uncached oracle pass; setup requires the existing scratch overlay, completing in 24s on 5 processors.
Mutants: the probe that skips parsing compiles and exits successfully, and is caught by the refusal check; there are no rule mutants because no rule implementation can reach JSX.
Not covered: these three rule ports, compiler/stage1 differential findings and fixes, findings per second, per-rule mutants and the full repository gate.

The next three names are @next/next/google-font-preconnect,
@next/next/inline-script-id and @next/next/next-script-for-ga. The first push was
already current. Fetch used --prune --recurse-submodules=no and all origin heads.
Fetched main is ef3d907ecdc4c771b016f7d9c52372def057a340. Selection inspected 305
origin refs and every unique claim Markdown blob. All 46 helper-ready names
occur in those files. Continuing in inventory array order, the syntax-only
predicate is needs_type_information=false, including waiting-on-helper entries.
selection.py reproduces the choice using 068060e for this branch's pre-claim
snapshot. Its captured output names the excluding claim for each helper rule.
Existing ports on other branches are not called main ports by this search.

The claim reserves the three directories under rules/. No incomplete rule
registration is installed. Probe sources, witnesses and logs live here under
claims/ because registry discovery refuses any rule directory without a complete
descriptor. The only new Adamic program is parser-probe.a. Witnesses are raw
TypeScript text with .ts.txt extensions. No shared parser, dispatch, oracle,
copied-file list, compiler source or cohere submodule source was changed.

Observed blocker: the current stage1 parser emits no JSX nodes. Its published
reports exclude JSX, and the positive witnesses actually fail parsing on source
Node, emitted JavaScript and ASan/UBSan native. Each exits 70 with empty stdout
and identical stderr bytes across the three backends. Failures are expected
GreaterThanToken, got Identifier at offsets 32, 68 and 34 respectively. They are
parser panics rather than sanitizer failures. Unmodified pinned Go cohere parses
these same sources as TSX and reports one finding each, with byte ranges 26..67,
61..67 and 27..33. The independent Go probe validates one finding per input.

Inference: registering JSX listeners now would produce missing findings or
parser failures, rather than the required Go findings. These rules need a JSX
parser and AST adapter first. Google font preconnect and Google Analytics also
need the inventory's JSX attribute/entity helpers; inline script id needs import
binding and JSX attribute helpers. This is a shared foundation dependency outside
this worker's territory. No regex substitute or silently empty port is delivered.
The existing .a registration/shared profile API problems remain, independently
of this additional JSX blocker.

Commands were run after source /workspace/adamic-tools/env.sh and all output was
written directly to logs. Reproduce from the repository root:

```sh
python3 stage1/cohere/lint/claims/wave1-12-next-evidence/selection.py > /tmp/next-selection.log 2>&1

go test ./stage1/cohere/lint/claims/wave1-12-next-evidence \
  -count=1 -v -timeout=10m > /tmp/next-probes.log 2>&1

ADAMIC_NEXT_PROBE_MUTANT=1 \
  go test ./stage1/cohere/lint/claims/wave1-12-next-evidence \
  -count=1 -v -timeout=10m > /tmp/next-control.log 2>&1

go vet ./stage1/cohere/lint/claims/wave1-12-next-evidence > /tmp/next-vet.log 2>&1

ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m \
  > /tmp/next-oracle.log 2>&1

(cd cohere && go test ./internal/lint/rules/next \
  -run '^(TestGoogleFontPreconnect|TestInlineScriptId|TestNextScriptForGa)' \
  -count=1 -v -timeout=10m > /tmp/next-upstream.log 2>&1)
```

The control run must fail. It changes only parser.file() to a console call in a
temporary .a copy. Both native and JavaScript artifacts compile; source Node,
emitted JavaScript and sanitized native complete normally with wrong output.
This proves the blocker check can detect acceptance. It is a probe control, not
one of the requested semantic rule mutants, and is not credited as a rule port.
An earlier control used an unsupported property-access statement and failed
lowering. It does not count. Two early harness runs used the wrong repository
root and a relative Go parser filename; neither is counted as validation.

The final byte-comparison probes pass in 15.028s. The compiling control is
caught on all nine witness/backend combinations in 16.495s, with exit 0 and
empty stderr before comparison on each. The pinned Go families pass in 0.044s. Owned vet exits 0 with no output. The
filtered uncached external oracle passes in 0.439s with native/node cache hits 0
and misses 1. No complete repository gate is claimed.

Unmodified bash cloud/setup.sh exits 1 while warming caches:
profile_test.go:32:23 cannot range over portFiles, now a function. Timings for Go,
clang, Node and submodules are all 0s; nproc prints 5. The previously proposed
compatibility overlay works around that known issue without editing shared
source. After moving probe artifacts out of rules/, this command exits 0:

```sh
GOFLAGS=-overlay=/tmp/lint-wave1-12/overlay.json bash cloud/setup.sh \
  > /tmp/next-setup-overlay.log 2>&1
```

Its lines are go ready 0s, clang ready 1s, node ready 1s, submodules ready 1s,
build cache warm 24s, done in 24s on 5 processors, cpu.max 400000 100000,
17.6 GB. The earlier overlay setup failed because the initial evidence-only rule
directories lacked descriptors; that layout was corrected and is not committed.

Findings-per-second measurements for these three ports are unavailable because
none can run its listener on the required positive input. Timing parser failure
or reporting zero findings would not measure the requested rule throughput.
