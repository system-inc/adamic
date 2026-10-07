Built: native scoped constant strings and RegExp-object identity, following local initializer chains and rejecting writes and cycles.
Commits: follows pushed f69aee8e on codex/typeaware-wave-09, based on fetched current main e8ba3d5d; no new claims.
Commands/output: verify_bindings.py PASS, 35 actual source files and 2227 identical bytes; expression-core regression passes all 296 cases across both Node modes and sanitized/native execution.
Mutants: ignoring writes differs at byte 240; rejecting every second binding in a chain differs at byte 883; both compile and exit 0 with empty stderr.
Not covered: global constructor reference tracking, checkedByACall integration, the complete regex rules and the native invalid-regexp engine.

constant_bindings.a consumes existing raw checker binding declarations and
declaration-file flags. No new checker question or registration was needed.
All constant eligibility, initializer selection, write detection, expression
folding and chain traversal run in native Adamic. It accepts one local scalar
variable declaration with an initializer. Const declarations follow production
Go's behavior even if erroneous source later writes them; non-const bindings
must have no additional write to that same binding. Shadowed same-spelling
names are distinguished by checker declaration identity. Names declared only
in declaration files and non-variable/destructured bindings decline.

Each recursive path records declaration identities, so a self-reference or a
cycle declines and a legitimate chain remains known. RegExp-object identity
follows only regex literals and eligible identifier initializers, keeping a
regex object distinct from its string representation. The expression fold can
use the regex text inside concatenations/templates while IsConstantRegExpIn's
native counterpart correctly declines those string-valued expressions.

The independent Go oracle loads the actual source/checker program and directly
calls unchanged reference.ConstantStringIn and reference.IsConstantRegExpIn.
It imports no bridge implementation. Native loads the same sources through its
own Parser and Rules node/parent tables, then asks only raw checker facts. Every
known flag, value string and regex-identity bit is compared in order. Tests cover
initialization, later assignment, compound/update/loop writes, nested and
parenthesized destructuring/rest, property/key reads, lexical shadowing, merged
declarations, alias chains, cycles, constants in interpolated templates, imports,
parameters, type assertions, destructured declarations and unsupported builtins.
Semantic source errors are intentionally allowed; parser errors are not filtered.

The initial class-return callback was refused by main's nominal ancestry check.
The fold result now has an explicit readonly interface, because that result is
a data shape passed across recursive callbacks. Concrete ConstantValue objects
still supply it. The unchanged expression oracle was rerun after that API change:
296 cases / 2646 bytes agree with Go, native, source Node and emitted JavaScript,
including ASan/UBSan/LeakSanitizer and the clean-running coercion mutant.
Scoped execution itself was compared native-to-Go, not through the Node bridge:
the existing oracle/adamic.mjs does not export the tsgo runtime functions.
That execution mode remains uncovered and is not claimed as a pass.

The scoped normal/sanitized streams match and have empty stderr. The two scoped
mutants are successful programs caught only by Go bytes. The chain mutant is
conservative, rather than disabling cycle detection and causing unbounded
recursion: a runtime crash would not prove the diagnostic/value comparison.
Actual self and mutual cycles remain positive tests for declining values.
Single normal native/Go times: 0.114480 / 0.063946 seconds. This includes source
loading and checking for the 35 controls; no full-rule throughput is claimed.
Toolchain setup is reused from the same workspace, previously 88 seconds; nproc 5.

Reproduce after sourcing /workspace/adamic-tools/env.sh:

```
python3 stage1/cohere/typeaware/wave09_core/misleading_character_class/verify_bindings.py > /workspace/wave-09-bindings-test.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/misleading_character_class/verify_constants.py > /workspace/wave-09-binding-core-regression.log 2>&1
```

The local verifier links /workspace/wave-09-core/checker.a and checker-asan.a,
previously built by label_var/verify.py with its owned scope-symbol overlay.
Those archives expose all existing raw questions used here; their shared ABI
was not changed. Prior bridge/released-handle/mutant results remain in the
landing evidence and were not repeated for this unchanged bridge implementation.
No shared compiler, harness, registration or production Go file changed.
New Adamic sources are .a. Complete streams, generated controls, measurements,
mutants and source hashes are retained in validation-bindings.

Both complete regex rules are still unfinished. This reader and the mapped
constructor judgments are not yet connected to a full native reference tracker
and source traversal with checkedByACall suppression. The exact pattern engine
needed for no-invalid-regexp remains absent. Full upstream matrices, all options,
the complete repository gate and native full-rule regex timing remain uncovered.
