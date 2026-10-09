StrictEq is defended by a package-unique production constant change.
LeavesCheckoutText is not uniquely defended after three honest attempts; its prior untrue finding is refuted by a shared production kill.
No test was modified, deleted, or weakened; all production changes are restored.

Main: b8bcadb2c493173855f19d7e5c508b34f5eeb5b6

CODE UNDER TEST: Go adaptation in adapt.go, including parseStatement, unchanged-source returns, inferred equality safety and replacement operators. ORACLE: self-written expected source fragments, exact unchanged strings and adaptation counts in adapt_test.go. No outside oracle was changed.

Coverage: each requested row and StrictEq subsumer TestClassifyAdapt have separate -coverpkg=./cmd/adamic-test262 profiles. LeavesCheckoutText was compared both with newer TestAdaptClassKeepsVars and the complete remaining package using -skip=^TestAdaptLeavesCheckoutText$. StrictEq has 31 blocks absent from ClassifyAdapt, including the inequality replacement at 1878; the two feed different operators, typeof and coercing/chained comparisons. Leaves has one block absent from the class guard (newline bookkeeping), and 0 blocks absent from the whole rest. Coverage is a lead, not the verdict.

Baseline: the whole package passed in 35.965 binary seconds, with 44 top-level results. Each production mutant ran the entire current package, including all three newer adapter guards. The unrelated TestCompilerStartupMeasurement opt-in skipped; its body builds/runs the compiler on program(Math.abs(...)) and does not call adaptation. Matrix records every executed passing, failing and skipped row. Each diff independently passed Go vet and compiled in its actual whole-package run; no panic or timeout occurred.

D1 changes != replacement from !== to ===, breaking only TestAdaptStrictEq. All 42 other executed rows passed, including ClassifyAdapt. Full passed names are in matrix.json. Its standalone diff is D1.diff and failing line is adapt_test.go:156. This is a real production replacement constant, not a test-specific selector.

D2 drops class-body skipping. Leaves passes and every other executed row passes. An independent source probe shows a real lost transformation: class Box {} var x=1; if (1 == 1) {} is rewritten to === clean, but stays == under D2. The survivor is an unguarded continuation after class input, not an equivalent candidate. Raw control/mutant output and observational main are retained.

D3 drops the class unsafe-vars marker. Leaves still passes because its class-only source contains no var rewrite. TestAdaptClassKeepsVars catches let x=1 in place of var x=1.

D4 reduces the implicit full-length bound of the unchanged-source return by one byte (saturating at zero to avoid an unrelated empty-source panic). It fails Leaves by removing its trailing newline, and also fails ClassKeepsVars and VarToLet. This is a nonempty real production mutation, not the audit empty-answer probe.

Friction and owner findings:
- A filesystem of total size 8.8 GB cannot have 15 GB free. Clearing the preceding unit cache recovered /tmp to 8.6 GB free; /workspace remained around 14 GB free. No workspace/tool content was removed and no disk failure occurred.
- The audit predates three adapter regression guards. The new class guard matters to the current matrix and was included.
- The brief allows a formerly untrue row with shared catches to become subsumed, but its final defense enum omits that value. rows.json uses not defended and names the actual shared catchers; the old untrue conclusion is explicitly refuted. This is not deletion evidence beyond the four shown mutations.
- LeavesCheckoutText names checkout-file preservation, but only compares an in-memory class source and zero counts. It never opens, reads, or checks a checkout file. Its assertion does protect unchanged returned text, as D4 proves.
- No executor twin or cost row is involved. StrictEq assertions directly check the named operator behavior.

Warm env.sh worked, setup skipped, nproc=5. npm ci in stage3/api succeeded before baseline. Mutant command walls (including vet) were 34.632, 34.559, 34.779 and 33.005 seconds; remaining-package coverage passed in 26.142 binary seconds. No external corpus or other package was run.
