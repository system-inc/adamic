Implemented react/jsx-no-undef source analysis using handed JSX nodes and existing raw checker facts; shared registration remains pending.
Implementation SHA: 661963afd; base contains current main c01907a7 and named harness ab70f38d4.
Checks: 53 controls/29 findings and both frozen corpora match Go; ASan/UBSan/LSan and released query PASS; lint zero, format idempotent.
Mutant: component-name RegExp first class inverted; compiled and exited 0 with empty stderr, caught only by Go finding bytes.
Not covered: shared registry checker context, source Node/emitted-JavaScript comparison, full repository gate, other two JSX analyses.

The analysis consumes the opening or self-closing node handed to visit. It never
reads that node's kind or refetches it. Tag-name children are read to identify the
reference and its exact span. Intrinsic/custom elements, namespaces and this roots
are declined; member references follow the leftmost object. Existing binding-origin
returns raw symbol identity/declaration records. Symbols count as in scope under the
same-file declaration, allowGlobals or CommonJS rules used by production Go.
No new bridge question or shared file is changed.

The component-name predicate is the fixed JS RegExp literal /^[^a-z][^-]*$/u,
created once at module initialization. This expresses Go's existing first-byte/dash
predicate; it is not a Go regexp compile site and has no shared-table row. There is
no matcher fallback or pattern-option dialect here. UTF-16 offsets are converted
only when constructing the finding through the existing byte-offset table.

The private suite selects each node once and hands it to this listener. Its
central switch follows the current shared driver's string-kind interface; the
listener itself performs no kind comparisons. Numeric kinds 286/287 remain in
this claim's rule.json/listeners.a and are independently held to Go registrations.
The private suite is a verification runner, not a new shared dispatch edit.

The independent oracle invokes the unchanged Go production rule with a separately
loaded checker program and preserves complete diagnostics, fix and suggestion
fields. Fifty table controls are extracted from the Go test source, including
positive and silent cases, exact member-reference spans, type-only declarations,
imports/import-equals, global options and declare-global. Three added controls
cover an astral character before a finding and two CommonJS inputs. Total: 53
controls, 29 findings, zero fixes and zero suggestions, normal/sanitized bytes
identical to Go. The frozen 287 repository roots and 77 compiler roots again match:
18,485 and 5,241 bytes, zero findings. These corpora contain no positive JSX
findings; positive control coverage above is essential.

Every final native sanitizer run exits 0 with empty stderr. The semantic mutation
changes the first RegExp class to [a-z]; the mutant compiles, exits 0 without
stderr, and differs only under independent finding comparison. Querying a released
binding-origin program still panics 70. No new registry-retention mutation was
needed for an unchanged ABI; its five existing checks passed in the preceding
harness rebase. The initial gate's mutant import rewriting accidentally rewrote
'adamic' as a path and failed compilation; it was fixed and that failure is not
counted as a killed mutant. Its log is preserved. An intermediate range-based
predicate was replaced with the fixed RegExp, and the complete final gate was
rerun after formatting. Only the final RegExp version is certified here.

Commands source /workspace/adamic-tools/env.sh, using the compiler and checker
archives freshly validated on the current-main plus named-harness base:

```sh
python3 stage1/cohere/typeaware/react-jsx-no-undef/verify.py \
 --artifacts /workspace/wave20-validation/undef-regexp-final \
 --compiler /workspace/wave20-validation/harness-third/adamic \
 --checker /workspace/wave20-validation/harness-third/checker.a \
 --sanitized-checker /workspace/wave20-validation/harness-third/checker-asan.a \
 > /workspace/wave20-validation/undef-regexp-final.log 2>&1
```

The existing isolated cohere formatter checked eight owned Adamic files (including
the two new modules) idempotently. Only new source formatting changed. The existing
configured Go lint adapter checked the two new .a files, using virtual TypeScript
names without physical .ts modules, and prints findings 0. No shared formatter,
harness or generator workaround was edited. Toolchain reused: prior setup 121s,
Go 1.27.1, clang 20.1.8, Node 24.19.0, nproc 5/quota four cores. No fresh setup or
full repository gate is asserted. Every process output was saved to a file.

Quiet three-round alternating whole-process medians after builds and sanitizers:
repository native .365531s / Go .216216s; compiler native 1.819600s / Go .365020s.
Every timed pair matches complete output bytes. Native remains slower, about 1.69
and 4.99 times respectively; no dispatch speedup is claimed.

## Shared certification boundary

The named RuleContext provides parser, scanner, settings and findings, but no
checker program/lease, config or complete root manifest. This body needs that raw
binding fact access. The existing type-aware Rules context provides it, so this
analysis can be tested faithfully in the private typed runner. A shared factory
cannot obtain the same program from RuleContext without a shared integration
change, which this worker is forbidden to make. This is the actual remaining
shared-harness gap; absence of JSX parsing or of handed-node visitors is no longer
a blocker. The earlier numeric-only explanation was too broad: a listener that
consumes its handed node does not need an entry-kind guard to implement this body.

The shared registered module and source-Node/emitted-JavaScript fact transport are
not present, so this is explicitly an analysis implementation rather than a fully
registered shared rule. Once a checker lease is exposed, the adapter belongs in
stage1/cohere/lint/rules/react-jsx-no-undef under CLAUDE's registry contract.
No zero-finding placeholder, fabricated checker or partial shared descriptor is
registered. The other two claimed JSX analyses remain unimplemented; their binding
and construction/stability work is still outstanding. Three HIR/SSA claims remain
parked. No additional rule was claimed. Existing nine completed rules are unchanged
and retain the last full rebase gate results. Only wave-20 is pushed.
