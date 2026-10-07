# Wave 27 numeric listener declarations

Nine owned .a declarations name each completed rule and its numeric
`syntaxKinds`, matching the entry listeners of the pinned production Go rule.
They are sidecar modules for the incoming shared kind-indexed driver; today's
owned suite runners do not consume them. No new rules were claimed.

| Rule | Entry numeric SyntaxKind |
| --- | --- |
| ORM nullable parity | Decorator 171 |
| Serializable nullable parity | Decorator 171 |
| Verify array parity | PropertyDeclaration 173, Parameter 170 |
| Process exit after output | SourceFile 307 |
| Uncleared race timeout | CallExpression 214 |
| Blocking standard streams | SourceFile 307 |
| Prefer regex literals | SourceFile 307 |
| Prefer rest parameters | Identifier 79 |
| Exhaustive dependencies | CallExpression 214 |

These are the existing Go listener entry kinds, including SourceFile for the
three whole-file analyses. They are not per-node callback implementations.
`manifest.a` imports every declaration and prints its name and numeric kinds.
`oracle.go.txt` independently prints the pinned Go ast.Kind constants, built
with a virtual main overlay inside cohere/TypeScript/tsc. No shared files change.
The cohere and compiler pins remain those in the landing report.

## Validation

After sourcing /workspace/adamic-tools/env.sh, the current-main native compiler
at /workspace/wave-27-scratch/landing-v2-third/adamic built manifest.a normally
and with --sanitize. Both binaries exited zero, with empty stderr and identical
bytes to the independently compiled Go oracle. All nine mutants increment the
first kind of exactly one declaration. Every mutant compiled, exited zero with
empty stderr, and disagreed only at the Go byte comparison. Evidence/mutants.json
names every mutation and first differing byte. Complete stdout, stderr, build
and validation logs are retained with SHA-256 hashes.

The first enum-source extraction skipped KindDeferKeyword's trailing comment,
producing an off-by-one for eight declarations. The Go comparison caught this
before publishing. Initial native and Go outputs are retained as initial-*.stdout.
The final values use the compiled Go constants, and the corrected normal,
sanitized and nine mutant comparisons pass. Sanitizer output is empty.

Build commands, from the repository unless noted:

```sh
/workspace/wave-27-scratch/landing-v2-third/adamic build stage1/cohere/typeaware/wave_27_third/listeners/manifest.a -o /workspace/wave-27-scratch/listeners/native
/workspace/wave-27-scratch/landing-v2-third/adamic build stage1/cohere/typeaware/wave_27_third/listeners/manifest.a -o /workspace/wave-27-scratch/listeners/native-asan --sanitize
# From cohere/TypeScript/tsc, overlay its virtual wave27_listener_oracle.go:
go build -overlay /workspace/wave-27-scratch/listeners/oracle-overlay.json -o /workspace/wave-27-scratch/listeners/oracle /workspace/adamic/cohere/TypeScript/tsc/wave27_listener_oracle.go
```

Test subprocess stdout/stderr were written to individual files. The final
validation prints PASS: nine listener declarations; normal and sanitized;
nine normally exiting mutants. The earlier initial disagreement was observed
in the tool output and retained as separate streams. House formatting and
Go formatting completed before the corrected builds.

## Remaining shared interface gap

The current shared ParseNode in stage1/typescript/parser/nodes.ts exposes only
`readonly kind: string`. There is no numeric field and no handed-node listener
interface. The existing nine implementations still scan or refetch nodes and
read string kinds. These declarations do not claim compliance with the complete
speed rule or an achieved speedup. Replacing those accesses while preserving
Go judgments depends on the incoming numeric-node interface and shared driver.
Changes to those shared files are outside Ahra's permitted rule directories.

The prior landing sources are byte-unchanged: evidence/unchanged-landing-inputs.json
verifies every recorded source hash. Current main is still e8ba3d5d, which the
branch already contains. The complete nine-rule finding/fix/suggestion gates,
released-handle checks and timings remain the results at de8cf26b; they were not
repeated for these unused metadata modules. New checks cover the declarations,
not the existing rule dispatch. No new checker question or handle lifecycle is
introduced. No emitted-JavaScript gate or new rule timing is claimed.

The three React claims remain pending because the JSX parser integration is
absent from main. The existing probe and independent Go disagreement remain
in the fourth-batch and landing reports. No further claims were taken.
