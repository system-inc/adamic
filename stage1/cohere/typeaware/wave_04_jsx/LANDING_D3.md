Rebased onto current main and lint area; all eight completed source-rule oracles are freshly green; constructed-context candidate remains blocked.
Commits: tested 3b2d60f1ecfe58530183b94d5c2eea10c551f1ca; main b6b1538b0; lint area d3a37422c; evidence commit follows on codex/typeaware-wave-04.
Commands: eight source rules pass Go/corpus/sanitizer/handle checks; candidate matches 109/110 inputs, both corpora and 4287 Unicode cases; aggregate remains failed.
Mutants: eight completed-rule byte mutants, six candidate/formatter mutants, two retained-handle mutants and the adopted typeof/one-byte mutants are caught.
Uncovered: Go satisfies panic, shared checker/factory integration, full source emitted-JavaScript comparison, old partial-helper revalidation and the full gate; no new claims.

The branch was fetched and rebased with merge topology preserved onto origin/area/stage1-lint d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898, which includes current main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06 and the required registry harness. Both are ancestors. The new main changes native typeof classification and null slot presence. No shared generator, harness or protected compiler file was manually edited; those changes arrived in the rebase. No branch other than codex/typeaware-wave-04 is pushed.

Setup: bash cloud/setup.sh wrote resume-setup.log. Go, clang, Node and submodule readiness each 0s; build cache warm 94s; total 94s on nproc 5, cgroup cpu.max 400000 100000. The printed /workspace/adamic-tools/env.sh was sourced for builds/tests. The workspace exhausted its 32 GB capacity. Removing named scratch artifacts was insufficient; go clean -cache recovered 17 GB. Both failed rebase attempts and the successful retry are retained. A diagnostic original-rule run began before the rebase finished and was stopped; it is not counted. Cache cleanup invalidated a separate in-flight linker input; that failure was retained and the unchanged test retried. All counted runs use the rebased compiler: both compiler paths hash to c84f7d4e43f1b44121ce2725f5bbdd0b87538d0f2409cb95c9241eaf5dbd1e86.

Every test wrote directly to a log under /workspace/wave04-context. Current commands and observations:

```sh
source /workspace/adamic-tools/env.sh
go build -o /workspace/typeaware-wave-04-landing-d65/adamic ./cmd/adamic
ADAMIC_WAVE04_ARTIFACTS=/workspace/typeaware-wave-04-landing-d65/original ADAMIC_WAVE04_COMPILER_MANIFEST=/workspace/typeaware-wave-04/compiler.manifest ADAMIC_WAVE04_REPOSITORY_MANIFEST=/workspace/typeaware-wave-04/repository.manifest ADAMIC_TYPESCRIPT_SOURCE=/workspace/typeaware-wave-04/typescript go test ./stage1/cohere/typeaware -run '^TestWave04AgreementAndMutants$' -count=1 -v -timeout=30m
python3 -B stage1/cohere/typeaware/wave_04_next/validate.py /workspace/typeaware-wave-04-landing-d65/next --adamic /workspace/typeaware-wave-04-landing-d65/adamic --compiler-manifest /workspace/typeaware-wave-04/compiler.manifest --compiler-config /workspace/typeaware-wave-04/typescript/src/compiler/tsconfig.json --repository-manifest /workspace/typeaware-wave-04/repository.manifest --blocking-controls /workspace/typeaware-wave-04-next/blocking-controls-final --process-controls /workspace/typeaware-wave-04-next/process-controls-ts
python3 -B stage1/cohere/typeaware/wave_04_jsx/jsx_fragments/validate_source.py /workspace/wave04-fragment-source/check
python3 -B stage1/cohere/typeaware/wave_04_jsx/jsx_no_undef/validate_source.py /workspace/wave04-fragment-source/undef-check
```

Original three rules: PASS 633.135s including cold archive construction. Controls 39 findings / 17123 bytes; repository 14 / 23598; compiler 200 / 77043, each normal and sanitized. Casing, matching-return and strict-void mutants compile and exit zero with empty stderr, caught at differing bytes 5752, 122 and 1588. Released program panics 70; retained-registry mutant exits zero and fails the required-panic assertion.

Continuation three: PASS positive controls and both frozen corpora, including complete findings/fixes/suggestions and ASan/UBSan/leaks. Timeout, process-output and blocking-stream span mutants compile and exit zero with empty stderr, caught at bytes 80, 79 and 9245. Released handle exits 70; retained-registry mutant exits zero and is caught. All archives and binaries were rebuilt; no rule check was skipped to recover from the cache failure.

Fragments: syntax 3698 bytes and element 1611 bytes match Go on positive controls. Both corpora match (6088 / 18485 bytes) in both modes under sanitizers. Span mutant caught solely by Go comparison; released handle exits 70.

