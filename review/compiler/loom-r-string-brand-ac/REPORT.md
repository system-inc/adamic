Built: phantom string brands from 2194037a plus five exact first-line a-check refusal headers.
Commits: source 2194037a6242a32e20f2f526fb9e9bf59d14b4c5; base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; delivery SHA in final response.
Commands and outputs: all 12 changed .a files compile/classify as expected; focused tests, counts reads, consumers, guards and vet pass.
Mutants: removed header and wrong reason each fail unchanged Gate.aCheck; all five original lowering mutants are caught again.
Not covered: full repository a-check scan, full packages/gate, WASI, ahra or Loom submission.

Branch compiler/loom-r-string-brand-ac starts from the exact requested base, with source 2194037a applied using git cherry-pick --no-commit. No conflicts. origin/main fetched to a newer revision, but the explicit pinned base takes precedence over the generic current-main instruction. This is one commit containing the original change plus the five headers and new evidence. Compiler Go, runtime C, tests and all program bodies are unchanged from 2194037a. delta.json records the exact five-file change beyond the source, excluding this evidence directory.

Read stage3/a-check-headers/README.md whole and followed its scan/harness contract. Fetched fbac28c62493f27a788edc02a18bc8edb68de5da and saved its exact cloud/fast-gate/run.py in pinned-gate.py. Gate.aCheck is unchanged. The harness only reuses the separately built branch compiler instead of rebuilding it, and substitutes that compiler binary path in spawn, with a 20-second subprocess deadline. Compiler build: timeout 180 go build -o /tmp/string-brand-ac-adamic ./cmd/adamic, exit 0. Header scan and mutant driver: timeout 240 python3 review/compiler/loom-r-string-brand-ac/headers.py, exit 0. This ran real branch compiler invocations for all initial, header, final and mutant scans; diagnostics were not fabricated or replayed. gate-provenance.json records the source revision and hash.

Status totals: checked 7; NotYet 0; refused 5; type error 0. Checked here means adamic c exits zero, not that a runtime execution was tested. In particular, the review void-view probe checks cleanly and takes no header. statuses.json includes every full diagnostic, exact added header and original body SHA-256. All bodies were verified byte for byte against 2194037a after the final scan. The initial unheadered predicate scan failed for the five negative fixtures, as expected; header and restored-final scans pass all 12.

Position audit: searched complete fixture paths and string_brand/void-view-probe references in *_test.go and testdata, and searched Markdown reports for affected path:line pins. internal/oracle/string_brand_test.go is the reader; it checks named member/cast refusals, not diagnostic positions. No test or report expectation requires changing. Historical diagnostic logs remain historical evidence. path-readers.log, report-position-search.log and position-audit.json record this audit.

Only the five refused oracle fixtures gain headers. They remain under string_brand_refused/, excluded from flow's root fixture lowering. No clean fixture or review program changes. Header comments shift refusal diagnostics by one line; final-scan.json records the actual diagnostics with headers in place.

Counts are byte-identical to 2194037a; no regeneration was performed. The original six measured rows remain intact, and focused recorded-count checks and the table check pass. counts-unchanged.json records the hash. No runtime line, feature set, IR map, product, package, dependency declaration, build-directory literal or machine-path literal is introduced by this header pass. All guard tests passed: TestEveryMapAProgramHoldsIsSorted, gob stability/nil SortedMap checks, TestRuntimeSetsAreEachUsed, TestNativeAgreementRowsNameRegisteredPrograms, TestEveryTestIsParallelOrSaysWhy and TestCallTargetReaders (-timeout 90s). Vet passes on the change's internal/lower and internal/oracle packages. No test was removed or split.

Reran all string-brand lower/oracle tests, six accepted fixtures against Node, native ASan/UBSan, release native and JavaScript backend, and leak checks; five named refusal checks; flow/fresh corpus remainders; native consumers; counts reads; guards and vet. Every go test command has -count=1 -timeout 90s -parallel 4, GOMAXPROCS=4 and ADAMIC_GATE_UNCACHED=1. All output is direct to logs; the driver is bounded by timeout 900 and each subprocess by 240 seconds. All observed test leaves and parents remain under 60 seconds. The counts-read command's wall time includes several independent measurement products and does not represent one test leaf. test-times.tsv/json lists every observed test. Source whitespace check passes.

Setup used GOPROXY=https://proxy.golang.org|direct and timeout 600 bash cloud/setup.sh; sourced /workspace/adamic-tools/env.sh. House clang 20.1.8, Go 1.27.1, Node v24.19.0. nproc reports 5, cpu.max=400000 100000 (four CPUs). Cumulative timings: submodules 0.055s, Go 0.060s, Node 0.089s, Markdown 0.205s (dependency step 0.025s), clang 0.401s, shared cache 0.798s, Go build 18.005s, test binaries deferred 18.075s, build cache warm 18.076s, done 18.101s. Setup succeeded without workaround; setup.log contains all timing lines.

