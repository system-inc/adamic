Built: six .a sidecar listener declarations with canonical rule names and numeric parser-kind arrays, ready for the shared kind-indexed driver.
Commits: follows pushed 22915212 on codex/typeaware-wave-09, based on fetched current main e8ba3d5d; no new claims.
Commands/output: listeners/verify.py PASS, 168 identical bytes across production Go listener maps, native, sanitizers, source Node and emitted JavaScript.
Mutant: radix listens to NewExpression 215 instead of CallExpression 214; builds and exits 0 with empty stderr, caught at byte 166 by Go comparison.
Not covered: numeric dispatch and handed-node execution; the current shared parser exposes only string kinds, so the legacy rule execution is not speed-compliant yet.

## Numeric declarations

Each file under listeners exports ruleName and syntaxKinds:readonly number[].
No declaration reads a node's kind or refetches a node. Numeric values come
directly from the pinned production Go parser's actual listener maps, rather
than numbers copied from another TypeScript release or a invented ordering.
The oracle invokes unchanged rule.Run factories and reads their map keys.

| Rule | Declaration | Numeric kinds |
| --- | --- | --- |
| @typescript-eslint/non-nullable-type-assertion-style | listeners/non_nullable_type_assertion_style.a | 217, 235 |
| no-invalid-regexp | listeners/no_invalid_regexp.a | 214, 215 |
| no-label-var | listeners/no_label_var.a | 257 |
| no-misleading-character-class | listeners/no_misleading_character_class.a | 13, 307 |
| prefer-const | listeners/prefer_const.a | 262 |
| radix | listeners/radix.a | 214 |

These include the source-file listener of the full Go misleading-character-class
rule, which initializes constructor tracking before regex literal listeners.
Declaring it does not assert the still-incomplete native tracker is implemented.
The numeric contract is pinned to this branch's Go parser. The shared numeric
parser/driver must expose the same IDs, or provide one explicit central mapping;
per-rule conversion from string names would violate the requested speed rule.

## Exact shared integration gap

stage1/typescript/parser/nodes.ts defines ParseNode.kind as string and its
constructor accepts a string. It has no numeric SyntaxKind property or numeric
kind accessor. The scanner similarly exposes string kinds. The current main,
origin/area/stage1-lint and origin/codex/lint-registration node definitions were
inspected read-only and retain that representation. The typeaware Rules.ask
path also fetches the node by index and serializes its string kind for the
checker bridge. No shared listener callback/handed-node context contract exists
on this branch for a rule to use instead.

Therefore this commit supplies the requested declarations now, but does not
claim existing legacy execution obeys the speed rule. Original rules and the
incremental constructor/constant helpers still contain their historical
string-kind comparisons. Replacing those while preserving byte parity requires
the shared numeric parser/context contract. A rule-local string conversion or
new node refetch would preserve the precise cost the user prohibited, so neither
was introduced. Shared parser, driver, generator, bridge and compiler files
were left untouched. This is the concrete shared API blocker for the next
execution migration, independent of the unfinished regex engine and tracker.

## Verification and scope

Reproduce after sourcing /workspace/adamic-tools/env.sh:

```
python3 stage1/cohere/typeaware/wave09_core/listeners/verify.py > /workspace/wave-09-listeners-test.log 2>&1
```

The independent Go oracle lives in owned testdata and is built via an overlay
inside cohere only to satisfy Go's internal import boundary. Normal native,
ASan/UBSan/LeakSanitizer, source Node and emitted JavaScript all print the exact
six numeric arrays and names. The successful wrong-kind mutant is caught only
by that byte comparison. Full streams, generated probes, measurements and
source hashes are retained in validation-listeners.

Single declaration-print times: native 0.001429 seconds, Go 0.015777 seconds.
Go constructs production rule maps while native prints constant metadata;
this is not a lint-performance comparison and demonstrates no lint speedup.
No finding/fix/suggestion implementation or checker ABI changed, so prior
rule corpus, released-handle and sanitizer results remain in validation-landing,
validation-bindings and the other component reports, without another full
corpus sweep for these independent metadata files. The full repository gate
and shared kind dispatch were not run. Toolchain setup is reused from this
workspace's 88-second setup; nproc remains 5.

The branch passed the landing cap before this work: fetched origin/main was
e8ba3d5d and already an ancestor of the pushed wave-09 tip. Publishing remains
to codex/typeaware-wave-09 only. No other branch, main or area branch is pushed.
No new rules are claimed while the existing complete regex ports are unfinished.
