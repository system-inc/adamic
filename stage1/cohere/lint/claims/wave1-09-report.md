Built: published the three-rule claim and reproducible JSX blocker evidence; no rules ported.
Commits: registration merge 5bfd332; pre-code claim 1b79ffa; evidence commit follows this report.
Commands and outputs: setup passed in 109s, nproc 5; Go rule suites passed; all three JSX probes exit 70 on all backends.
Mutants: no new rule mutants; existing registry rejection controls passed, with no semantic-parity credit.
Not covered: rule findings/fix parity, rule mutants, corpora, throughput and full gate; helper merge remains blocked.

Assigned rules, from the ordered handoff linked by helpers/REPORT.md, are positions
25 `react/jsx-no-useless-fragment`, 26 `react/no-invalid-html-attribute`, and
27 `react/no-unescaped-entities`. The ordered names live in HELPERS.md and
helpers/readiness.json on the helper branch, rather than in REPORT.md itself.
A search of source/registration candidates in stage1/cohere/lint across every
fetched origin ref found none of these names outside helper/inventory/testdata
artifacts. No rule is skipped as already ported. This is a bounded source search,
not proof about unpushed work or differently named implementations.

Observed foundation state:

- Started from origin/main `d090af531216ddd3c25a0dede6b82d7c0a6edf76`.
- Merged registration `48ecd9302bf3954a4ddbbd14c28ba09148c1c1a8` in `5bfd332`.
- Helpers `5d13f5baaecaf11d4ea62de693426f69a1f41bba` do not merge cleanly.
  Six conflicts are recorded in wave1-09-evidence/helpers-merge.log. Its inherited
  monolithic lint changes conflict with the directory registration implementation.
  The attempted merge was aborted, preserving the registration foundation.
- docs/parallel-work.md is absent on main and both named foundations. The available
  directory ownership contract is docs/lint-registration.md and CLAUDE.md.
- The initial clone fetched only main. All origin branch refs were explicitly
  fetched. Recursive fetching of historical TypeScript submodule commits was
  stopped after parent branch refs were available; setup subsequently initialized
  the actually pinned submodules successfully.

Observed parser blocker:

The shared parser documents JSX as outside coverage in
stage1/typescript/parser/GAPS.md and REPORT.md. The three rule bodies subscribe to
JsxFragment/JsxElement/JsxSelfClosingElement, JsxAttribute, and JsxText respectively.
NoInvalidHtmlAttribute also has a React.createElement call arm, but porting only
that arm would not meet whole-rule parity.

Concrete inputs and results, on the existing parser without modifying it:

| Probe | Node source | Emitted JS | ASan/UBSan native |
|---|---:|---:|---:|
| ordinary declaration control | exit 0 | exit 0 | exit 0 |
| fragment: `<><Child /></>` | exit 70 | exit 70 | exit 70 |
| attribute: `<div rel="stylesheet" />` | exit 70 | exit 70 | exit 70 |
| text: `<div>don't</div>` | exit 70 | exit 70 | exit 70 |

All four logs compare byte for byte across the three backends. The failing
messages are `expected GreaterThanToken, got SlashToken at 23`,
`expected GreaterThanToken, got Identifier at 19`, and
`expected semicolon at 22`. Full diagnostics and raw input are in the evidence
folder. These are parser rejections, not clean lint results or successful ports.
The source-on-Node control establishes that the probe runner executes and parses
ordinary input. Native builds use native.Options{Sanitize: true}.

Reproduce from the repository root after sourcing the setup environment:

```
go run stage1/cohere/lint/claims/wave1-09-evidence/build_probe.go
node --disable-warning=ExperimentalWarning oracle/node.mjs stage1/typescript/parser/main.ts stage1/cohere/lint/claims/wave1-09-evidence/fragment.tsx.txt --whole
node --disable-warning=ExperimentalWarning oracle/node.mjs /tmp/lint-wave1-09-parser.mjs stage1/cohere/lint/claims/wave1-09-evidence/fragment.tsx.txt --whole
/tmp/lint-wave1-09-parser stage1/cohere/lint/claims/wave1-09-evidence/fragment.tsx.txt --whole
```

Repeat the last three commands with attribute, entities and control. Redirect
stdout/stderr to a log for every run. build_probe.go is a build-ignored Go harness,
not a new Adamic program; raw inputs use .tsx.txt. No new .ts file was written.

Validation and timing:

`bash cloud/setup.sh > /tmp/lint-wave1-09-setup.log 2>&1` exited 0.
Sourced `/workspace/adamic-tools/env.sh`, the configured tool path.
Go 1.27.1, clang 20.1.8, Node 24.19.0. Timing lines:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (109s)
setup: done in 109s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`go test ./internal/lint/rules/react -run '^(TestJsxNoUselessFragment|TestNoInvalidHtmlAttribute|TestNoUnescapedEntities)' -count=1 -v -timeout 15m`
from cohere exited 0: 45 top-level tests passed, package time 0.026s after build.
The complete log is upstream-rules.log. These validate upstream Go behavior, not
Adamic rule parity. `go test ./stage1/cohere/lint/registry -count=1 -v` exited 0;
registry-tests.log contains the descriptor and generation checks.
`ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout 10m` exited 0 in 5.590s: all six input fixtures passed, probe hits 0 and misses 6. The full output is filtered-oracle.log.
Setup compiled all packages/test binaries but ran no tests in its warm-cache step.
No full gate or repository-wide vet was run for this claim/evidence-only unit.

Findings per second: native unavailable, Node unavailable, Go unavailable.
No implemented rules or timed finding-count corpus exists here. The successful
Go unit-test duration is not a findings-throughput measurement. No new rule mutants were run, and no parser failure is credited as a semantic comparison mutant. The existing registry suite did run eleven invalid-descriptor controls: duplicate public name, unknown field, missing named hook, bad kind, missing oracle export, no listener, missing factory, missing class, missing finish hook, unsafe public name, and duplicate oracle adapter. Each was caught by descriptor validation, as logged; these do not satisfy the requested per-rule comparison mutants.

Inference and required handoff:

These rules cannot meet the requested contract using the current shared parser.
The missing JSX AST path must be implemented and held to the Go parser first;
that shared parser is outside this worker's assigned rule directories. Reconcile
helpers' inherited lint history with registration, and supply the missing parallel
work document. Then this claim can resume full rule ports. Claim and evidence are
pushed on codex/lint-wave1-09; no pull request is opened.
