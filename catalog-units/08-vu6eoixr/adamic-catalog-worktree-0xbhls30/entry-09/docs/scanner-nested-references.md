# Scanner nested references

Closes the three nested-reference probes copied unchanged from scanner proof
`dc9f8482`. Base main: `48c05d09`. Prerequisite nested-functions tip:
`79c0361`, merged as `6167987d`. The overload worker's `433b2abe` and final
`8b7b8db9` are merged as `3cf15aed` and `f83f40f1`; its overload guards are
retained.

Nested declaration values are canonical per declaring activation. A synthetic
captured identity cell anchors a weak list of live native closure values.
References carry capture cells through intervening closures, with completed
group layouts propagated after later sibling bodies have been lowered.
The weak list never owns a closure; a closure unlinks before releasing its
cells. The JavaScript backend uses the same activation identity and code key.
This preserves equality within one activation and separation between calls,
without strong sibling-binding cycles or a native garbage collector.

The generic probe additionally requires signature and callback result types to
use the concrete specialization. A return of a void call evaluates that call
before returning. Conservative scope: only checker-proven void calls in a
void-returning body get this treatment; other void-valued expressions retain
their existing lowering boundaries.

## Observed outputs

Each control exits 0 with empty stderr in Node, native ASan/UBSan, native
release, and the JavaScript backend. Native counted/leak checks are clean.

| Probe | Source Node | Native | JavaScript backend | Native release |
| --- | --- | --- | --- | --- |
| nested-sibling-callback.a | `x\n` | `x\n` | `x\n` | `x\n` |
| nested-generic-sibling-call.a | `x\n` | `x\n` | `x\n` | `x\n` |
| nested-ancestor-call.a | `1\n` | `1\n` | `1\n` | `1\n` |

The independent identity fixture prints identically in all four executions:
`true true true false\n11 12 21\ntrue 7\n32 33\n`. It checks sibling
references, deeper references, distinct activations, escaping capture-free
functions, and recreating a value after its last strong reference was released.

## Mutants run

Each probe has one IR body mutant. Both mutated backends exit 0, with empty
stderr and clean native leak checks. Only source Node's stdout comparison
catches them; no clang or sanitizer failure is used as the catcher.

| Probe | Mutant | Source Node | Native and JavaScript mutant | Catcher |
| --- | --- | --- | --- | --- |
| sibling callback | empty reached `error` body | `x\n` | empty | stdout differs |
| generic sibling call | empty reached `worker` body | `x\n` | empty | stdout differs |
| ancestor call | reached `read` returns zero | `1\n` | `0\n` | stdout differs |

An additional identity mutant clears canonical-frame metadata. Both backends
print `false false false false\n11 12 21\nfalse 7\n32 33\n`, caught only
by stdout comparison with the independent Node identity fixture.

## Toolchain and scope

`GOPROXY='https://proxy.golang.org|direct'` was set before setup. The successful
final `bash cloud/setup.sh` reports cumulative seconds: Go 0.015, Node 0.017,
submodules 0.051, markdown 0.057, clang 0.152, build 24.552, deferred test
binaries 24.653, build cache 24.655, done 24.679. `nproc` is 5; CPU quota is 4.
Go 1.27.1, clang 20.1.8, Node 24.19.0. Environment:
`source /workspace/adamic-tools/env.sh`.

The initial successful setup took 37.199 seconds on the pre-checkout working
tree. A subsequent setup aligned main's submodule, but its build overlapped
compiler edits and failed with inconsistent Go sources; that result is
discarded. The stable setup above passes. A detached baseline worktree build
requires `-buildvcs=false` because it shares the cohere submodule by symlink.
It observes the sibling and ancestor reference stops, and the generic probe's
earlier unspecialized `T` return stop. Those are baseline observations rather
than attribution to the overload change.

No full scanner executable, token-stream comparison, scanner performance,
generic functions as first-class values, dynamic nested `this`, or block-scoped
nested declarations is claimed. No cohere code was copied. The worker gate
covers the affected packages and full oracle suite, rather than every repository
package. The focused probe tests use uncached native and release builds.

## Worker gate

All test output is redirected to logs, never piped. Commands:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestScannerNestedReference|TestNestedReferenceIdentityMutant' -count=1 -v > /tmp/scanner-nested-focus.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/scanner-nested-counts.log 2>&1
go test ./internal/lower ./internal/native ./internal/javascript ./internal/ir ./internal/flow ./internal/fresh -count=1 > /tmp/scanner-nested-packages.log 2>&1
go test ./internal/oracle -count=1 > /tmp/scanner-nested-oracle.log 2>&1
go vet ./... > /tmp/scanner-nested-vet.log 2>&1
gofmt -l cmd internal > /tmp/scanner-nested-gofmt.log
git diff --check
```

Focused uncached oracle: PASS, 4.692s, Node cache hits 0 / misses 16.
Counts update: PASS, 45.840s. Packages: lower PASS 50.325s; native PASS
243.764s; JavaScript has no package-local tests; IR PASS 16.622s; flow PASS
142.177s; fresh PASS 66.936s. Vet, final formatting and whitespace checks are
clean. Empty blank-line artifacts in locals.go and emit_maps.go from the
prerequisite merge are removed by gofmt.

Full oracle package: PASS, 244.805s. Raw setup, focused oracle, counts,
affected-package and full-oracle logs are preserved under
`stage3/drivers/scanner/evidence/nested-references/`. Main was fetched again
before landing and remained `48c05d09`; merging origin/main reports already
up to date.
