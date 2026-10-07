Withdrawn: NewTheme is owned by slot 08 under e069e11a at 02:30:35 UTC, before our f0b136e5 at 02:32:43 UTC. No active duplicate helper or tests are delivered. Snapshots and logs below are retained as withdrawn evidence only; zero dependencies are counted from this work.

Built: collapse.NewTheme, one .a helper, with independent fresh store state.
Commits: claim f0b136e5 pushed before code; implementation commit is this report's commit.
Commands: owned go test PASS 3.258s, 1,131 cases and 27,144 Go bytes; owned go vet PASS, empty output; setup 77s, nproc 5.
Mutants: dead counter starts at 1 and shared values map; both compile, finish cleanly and differ from Go on source Node, emitted JavaScript and sanitized native.
Not covered: whole rule findings, all shared helpers, full gate, concurrent mutation; six dependencies removed, zero additional completely helper-ready rules.

The Go oracle calls the unmodified NewTheme constructor at cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. An oracle-only overlay exposes private store fields and mutation probes; the submodule is unchanged. For each case it observes the initial first theme, a second theme after mutation of the first, and a third fresh theme. The map is initialized and its retained alias is never shared between calls. The public state shape matches the independently owned Theme mutation helpers. Go's nil order slice is represented by an empty Adamic array; allocation representation is outside this observation contract, while its emptiness and independence are observed.

Capture reads every Go string literal from all test files matching each of six consumer stems. These are constructor independence controls derived from consumer literals, including settings and expected messages as well as source fixtures, not whole rule-finding comparisons. Constructor behavior does not depend on an AST or stylesheet. Two additional controls cover empty and NUL/supplementary-Unicode keys. Counts: canonical 204, class order 316, variant order 52, shorthand 178, conflicting 195, unknown 184, plus 2 controls. Three snapshots per case compare byte for byte against Go.

Consumers:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

These counts remove one ledger dependency per consumer. Design-system loading, CSS parsing, other store operations and rule ports remain outstanding; no whole rule is marked unblocked or implemented. No findings-per-second metric is meaningful for a constructor that emits no findings. The timing above measures the complete comparison, compilation and mutants, not throughput.

Reproduce with the setup environment sourced:

```
go test ./stage1/cohere/lint/helpers/from-wave1-13 -count=1 -v -timeout=10m > /tmp/new-theme.log 2>&1
go vet ./stage1/cohere/lint/helpers/from-wave1-13 > /tmp/new-theme-vet.log 2>&1
```

Setup reported Go, clang, Node and submodules each ready in 0s, cache warm 77s, total 77s on 5 processors, cpu.max 400000 100000, 17.6 GB. Earlier owned test-driver attempts exposed an unavailable input API, unsupported number/string concatenation and single-argument console typing; the final driver uses existing readTextFile and template literals. Initial failure logs are retained. No compiler or shared harness files were changed.
