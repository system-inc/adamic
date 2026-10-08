# String destructuring scanner unit

Built on origin/main 48c05d091f0a43c31cbe051b1d6578d99eeedf19, branch
codex/scanner-string-destructuring. Source probe is unchanged from dc9f8482:
string-length-destructuring.a, recorded as the scanner.ts:776 stop.
Its source comment identifies upstream scanner.ts:1333.

Object binding declarations from strings now hold a string, read length in
UTF-16 units, and read canonical numeric properties through StringIndex.
An absent index takes its default lazily. The held source survives defaults
that replace the original variable. Existing runtime operations provide both
backends; no native emitter or runtime changes were needed.

Conservative scope assumption: this unit covers declarations, with length and
canonical decimal indices representable as uint32. String parameter patterns,
array patterns, nested/rest bindings, computed keys, other properties, indices
outside that range, and defaults requiring another representation remain NotYet.
No full scanner native build or full repository gate is claimed.

## Observations

Baseline with the pinned main submodule, original probe:

```text
Node: 10, exit 0, empty stderr
adamic: /tmp/string-length-destructuring.a:2:7: stage 0 can't lower destructuring a string yet
compiler exit 1
```

Final original probe, stdout side by side (each exit 0, empty stderr):

| Source Node | Native | JavaScript backend on Node |
| --- | --- | --- |
| `10\n` | `10\n` | `10\n` |

Extended fixture, identical on source Node and both backends:

```text
4|a|55356|57101|missing|11
0|empty
default|o|new
```

This observes UTF-16 length and both surrogate indices, a present-index default
skipped, an absent-index default called, empty-string defaults, one source
call, and a later property read from the held original source. Native sanitizer
and leak checks pass. Counts regeneration changed only the two new rows:
length probe 2/2/1/3/2/0; extended fixture 16/16/14/33/9/0, in
allocations/frees/retains/releases/peak/regions order.

One semantic mutant: add one to the destructured length. The ordinary uncached
oracle test exits 1 with precisely these mismatches:

```text
node:   exit 0, stdout "10\n", stderr ""
native: exit 0, stdout "11\n", stderr ""
backend: exit 0, stdout "11\n", stderr ""
```

The compiler mutation was restored with a shell EXIT trap. The permanent
TestScannerStringDestructuringMutant also builds a length-plus-one C mutant
under sanitizers with leak detection. It requires clean execution and the
specific `stdout differs` comparison against source Node; build errors or
sanitizer failures do not count as kills.

## Commands and logs

Every test wrote to a log before the log was read. Environment file:
/workspace/adamic-tools/env.sh, sourced in every toolchain shell.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/scanner-destructuring-setup-main.log 2>&1
go test ./internal/lower -count=1 -timeout 30m > /tmp/scanner-destructuring-lower.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestScannerStringDestructuringMutant|TestNativeAgreesWithNode/internal/oracle/testdata/scanner_string' -v -count=1 -timeout 30m > /tmp/scanner-destructuring-focus-final.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/scanner-destructuring-counts.log 2>&1
go vet ./... > /tmp/scanner-destructuring-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestScannerStringDestructuringMutant|TestNativeAgreesWithNode/internal/oracle/testdata/scanner_string|TestCountsAreRecorded' -count=1 -timeout 30m > /tmp/scanner-destructuring-final.log 2>&1
gofmt -l cmd internal > /tmp/scanner-destructuring-format.log
git diff --check
```

Lower package: ok, 13.257s. Focused oracle and permanent mutant: ok, 0.455s.
Counts update: ok, 17.608s. Vet: exit 0, empty output.
Actual mutant failure: /tmp/scanner-destructuring-mutant-failure.log.
Baseline: /tmp/scanner-destructuring-baseline.log. Explicit probe stdout:
/tmp/scanner-length-{node,native,js}.log. Final oracle includes the complete
counts check, with source/native/JavaScript comparisons filtered to this family.

Setup on main: Go ready 0.016s, Node ready 0.016s, markdown dependencies ready
0.058s, clang ready 0.139s, submodules ready 2.681s, Go build ready 208.466s,
test binaries deferred 208.572s, cache warm 208.573s, done 208.606s.
nproc 5, CPU quota 4, Go 1.27.1, clang 20.1.8, Node 24.19.0.

Initial recursive fetch stalled in submodule fetching and was interrupted;
main itself was fetched correctly, and later fetches disabled recursion.
First setup completed on the initial checkout in 37.307s. After switching to
main, a premature probe compile failed exactly with `no required module
provides package github.com/system-inc/cohere/rule_runner`. Rerunning setup
synchronized cohere to main's recorded revision and resolved that failure.
The first extended fixture used optional method-call syntax and stopped at
`a call through ?. (an optional call)`; coalescing before the call removed that
unrelated feature dependency. A first mutant helper used a relative path and
failed cache path normalization; the final helper uses an absolute path.
Direct execution of emitted JavaScript could not resolve package adamic;
running it through oracle/node.mjs supplies the repository runtime resolver.
None of these failed attempts is counted as semantic evidence.