Undefined name: local 5436 bytes and globals 4752 bytes match on positive controls; both corpora match under sanitizers. The .cjs global-scope exception also matches. Span mutant caught solely by comparison; released handle exits 70.

Constructed-context candidate:

```sh
python3 -B stage1/cohere/typeaware/wave_04_jsx/jsx_no_constructed_context_values/validate_source.py /workspace/wave04-context/aggregate-current --normal-only
python3 -B stage1/cohere/typeaware/wave_04_jsx/jsx_no_constructed_context_values/validate_unaffected.py
python3 -B stage1/cohere/typeaware/wave_04_jsx/jsx_no_constructed_context_values/validate_remaining.py
python3 -B stage1/cohere/typeaware/wave_04_jsx/jsx_no_constructed_context_values/validate_unicode.py
```

The aggregate validator remains FAILED, exit 1 because production Go exits 2. Every individual input was attempted: 109/110 complete wire matches; control-035.tsx alone triggers the Go panic. All 110 inputs executed under ASan/UBSan/leak checks; the panic witness received memory-only checking, not claimed Go agreement. Both corpora match normal/sanitized output, 6088 and 18485 bytes, zero findings. Six message IDs are exercised by matching controls, with exact zero fixes/suggestions as in Go. Span, callee-depth, escape and primitive-type mutants compile and finish sanitizer-clean, caught by Go byte comparisons on controls 0, 83, 88 and 86. Released handle exits 70. The remaining validator intentionally returns 1 to retain the full-source blocker. Unicode/quote: PASS 4287 inputs / 77340 bytes across Go, native, sanitized native, source Node and emitted JS; uppercase and DEL-quote mutants caught.

The exact valid reproducer remains:

```tsx
declare const Ctx:any;
function Component(){return <Ctx.Provider value={({} satisfies object)}/>;}
```

The unchanged pinned cohere revision is 715ba94f3608a6500086b1076ce5cb7e51b836db. cohere/internal/lint/rules/react/jsx_no_constructed_context_values.go:484 calls AsAsExpression in its SatisfiesExpression arm. Go panics `interface conversion: ast.nodeData is *ast.SatisfiesExpression, not *ast.AsExpression`. This rule is blocked, not HIR-parked, and its input was not omitted. The witness is jsx_no_constructed_context_values/testdata/satisfies.tsx.txt; full stacks and nonzero results are archived.

Current observed wall times (overlapping builds/tests, not a controlled performance benchmark):

| Rule runner | Compiler native / Go | Repository native / Go |
| --- | --- | --- |
| Constructed-context candidate | 3.058190s / 0.511799s | 1.606534s / 0.315710s |
| Fragments | 2.914441s / 0.632943s | 0.400050s / 0.292264s |
| Undefined name | 1.482058s / 0.404665s | 0.384635s / 0.191924s |
| Continuation three | 1.967847s / 0.435370s | See complete invocation timings in resume-next.log |

Additional focused checks:

- go test ./bridge/tsgo/checker ./stage1/cohere/lint/registry -count=1: PASS 1.406s and 1.028s.
- go test ./stage1/cohere/typeaware -run '^(TestPinnedTypeFlags|TestSixPinnedFlags)$' -count=1: PASS 0.047s.
- Filtered internal/oracle run: PASS 59.756s, all seven newly adopted typeof fixtures against Node, native and emitted JavaScript, plus constructor, string, slot-presence and one-byte mutants. Worker caching is allowed and cache hit/miss counts are retained. Its combined path filter can exclude null-mutant children, so a separate exact root run executed all five: go test ./internal/oracle -run '^TestTypeOfNullMutant$' -count=1 -v -timeout=30m, PASS 55.152s. All five null-restoration mutants finish cleanly and fail Node stdout comparison; no compiler file was mutated on disk.
- go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware ./stage1/cohere/lint/registry: PASS, empty log. git diff --check: empty.

This is a filtered worker validation, not the full repository gate. The 17 required-input correctness checks were not invoked; none was relaxed, removed or skipped. The older parked React/partial-helper matrices were not rerun in this unit. Standard shared RuleContext still lacks checker/program access for the private JSX source visitors; factory integration remains unfinished. Full source checker execution through emitted JavaScript and virtual default-library foreign declaration paths remain uncovered. All three JSX claims stay reserved, existing HIR claims retain their earlier status, and no additional rule was claimed.

Evidence: evidence/landing-d3/validation.tar.gz and summary.json contain raw outputs, complete diagnostics, timings, retained failures, sanitizer stderr, mutant results and the exact tested commit/compiler identity.
