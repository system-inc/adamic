Built the Node argument absence and consumption proof; 17 fixtures remain blocked at Error-to-ErrnoException initializers.
Branch: codex/node-options-widening; refusal-3 merge cbde74ef, latest p2b merge 1d5e5647 (f133f7c7).
Merged boundary checks pass (6.682s), new Node/native/JavaScript oracle passes (1.283s), restored predicate passes (3.319s).
All three production predicate mutants fail their intended absence/consumption assertion, before and after the compiler merge.
Limits: flow, full oracle, counts and configured WASI host gates fail; the 17 error views require a separate compiler-approved repair.

`internal/lower/library_node_options.go` implements `nodeHostConsumesArgument`.
It recognizes pinned Node declarations rather than names supplied by application
code. Only the option positions of statSync, mkdirSync, readFileSync, rmSync,
mkdtempSync and writeFileSync qualify: the fs-file lowerer replaces these objects
with scalar runtime arguments. A path, buffer, callback, or unrecognized argument
does not qualify. readdirSync does not qualify because its options object is passed
onward by the current lowerer.

The absence proof reuses `exactObject`, which follows unannotated const bindings
to their complete literal shapes and rejects annotations, casts, spreads and
parameters. Every declared field must also have a scalar type. This prevents an
exact outer object from concealing a nested optional widening through another
object or a callback. This is deliberately conservative: unsupported class or
other exact-shape proofs are not inferred. The lowerer's existing named NotYet
checks still govern the admitted overloads and option values. No runtime-shape
reading or wider stored object view was added.

Compiler's existing call at `internal/lower/optional_widening.go:135` is unchanged.
The false stub in optional_node_host.go was removed in favor of the declaration
in library_node_options.go. The predicate itself also verifies the node:* import
binding and the pinned declaration, so an independently invoked predicate cannot
bless an ordinary function merely because its name matches a Node member.

The requested refusal-3 branch was initially absent, then published at a21fa7a8
and merged with a merge commit. Before it appeared, refusal-2 was inspected
through a temporary Go overlay without merging it. The actual merged compiler's
relation walk confirms that all 17 named fixtures first refuse their error printers'
Error-to-ErrnoException initializers, with missing optional `errno`. That is outside a Node argument
position and must not be exempted by this predicate. The stat and rm fixtures'
plain const options satisfy the absence proof without changes. No acceptance
fixture has been rewritten, and none of the 17 is claimed green under refusal-3.
The separate new node_options_widening.a oracle passes with its original source
on Node, generated JavaScript, native, sanitizers and the leak check.

The independently verified diagnostic from the prior refusal is:

```text
Adamic 0.1 refuses optional property errno in ErrnoException absent from structural source Error, which can hide fields; declare errno on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

Its source reproducer is:

```a
import {rmSync} from 'node:fs'; function report(error:Error):void {const errno:NodeJS.ErrnoException=error; console.log(errno.code??'');}
```

The complete per-fixture observations are recorded in both the prior-refusal and
merged-refusal inventory logs. The latter uses a21fa7a8's actual production rule.

The mutants actually run and restored are:

| Mutant | Assertion that catches it |
|---|---|
| Admit a non-Node call | non-node argument must not be consumed |
| Admit an annotated structural subtype with hidden force | structural subtype must not prove absence |
| Admit an options object passed onward by readdirSync | forwarded object must not be consumed |

Each mutant exits 1 with `consumes argument = true, want false`. Compiler or
backend failures do not count as kills. The restored suite exits 0. The separate
hidden_force.a fixture records actual Node behavior: its narrower structural view
hides force='yes', and Node validates that field and rejects its string value.

Commands completed so far:

```sh
go test ./internal/lower -run '^TestNodeHostConsumesArgument$' -count=1
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/lower/testdata/node_options_widening/hidden_force.a
```

Output was sent to local evidence logs, which are ignored and are not exported.
Temporary Go overlays verified the previous
refusal's diagnostic and the first refusal in each of the 17 fixtures. Those
overlays did not modify the repository's compiler files.

The original p2b flow failure on process_bad_code.a was reproduced with the new
predicate excluded through a Go overlay (0.205s). Integration's f133f7c7 fixes it;
the focused flow test now passes (0.181s). The merged compiler's complete flow
gate additionally fails to lower the error-printer views. The embedded
internal/load/node_fs_file_system.a:37 has the same initializer refusal, which
also affects the system and buffer fixtures.

Full Linux counts regeneration was attempted and failed (206.602s), so the
harness correctly preserved its table rather than dropping unlowered rows.
The new oracle's row was measured separately on Linux: 5/5/1/7/5/0, and only
that row was added. The focused TestCountsAreRecorded comparison passes after insertion (2.543s).
The complete table cannot be regenerated while the required fixtures refuse.

Setup first failed to initialize copied dependency archives as submodules. After
attaching their pinned local Git metadata, setup passed. Timings: Go 0.192s,
Node 0.333s, submodules 0.509s, clang 1.312s, build 528.358s, total 528.799s.
`nproc` is 5, cgroup quota is 4 CPUs, and memory is 17.6 GB. Every Go shell sets
GOPROXY='https://proxy.golang.org|direct'.

Only the named worker branch will be pushed. Main will never be pushed, and no
force push or history rewrite is authorized by this work.

Final package command: `go test ./internal/lower ./internal/ir ./internal/flow -count=1`.
Lower passes in 164.329s; IR passes in 33.104s; flow fails in 304.607s at the
Error-to-ErrnoException views listed above. `go vet` on lower, IR, flow and
oracle passes. No fixture source was rewritten.

WASI SDK 27 was installed with `cloud/setup.sh --wasi-sdk` (83.125s), with
WASI_SYSROOT pointing at its share/wasi-sysroot directory. The first broad
oracle run started before that configuration and is not evidence of a configured
WASI pass; a separate complete configured TestWASI group was run.

Configured WASI observations: TestWASIAgreesWithNode passes (382.95s), including
the new options oracle (0.77s). WASI comparison mutants and runner mutants pass
(0.87s and 1.63s). WASIHostRuntimeRefusals passes (2.68s). WASI input and file
comparisons fail at the same Error-to-ErrnoException initializers, before runtime.

Final gate results: the whole oracle command
`ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -count=1 -timeout=30m -parallel=2`
failed in 974.507s. In addition to error-view refusals, this run records a regex
long-backtrack timeout and native regex timing failure. Its WASI tests lacked
WASI_SYSROOT; those configuration failures were followed by the correctly
configured complete group, not treated as semantic observations.

`ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestWASI' -count=1 -v -timeout=30m`
with SDK 27's WASI_SYSROOT fails in 550.389s solely in the host input/file groups
at error-view refusals. The ordinary comparisons, emission, runtime refusals and
both WASI mutant suites pass. No runtime C files or target guards were changed.

This branch is a reviewable blocked checkpoint, not a claim that all gates or the
17 requested fixtures are green. Fixing the Error-to-ErrnoException initializers
requires a separately authorized checked error view; widening this consumption
predicate to initializers would violate the requested rule.
