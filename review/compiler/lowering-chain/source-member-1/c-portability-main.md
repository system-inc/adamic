Rebuilt portable C static globals and long-string byte storage for #277egjv and #9y65q7e.
Cherry-picks: 56ee1ddc3 from a7bb8464; 230a996db from 94909087, on main 7459656e.
Clang/GCC C11 checks, Node comparisons in both backends, ASan/UBSan/leaks, counts and focused vet pass.
Mutants caught: restored plain char produces 4,200 GCC overflow warnings; restored aggregate compound literals produce one pedantic warning per type, without -Werror.
Not covered: MSVC, a physical arm machine, or the full repository gate; opposite platform behavior is not claimed.

## Changes

Static optional-number and optional-boolean globals use plain initializer lists. Local expression
contexts retain compound literals. Long literal byte arrays use `unsigned char`; the runtime header
casts their address to `const char *` for the existing immutable string representation. Numeric byte
readers already use unsigned char. No emitter code from the older base was restored.

The source fixture crosses C11's 4095-byte literal boundary with ASCII/accent, supplementary and
mixed Unicode strings. Its long strings contain every valid high UTF-8 byte: 80..BF and C2..F4,
115 byte values. C0, C1 and F5..FF cannot occur in valid UTF-8. It checks emitted text, UTF-16 length,
UTF-8 length and byte sum, charCodeAt, codePointAt, slicing, equality, includes and undefined optional
globals. Source Node supplies the observations independently of lowered IR.

## Cherry-picks and conflicts

| Source commit | Rebuilt commit | Resolution |
|---|---|---|
| a7bb84640d1e7eaac0ea18bceeffce8f2baba228 | 56ee1ddc3e209b9323dd208ebd7957c6c1a4533c | emit_values.go retains main's absent helper and adds zeroInitializer. counts.md is regenerated on Linux, preserving every main row. Emitter and header merge automatically. |
| 949090876a7a953261c1d45f001f92c64c6c27d0 | 230a996db7b1bd5ee31f6b882e1b6c57e32d0d6d | Applied without conflicts; retains the complete UTF-8 fixture. |

Main advanced to 333fbf339ec75406f0c17bd79ca10b2d28e4d28b during this work. That range changes
stage1 tests only, with no portability-file overlap. The branch merges it without rebasing and
reruns its focused proofs. No other worker topic branch is merged. The numbered roadmap step was
not supplied for this unit; it lands the portable-C work tracked by the two named tasks.

## Proofs and test grain

Both complete native programs and their runtime archives are built with the selected compiler:
clang 20.1.8 or GCC 14.2.0. Flags include `-std=c11 -Wpedantic -Werror -fsigned-char`, address and
undefined-behavior sanitizers, no sanitizer recovery, and leak detection. Native execution must exit
0 with empty stderr and stdout byte-for-byte equal to source Node. Generated JavaScript meets the
same observation. Native O0 keeps cold runtime preparation in the leaf budget; production flags
remain unchanged. The normal uncached oracle also checks production native builds and sanitizers.

All new or touched test leaves are top-level, parallel functions. Observed seconds include each
leaf's setup; the first successful full-runtime build used a new portability cache namespace.
No leaf approaches the 60-second limit.

| Leaf | First successful run, seconds | Final focused run, seconds |
|---|---:|---:|
| TestConstantPortabilityClang | 0.05 | 0.04 |
| TestConstantPortabilityGCC | 0.03 | 0.03 |
| TestLongStringBytesMutant | 0.29 | 0.29 |
| TestStaticNumberInitializerMutant | 0.02 | 0.04 |
| TestStaticBooleanInitializerMutant | 0.03 | 0.03 |
| TestLongUnicodeLiteralByteCoverage | 0.07 | 0.11 |
| TestLongUnicodePortabilityClang | 2.98 | 0.58 |
| TestLongUnicodePortabilityGCC | 4.16 | 0.49 |
| Standard oracle leaf: long_unicode_literals.a | 0.54 | 0.54 |

Two mutant families have three checks. Restoring plain char in the emitted 4,200-byte Unicode
sample produces exactly 4,200 `[-Woverflow]` warnings. Restoring either optional scalar's compound
literal produces exactly one `[-Wpedantic]` warning. Mutant compilation succeeds with `-Werror`
absent; the warning assertions catch them. A compiler error is not accepted as that proof.
Clang accepts the old constructs in this test, so GCC supplies the portability diagnostic proof.

## Commands and outputs

Every test command writes stdout and stderr to a log file, with no test-output pipeline.
No whole package or whole repository test gate was run. Commands use the setup environment.

```text
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
go test ./internal/native -run 'TestConstantPortability|TestLongStringBytesMutant|TestStatic.*InitializerMutant|TestLongUnicode' -count=1 -v -timeout 5m
  pass on the merged tip, 0.591s
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/long_unicode_literals.a$' -count=1 -v -timeout 5m
  pass on the merged tip, 0.589s; native misses=3, Node misses=2
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 15m -args -update-counts
  pass, 48.925s
go vet ./internal/native ./internal/oracle
  exit 0, no diagnostics
```

Linux regeneration adds only this row (allocations, frees, retains, releases, peak, regions):
`long_unicode_literals.a`: **32 / 32 / 17 / 51 / 5 / 0**. The old branch's 18/52 retain/release
counts are not copied; current main removes one pair. Existing rows and fixture status records
remain unchanged.

Initial setup failed because its compilation overlapped the cherry-pick conflict in emit_values.go:
`syntax error: non-declaration statement outside function body`, `unexpected ==, expected }`,
`unexpected >>, expected }`. The conflict was resolved and setup rerun successfully, with nproc=5
and cgroup quota 4 CPUs. Successful timing lines: Node ready 0.027s; Go ready 0.039s; markdown
step 0.007s and ready 0.083s; submodules ready 0.108s; clang ready 0.252s; build ready 39.709s;
test binaries deferred 39.852s; cache warm 39.854s; done 39.887s. No toolchain failure remains.

The first direct Node harness run saw the standard experimental type-stripper warning. It now
uses the repository oracle's `--disable-warning=ExperimentalWarning`; all other stderr stays fatal.

## Integration lane checks

After committing, run from the repository root:

```text
git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

The committed merge passed: `lane checks 1.6 s: gofmt and tools on 5 Go files, t.Parallel on 2 test packages; vet 2 packages`. The final committed tip is checked again before pushing. GCC is
installed on this box and both GCC leaves actually ran. The upstream tools manifest does not yet
declare GCC; the lane's literal-command scan accepts the original branch's parameterized compiler
harness. GCC provisioning on other gate boxes remains the devtools owner's responsibility.

Evidence: [logs.tar.gz](c-portability-main-evidence/logs.tar.gz).
