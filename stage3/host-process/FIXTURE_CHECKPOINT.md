Built acceptance proof for unchanged 23_newLine, using the existing node:os namespace lowering; no new require implementation.
Branch codex/host-process-land, base 12b84d61 (contains area/library 53f44e05); no trial-base changes imported.
Commands: newline and namespace oracle PASS 2.913s; builds reproduce the two language blockers below.
Mutants: native and JavaScript LF-to-CRLF substitutions run with exit 0 and empty stderr, rejected only by Node stdout comparison.
Not covered: 16 and 17 language blockers, macOS execution, whole oracle rerun; no runtime/lowering change required for 23.

| Fixture | Native | JavaScript | Node status.json |
| --- | --- | --- | --- |
| 23_newLine | Agrees | Agrees | Agrees |
| 16_getEnvironmentVariable | NotYet: string BinaryExpression | Same language lowering | Source unchanged |
| 17_write | NotYet: literal method call with unrepresented result | Same language lowering | Source unchanged |

Language reproducer for 16: `const value: string = ""; console.log(value || "fallback");` fails without Node APIs, at the string `||`. The host fixture fails at `return process.env[name] || "";`.

Language reproducer for 17: `const system = { write(s: string): void { console.log(s); } }; system.write("hello");` fails without Node APIs. The exact fixture fails at its first `nodeSystem.write("first")`, not at process.stdout.write's boolean. This branch's lowerer rejects the literal method's void result before lowering its body. No workaround was made.

The compiler namespace identity fix already on this branch resolves `_os.EOL`; the new oracle test executes the full fixture, including real mkdtempSync/rmSync setup and finally cleanup, on sanitized native and generated JavaScript. It also checks source Node against the recorded status.json. Counts need no regeneration: no registered generic fixture or runtime changed. Main was not pushed; no force-push or rebase.
