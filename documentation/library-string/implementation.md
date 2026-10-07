# String library slice

Branch: `codex/library-string`, based on `origin/main` (`fe3b9f2`), with runner branch `origin/cloud/grok-test262-runner` fast-forwarded at `bb766d4` before the baseline. Test262 is pinned to `7ab7fafa0003f73fc85c1b95d88094d33f7eb8bd`. The runner and corpus files were read before changing lowering; the runner's classifier was not changed.

## What changed

`internal/lower/library_string.go` contains the slice. Small dispatch and refusal hooks are in `expression.go`, `object.go`, and `refusals.go`. All operations use existing IR and native runtime operations; no new runtime C is necessary, and `emit.go`, `native.go`, and `lower.go` are untouched. Helpers use ordinary function calls to snapshot receiver and arguments once, in source order, before repeated reads.

- `String()` and `String(value)` convert supported primitives, including null literals, undefined, optional primitives and primitive unions. `typeof String` and callable intrinsic String methods are observable as `"function"`.
- Immediate `String.prototype.x.call(receiver, ...)` lowers supported methods with a proven present primitive receiver. Generic methods convert numeric and boolean receivers. `toString` and `valueOf` require primitive string receivers because they require a String internal slot rather than generic ToString. Null and undefined receivers refuse because their TypeError is not catchable natively. Detached method values still refuse. The exemption checks the checker symbol, so local objects named String do not become intrinsics.
- `charAt`, `substring`, `concat`, `toString`, and `valueOf` now lower on primitive strings. Index handling covers NaN, infinities, fractions, endpoint clamping/swapping, and UTF-16 surrogate units. `at`, `codePointAt`, and `normalize` already existed; fixtures and mutants exercise them and the new explicit receiver path.
- `String.raw(template, ...substitutions)` supports a present raw string array and primitive substitutions. It observes the raw field after evaluating all arguments, handles missing/excess substitutions and empty arrays, and never appends a substitution after the final segment. Its first argument gets a narrowly scoped read-only widening exemption because the intrinsic cannot write through the contextual wider type. A refusal probe ensures ordinary functions retain the wider-view check.
- `String.raw` tagged templates preserve raw escapes, invalid cooked escapes, escaped delimiters, substitution evaluation order, and ECMAScript CR/CRLF normalization.
- `fromCharCode` was already variadic. The oracle now exercises 300 literal arguments, 300 spread arguments, UTF-16 wrapping, and surrogate pairs. Existing string-pattern `replace` and `replaceAll` are covered with replacement dollar patterns.

## Intentional limits

Reading `String` into a first-class variable still refuses explicitly. The declaration combines an any-typed callable, construction and static members; the current monomorphic closure ABI has no representation for that intrinsic object. Treating it as a string-only closure would be unsound for legal numeric and boolean calls. Supporting `typeof String` is safe because it does not materialize that callable object.

Object/array/function ToPrimitive, boxed String construction/internal slots, property descriptors, getters, non-array array-like raw objects, arbitrary tags, detached methods, and unsupported argument representations remain refused. The existing type system's diagnostics are retained rather than accepting test262's untyped JavaScript style.

`localeCompare` remains refused. Node uses locale collation, which is not ordinal UTF-16 comparison even for ASCII: `"Z" < "a"` is true while `"Z".localeCompare("a")` returns 1; `"a" < "B"` is false while its localeCompare returns -1. No general ordinal subset is claimed. A refusal probe prevents an ordinal approximation.

`docs/0.1.md` explicitly places both `==` and `!=` in Never. Loose inequality remains deliberately refused, with a regression probe; this is not a lowering gap.

The RegExp unit has not landed. Regex-dependent `match`, `matchAll`, `search`, RegExp-pattern `replace`/`replaceAll`/`split`, replacement callbacks, and Symbol.match/Symbol.matchAll/Symbol.replace/Symbol.search/Symbol.split behavior are not implemented. Existing runner skips for unsupported syntax and harness features remain unchanged. String-pattern replacement is covered; no regex behavior is approximated.

## Oracle fixtures and mutant proof

Five fixtures are registered in `internal/oracle/oracle_test.go` with measured rows in `counts.md`: conversion, prototype, indices, raw, existing. Nine negative probes verify that unsafe shapes still refuse.

`TestLibraryStringMutants` runs 12 executable mutants. Each compiles, exits 0, has clean sanitizer and leak results, and is caught solely by `stdout differs` against the original fixture on Node. See [mutants.log](mutants.log) for every result.

| Family | Mutation caught by Node |
| --- | --- |
| conversion | Negate NumberToString input |
| prototype | trim becomes trimStart |
| charAt | Shift the UTF-16 index by one |
| at | Shift the index by one |
| codePointAt | Shift the index by one |
| substring | Use the end as the slice start |
| concat | Drop the final concatenated part |
| raw | Permit substitution after the final segment |
| rawTemplate | Cook a raw backslash-n escape into a newline |
| fromCharCode | Use code point construction instead of UTF-16 code units |
| normalize | NFC becomes NFD |
| replace | Replace all occurrences instead of the first |

An initial raw mutant changed only the first generated helper and survived because that call had no observable extra substitution. The final mutant changes every substitution boundary, and the fixture with excess substitutions kills it. The surviving attempt was not counted as successful proof.

## Toolchain and validation commands

All test output was redirected to log files, never piped. Commands below run after `source /workspace/adamic-tools/env.sh` (the actual setup path).

`bash cloud/setup.sh > /tmp/adamic-string-setup.log 2>&1` succeeded. Setup timing lines: go ready 0s; clang ready 0s; node ready 0s; submodules ready 0s; build cache warm 14s; done 14s. `nproc` is 5; cgroup cpu.max is `400000 100000` (four CPUs). Go 1.27.1, clang 20.1.8, Node v24.19.0. Full setup output: [setup.log](setup.log).

- Baseline: `go run ./cmd/adamic-test262 -json -test262 /tmp/adamic-string-test262 -work /tmp/adamic-string-before built-ins/String > /tmp/adamic-string-before.json 2> /tmp/adamic-string-before.log`.
- Final: same command with work and output prefix `/tmp/adamic-string-complete`.
- Fixtures and mutants: `go test ./internal/lower ./internal/oracle -run 'TestLibraryString|TestNativeAgreesWithNode/internal/oracle/testdata/library_string_raw' -v -count=1 -timeout 10m > /tmp/adamic-string-tags.log 2>&1`. PASS: lower 0.762s; oracle 20.784s; all 12 Node-only mutants and nine refusal probes passed.
- Package checks: `go test ./internal/lower ./cmd/adamic-test262 -count=1 -timeout 10m > /tmp/adamic-string-packages.log 2>&1`. PASS: lower 13.032s; runner 44.399s. See [packages.log](packages.log).
- Counts: `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/adamic-string-counts.log 2>&1`. PASS, 185.024s. After adding tagged raw cases and the 300 literal arguments, filtered measurements refreshed only those two rows; the final full gate checks all rows.
- `go vet ./... > /tmp/adamic-string-vet.log 2>&1` succeeded with empty output; `gofmt -l cmd internal` produced empty output; `git diff --check` succeeded.
- Full gate: `go test -count=1 -timeout 30m ./... > /tmp/adamic-string-gate.log 2>&1`. PASS across all packages, including oracle 890.890s and native 387.672s; all counts match. Complete output and package timings are recorded in [gate.log](gate.log).

Before/after totals, full directory tables, refusal tables, skip tables, and every newly passing test are recorded in [measurements.md](measurements.md), with unabridged runner JSON in [before.json](before.json) and [after.json](after.json).
