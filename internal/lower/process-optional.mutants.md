# Optional process paths (#jwyh16h)

`TestProcessOptionalStdoutHandleSysInitializer` lowers sys.ts's exact initializer,
including both assertions. `TestNodeProcessTerminal` runs that initializer and the
sys.ts `if (handle && handle.setBlocking)` guard on native against Node for file,
pipe and TTY stdout, and asserts the C calls
`adamic_node_stdout_handle()`. The existing terminal WASI refusal remains named.

| Mutation | Named pin | Catch line |
| --- | --- | --- |
| `processPathPresent` admits every optional receiver | `TestProcessOptionalAbsentReceiverStopsByName` | `absent receiver must stop processPath, got "process.stdout.columns.toFixed"` |

The pin checks that corrected Node declarations leave columns nullable, that
processPath stops at its optional receiver, and that full lowering keeps the exact
`stage 0 can't lower a call through ?. (an optional call) yet` refusal. The direct
path check matters: the generic optional-call refusal remains independent of the
host path recognizer, and must not mask an incorrectly admitted absent receiver.
`TestProcessOptionalEverywhereMutantIsCaught` runs this pin from the privately
mutated test binary and requires its exit 1 and the catch line, never a build error.
Optional element links remain stopped, preserving their existing gate.

## Open finding: JavaScript host-method guard-and-call (#jwyh16h)

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
