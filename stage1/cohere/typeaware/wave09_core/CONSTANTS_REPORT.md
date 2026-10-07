Built: native expression-folding core for the constructor constant reader, parsing actual source rather than receiving AST fixture tables.
Commits: follows pushed 59964b46 on codex/typeaware-wave-09, still based on current fetched main e8ba3d5d; no new claims.
Commands/output: verify_constants.py PASS on 296 expressions against unmodified Go reference.ConstantString, native, sanitizers, source Node and emitted JavaScript.
Mutant: require both operands of + to be strings; builds, exits 0 with empty stderr, caught by the independent Go byte comparison.
Not covered: scoped binding resolution, post-declaration write detection, reference tracking, complete constructor listener and native no-invalid-regexp engine.

constant_value.a ports reference.constantValue's expression core. Its result
keeps the text, whether it is a string, and whether it is known, because the
string bit decides Go's deliberately limited + folding. Parentheses, strings,
substitution-free templates, numeric literals, booleans, null, template
substitutions and string concatenations match the production reader. Numeric
addition, unary expressions, type assertions, unknown names and safe builtin
calls decline exactly as that Go reader does. The constantString entry point
declines regex literals and identifiers without scope; constantValue exposes a
separate typed resolver callback for the future scoped adapter. No Go verdict
is supplied to the native fold.

The independent oracle parses each expression with typescript-go, then calls
the unchanged public Go reference.ConstantString. Native parses the same actual
source with Adamic's Parser and locates the initializer in its own node table.
Every known flag and complete escaped value is compared byte for byte, including
unknown results. Numeric spellings, Unicode and nested templates are exercised.
The 16 primitive shapes, their 256 ordered pairs, 16 template substitutions and
eight additional cases total 296 expressions. Both Native/Node backends use
the actual source parser. The Go oracle is placed inside cohere via a temporary
overlay only to satisfy its internal-package import boundary.

Reproduce after sourcing /workspace/adamic-tools/env.sh:

```
python3 stage1/cohere/typeaware/wave09_core/misleading_character_class/verify_constants.py > /workspace/wave-09-constants-test.log 2>&1
```

ASan, UBSan and LeakSanitizer pass with empty stderr. The semantic coercion
mutant changes which genuine source expressions are known, and only the Go
comparison catches it. Exact bytes, first differing byte and source hashes are
retained in validation-constants. Single whole-process core times: native
0.007563 seconds / Go 0.007706 seconds; native fixtures are embedded while Go
loads JSON. This is component timing, not full-rule throughput. Setup remains
the same workspace's 88-second setup with nproc 5.

The original rules and mapped constructor component are unchanged. No shared
compiler, harness, registration or production Go rule was edited. Existing
landing corpus, checker ownership and released-handle checks are retained in
validation-landing; the checker ABI did not change here. All new Adamic files
are .a. The complete regex claims remain incomplete: this fold is not a native
symbol resolver, does not decide whether let is effectively constant, and is
not yet wired to the whole constructor listener. No full repository gate or
complete upstream matrix was run for this component.