Assumptions: use the explicit pinned base; include every .a changed by the source, including review/; use the documented pinned predicate; copy the full measured refusal reason including its fix text into the header. No cohere code was copied. No PR or Loom submission was made. Push only the new own branch, without force.

Per-file status table:

| Path | Status | Header added |
|---|---|---|
| internal/oracle/testdata/string_brand_escape.a | checked | no |
| internal/oracle/testdata/string_brand_map.a | checked | no |
| internal/oracle/testdata/string_brand_optional.a | checked | no |
| internal/oracle/testdata/string_brand_plain.a | checked | no |
| internal/oracle/testdata/string_brand_refused/destructure.a | refused | yes |
| internal/oracle/testdata/string_brand_refused/index.a | refused | yes |
| internal/oracle/testdata/string_brand_refused/read.a | refused | yes |
| internal/oracle/testdata/string_brand_refused/real.a | refused | yes |
| internal/oracle/testdata/string_brand_refused/unknown.a | refused | yes |
| internal/oracle/testdata/string_brand_unescape.a | checked | no |
| internal/oracle/testdata/string_brand_value.a | checked | no |
| review/compiler/string-brand-proof/void-view-probe.a | checked | no |

Exact headers added:

internal/oracle/testdata/string_brand_refused/destructure.a:

```
// a-check: refused a string brand member __brand read at runtime; keep the brand phantom and unread, or use an explicit object with a string field
```

internal/oracle/testdata/string_brand_refused/index.a:

```
// a-check: refused a string brand member __brand read at runtime; keep the brand phantom and unread, or use an explicit object with a string field
```

internal/oracle/testdata/string_brand_refused/read.a:

```
// a-check: refused a string brand member __brand read at runtime; keep the brand phantom and unread, or use an explicit object with a string field
```

internal/oracle/testdata/string_brand_refused/real.a:

```
// a-check: refused a string brand member actual with a runtime layout; keep the brand phantom and unread, or use an explicit object with a string field
```

internal/oracle/testdata/string_brand_refused/unknown.a:

```
// a-check: refused a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)
```

Header mutants:

- removed-header on internal/oracle/testdata/string_brand_refused/destructure.a: exit 1, caught by unchanged pinned Gate.aCheck with real branch compiler.
- wrong-reason on internal/oracle/testdata/string_brand_refused/destructure.a: exit 1, caught by unchanged pinned Gate.aCheck with real branch compiler.

Original lowering mutants rerun:

- read-brand: TestStringBrandReadRefused (0.03s).
- drop-undefined: TestStringBrandOptional (2.85s).
- real-field-phantom: TestStringBrandRealRefused (0.03s).
- unchecked-missing-cast: TestStringBrandCastChecksMissing (0.03s).
- literal-refinement: TestStringBrandCastKeepsLiterals (0.03s).

Commands and outcomes (exact argv in checks.json):

- counts-read: exit 0, 69.82s.
- counts-table: exit 0, 2.832s.
- lower-focus: exit 0, 3.125s.
- oracle-focus: exit 0, 6.054s.
- flow-consumers: exit 0, 5.889s.
- fresh-consumers: exit 0, 3.739s.
- native-consumers: exit 0, 2.721s.
- guards-ir: exit 0, 2.579s.
- guards-runtime: exit 0, 3.642s.
- guards-registration: exit 0, 2.527s.
- guards-parallel: exit 0, 1.117s.
- readers: exit 0, 4.555s.
- vet: exit 0, 0.922s.
- mutant-real-field-phantom: exit 1, 3.03s.
- mutant-drop-undefined: exit 1, 5.805s.
- mutant-read-brand: exit 1, 3.031s.
- mutant-literal-refinement: exit 1, 2.872s.
- mutant-unchecked-missing-cast: exit 1, 3.075s.

Added/touched test timings inherited from the source change:

- TestStringBrandCastChecksMissing: 0.06s.
- TestClockGenericReturnsT01DeclinesBrandedPrimitives: 0.1s.
- TestStringBrandCastKeepsLiterals: 0.06s.
- TestStringBrandRepresentation: 0.12s.
- TestStringBrandRepresentationRejectsLayout: 0.3s.
- TestStringBrandIndexRefused: 0.07s.
- TestStringBrandDestructureRefused: 0.09s.
- TestStringBrandReadRefused: 0.04s.
- TestStringBrandPlain: 2.91s.
- TestStringBrandValue: 3.02s.
- TestStringBrandOptional: 2.91s.
- TestStringBrandEscape: 3.03s.
- TestStringBrandRealRefused: 0.05s.
- TestStringBrandUnknownCastRefused: 0.04s.
- TestStringBrandMap: 0.4s.
- TestStringBrandUnescape: 0.43s.

Maximum observed individual test: 23.17s.
