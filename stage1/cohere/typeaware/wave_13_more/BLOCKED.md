Built: three core rule implementations are present; no-obj-calls and no-promise-executor-return agree on all exported controls, while no-object-constructor has three parser refusals.
Commits: claim a215986d was pushed before implementation; code cc7b89d0.
Commands and outputs: production upstream assertions pass; normal comparison is 430/433 controls equal, with exactly the three refusals below; both frozen corpora agree.
Mutants: namespace alias, object-literal suggestion, executor span and shorthand-binding mutants compile and exit 0, caught only by byte comparison; suffix and node-kind mutants fail dedicated tests.
Not covered: native findings/suggestions for the three refused inputs below; shared parser fixes are outside this unit's scope.

The production Go oracle accepts these upstream fixtures. The native suite exits
70 while parsing, before any rule can run:

| Case | Source shape | Native refusal |
| --- | --- | --- |
| 53d1c0a9ffc2fefa | `<foo />` then `Object()` in .ts | expected GreaterThanToken, got SlashToken at 5 |
| aff6ced2ca61fe89 | `<foo></foo>` then `Object()` in .ts | expected GreaterThanToken, got Identifier at 7 |
| defcb4c4ce6a2921 | a `yield` label and `break yield` | expected semicolon at 37 |

The missing surface is stage1/typescript/parser/parser.ts, a shared parser, not
rule logic or suggestion serialization. Ahra's instruction is to stop for other
blockers rather than edit shared files. No parser edit, special source rewrite,
spelling-only fallback or silent empty result was added. The rule implementations
and all supported controls are retained for integration. No further rules will
be claimed while this claim remains partially blocked.

## Recheck after the next keep-cooking request

Fetched all origin heads again. Main at e011f8f6 retains byte-identical parser.ts
and statements.ts at the three refusal sites. origin/codex/stage1-jsx-lint has
JSX support at e715ef4a, but its constructor enables JSX only for .tsx and
JavaScript paths, so it does not cover the two exported .ts inputs. Its label
statement condition still requires an Identifier and does not cover YieldKeyword.
The shared harness branch is now present; it does not close these parser gaps.

Replayed the exact three frozen controls: Go exits 0 and emits 699, 699 and
611 bytes; native exits 70 on each with the same parser refusal and zero stdout
bytes. Raw outputs and fetched ref SHAs are retained in validation/resume/.
No rule or parser source changed, so prior parity, mutant, sanitizer and timing
evidence remains unchanged. Stopped under Ahra's shared-file restriction and
took no further claims. nproc remains 5; the existing toolchain was reused.
