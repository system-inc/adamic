# nexus/correctness-no-process-exit-after-output: shared dependency missing

Audited against origin/area/stage1-lint at c4bdc23fa86d55cf7e579989201c11258f4d3a62, with cohere pinned at 7945d102a6c18dd36adf9114a758ce646e8b2359.

Exact upstream dependency: `checker.Checker.GetAliasedSymbol` at `cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output.go:413`.

Required answer/helper: Resolved alias symbols and their declarations.

Shared checker evidence: `bridge/tsgo/checker/facts.go` on this area pin exposes `symbol-origin` (only the filename of `symbol.ValueDeclaration`), `type-origin` (the type's symbol, not the identifier's resolved binding), and `declarations` (class/interface selectors only). None supplies this dependency. `scope-locals` is a binder-table enumeration, not reference resolution. The shared lint context has no reference tracker or high-level IR helper.

The old wave branch contains additional bridge questions. They are not on the area branch, so this audit does not treat them as shared APIs or import a private checker/helper.

Reproducer: `reproducer.ts.txt`. This is a source-level example requiring the missing decision, not a claim of an executed oracle comparison. React and Node examples need their matching declaration inputs. No descriptor is registered: an incomplete rule must not appear green by producing no findings.

Stopped before implementation as instructed. Upstream cases matched: 0. Node/emitted-JavaScript/sanitized-native parity: not run for this rule. Mutant: not created or run because no complete shared-checker implementation exists.
