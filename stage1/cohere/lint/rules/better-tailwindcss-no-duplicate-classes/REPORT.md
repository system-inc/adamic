Status: .a listener, template duplicate core and disjoint edits implemented; multiple-edit integration explicitly refuses.

The next paragraph records the earlier dependency-only investigation. Current results supersede it and are linked below.

Go uses the class-literal reader, template segments, and several disjoint deletions within one diagnostic. The shared Go harness testdata/oracle.go lines 79-82 rejects len(d.Fixes) != 1 with unexpected fix shape; Finding has only editStart, editEnd and one replacement. This is a shared harness gap. Flattening those edits into a whole-literal rewrite would change the exact fix contract and conflict behavior. The rule also needs the absent class-literal/template-segment reader. Work stopped on the no-unknown-classes dependency investigation under Ahra's instruction to stop on other blockers. No parity, mutant or throughput claim is made for this rule.

The dependency-only investigation above is historical. Current implementation, successful bounded comparisons, semantic mutants, exact limitations and retained logs are in ../better-tailwindcss-no-deprecated-classes/testdata/evidence/REPORT.md. It supersedes the earlier claim that no implementation or registration exists.
