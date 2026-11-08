# Optional process paths (#jwyh16h)

`TestProcessOptionalStdoutHandleSysInitializer` lowers sys.ts's exact initializer,
including both assertions. `TestNodeProcessTerminal` runs that initializer and the
sys.ts `if (handle && handle.setBlocking)` guard on native against Node for file,
pipe and TTY stdout, and asserts the C calls
`adamic_node_stdout_handle()`. The existing terminal WASI refusal remains named.

| Mutation | Named pin | Catch line |
| --- | --- | --- |
| `processPathPresent` admits every optional receiver | `TestProcessOptionalAbsentReceiverStopsByName` | `absent receiver must stop processPath, got "process.stdout.columns.toFixed"` |
| Drop the non-nullable comparison only | `TestProcessOptionalNullableReceiverStopsByName` | `nullable receiver must stop processPath, got "process.exitCode.toString"` |
| Drop the SymbolFlagsOptional check: equivalent, because every optional declared type checked here is rejected by the non-nullable comparison | `TestProcessOptionalFlagIsRedundantForReachableMembers` | equivalence proof; no red pin is claimed |

The pin checks that corrected Node declarations leave columns nullable, that
processPath stops at its optional receiver, and that full lowering keeps the exact
`stage 0 can't lower a call through ?. (an optional call) yet` refusal. The direct
path check matters: the generic optional-call refusal remains independent of the
host path recognizer, and must not mask an incorrectly admitted absent receiver.
`TestProcessOptionalEverywhereMutantIsCaught` runs this pin from the privately
mutated test binary and requires its exit 1 and the catch line, never a build error.
Optional element links remain stopped, preserving their existing gate.

## Open finding: JavaScript host-method guard-and-call (#qa1g5f3, Library's)

The fixture uses the exact sys.ts 1609–1612 initializer and guard-and-call, with
no optional call. JavaScript agrees with Node for file stdout (no handle). On
pipe and TTY stdout its named `javascript` row asserts exactly
`stdout="handle function\n"`,
`stderr="adamic: panic: TypeError: closure.code is not a function\n"`, exit 70.
A changed result fails the row: it is an open finding, never a skip or a claim
of Node agreement. Native remains held to Node on all three destinations.

Hand witness on unmodified main d15620712, Node v24.19.0, Linux, pipe stdout:

```typescript
import 'node:process';
const handle = (process.stdout as any)._handle as { setBlocking?: (value: boolean) => void; };
if (handle && handle.setBlocking) { handle.setBlocking(true); }
console.log('set blocking');
```

| Runner | stdout | stderr | Exit |
| --- | --- | --- | --- |
| Node source | `"set blocking\n"` | `""` | 0 |
| Sanitized native | `"set blocking\n"` | `""` | 0 |
| JavaScript backend | `""` | `"adamic: panic: TypeError: closure.code is not a function\n"` | 70 |

The guarded optional function property is called through `adamicCall` in the
JavaScript backend (`internal/javascript/javascript.go`, CallClosure's fallback),
bypassing `adamicCallee`'s host-method adapter. The same failure with a plain dot
proves it predates optional process-path admission. The compiler used here was
built on unmodified d15620712 before the optional-path edit.

## Uncast present-receiver read (#5d5xapx)

`TestProcessOptionalPresentReadMatchesDot` compares the complete statement IR of
`const value = process?.execArgv` with the dot form, structurally, and requires the
existing `ir.ProcessCall{Operation: "execArgv", Of: ir.Array}` operation. The
receiver is uncast `process`; `nodeProcessValue` recognizes its read via processPath.
The oracle fixture `process_optional_present_read.a` reads both forms and compares
array identity within each process, avoiding machine-specific command arguments.

`process.stdout?.write("x")` is deliberately not this pin: nodeStdoutWriteRefusal
has its own `node:process.Process.stdout.write through an optional chain` refusal.
That refusal is preserved, including the control owned by #v57er4w.

## Per-clause evidence (#5d5xapx)

Clause (b)'s real member is `node:process.Process.exitCode`, published at
`internal/load/node_types/node_modules/@types/node/process.d.ts:1127` as
`exitCode: number | string | null | undefined;`, without `?`. The pin asserts its
registration, lack of SymbolFlagsOptional, declared nullability, closed processPath,
and the exact existing mixed-union `toString` NotYet. The nullable-only mutant
must fail that pin with exit 1 and the catch line above, then run the columns pin
from the same binary and require its PASS and exit 0.

`TestProcessOptionalFlagIsRedundantForReachableMembers` uses load.Load's real
ExactOptionalPropertyTypes configuration and the corrected Node declarations.
It walks the process property graph through unregistered intermediate properties
and index value types, checking every registered optional member. It also
explicitly checks columns (`tty.d.ts:183`, corrected to `columns?: number`), which
has its own host adapter and is not in implementedNodeMembers. For every optional
symbol checked, the declared type must differ from GetNonNullableType(declared).
A counterexample fails the proof; no synthetic member, registration or axiom is
introduced. Thus removing only the optional-flag check is recorded as equivalent,
while the existing everywhere mutant remains required.
