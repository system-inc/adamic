# Range and case helpers

Three separate Adamic .a helpers port the pinned cohere syntax/table operations, not regex matching.

- joinRanges(atoms, unicodeMode) returns ordered atoms and an error string. Rune/dash/rune triples become ranges; reversed endpoints fail in both modes. Other dash triples stay separate under Annex B and fail under Unicode. Atom fields are readonly value projections; callers preserve the original order and supply complete fields. The atom kind numbers are the private Go classAtomKind enumeration, not AST kinds.
- caseExtras(atoms, unicodeMode, groups, atomCovers, literalRune) delegates equivalence-group selection, coverage and literal formatting. It visits groups in their supplied order, skips groups without a covered member, and appends only uncovered members in group order. Supplied callbacks are the actual external dependencies, not invented matching behavior.
- buildCaseTables(unicodeMode, caseRanges, canonicalize, simpleFold) records every range member and canonical member, walks each complete simple-fold orbit in Unicode mode, and seeds the three pairs absent from Go CaseRanges. It deduplicates, retains groups of at least two, sorts each group numerically, and gives every member the same shared readonly array as its group. Supply pinned complete Go Unicode ranges and canonicalization/fold behavior; malformed or nonclosing orbits are outside the contract.

Go map traversal does not specify group-list ordering. The returned contents, member order and shared member-array identity are compared after ordering groups by their first member solely in the test observation. The helper itself leaves group traversal in map order. caseExtras is compared byte-for-byte against the exact group order supplied by that oracle run. Go may supply a different order in another process; the raw corpus captures each run's order.

The private Go overlay leaves production cohere untouched. It renames only dependency calls inside the original caseExtras and buildCaseTables bodies; joinRanges runs unchanged. The oracle parses all four consumer test files with Go's parser, captures their string values, and uses them as raw atom sequences and candidate class bodies through actual Go classAtoms. This is bounded helper input generation; it does not parse those strings as complete rule programs or run complete lint findings. Dependencies are adapters supplied from real Go; they are not newly implemented shared helpers.

With the toolchain environment sourced, run:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch21 -count=1 -v -timeout=20m > /tmp/lint05-batch21-helpers.log 2>&1
```

Source Node, emitted JavaScript and ASan/UBSan native must match actual Go. Every native baseline and independent mutant must exit zero with empty stderr; compile errors, sanitizer errors and panics receive no semantic-mutant credit. Compressed raw corpora, exact expected output and mutant witnesses are retained in evidence. See REPORT.md for commands, measured results and limits.
