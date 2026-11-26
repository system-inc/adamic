# for...of object patterns: the pin's wrong versions (#rm22wzv)

`TestForOfObjectPatternReadsAsADeclaration` holds `testdata/for_of_object_pattern/getters.ts` to Node on sanitized
native, release native and the JavaScript backend. Its getters log each read, so the order and count of reads per step
are in its output.

1. **destructureFrom reading a pattern's bindings in reverse** (`internal/lower/collections.go`, its binding loop run
   from the last element to the first, each binding keeping its index). The C compiles, and each step runs `b`'s getter
   before `a`'s. Red at `TestForOfObjectPatternReadsAsADeclaration/getters.ts`, "stdout differs": the native run prints
   `read b of one` first, where Node prints `read a of one` (Loom's J-1wty85y at wip f6ba0c39).

2. **The viewed-array use left to the identifier road alone** (`internal/lower/object.go`, forOf's `recordViewArrayUse`
   condition back to `ast.IsIdentifier(name)`). The pattern loop over a checked view's array with no element check
   lowers and reads the lie raw. Red at `TestForOfObjectPatternRecordsAViewedArrayUse/object_pattern` in
   internal/lower, which wants "a checked view's array read through for-of (not checked yet)" (Lower's must on
   #rm22wzv). The array's element union is widened with a program class arm, held as Node is: a Map arm, as the
   first cut had it, is held differently, so destructureFrom refused the pattern ("destructuring a union of differently
   held members") before the record was read, and the pin was red without the mutant (lower-whole J-m2gs8j4 at
   8bdc794d). Which arms reach the record was probed on J-f2021sz.

A wrong version of the loop's own code fails to compile on native instead of reaching the pin, so none of them is
the pin's evidence: the destructure prefix dropped (J-k6fpmgz), run after the body (J-2zncxz9), or run twice
(J-ds2j2xc) each leaves clang a name read before, or declared twice beside, its declaration. That is the loud failure
this road wants, never a silent one.

`road_destructure_for_of.ts`'s outcome in `source_slot_certificate_test.go` moves from its NotYet to its declaration
twin's checked stop, measured on J-k6fpmgz: `stops: adamic: panic: field read failed: child.other is not
initialized`, with its reads at lines 8 and 9 checked. Its repaired twin agrees with those reads unchecked.
