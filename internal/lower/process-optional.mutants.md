# Optional process paths (#jwyh16h)

`TestProcessOptionalStdoutHandleSysInitializer` lowers sys.ts's exact initializer,
including both assertions. `TestNodeProcessTerminal` runs that initializer and the
sys.ts `if (handle && handle.setBlocking)` guard on native and JavaScript against
Node for file, pipe and TTY stdout, and asserts the C calls
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
