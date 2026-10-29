Mutant evidence for #jk7tph7 B1's checker record in proofs/checks.json.

Mutant: undo both adapted return expressions in src/compiler/sys.ts, leaving
System's two signatures unchanged, and serve head A's corrected tty.d.ts.
The main control is byte-identical to with-92 except for these two expressions,
as TestTTYSourceScope checks across all 746 source files. The served text was
exported by load.CorrectNodeHostAxioms at head A 92d1abe8; head-a.json pins it.

Named check: TestTTYHostAxiomTypecheck in proofs/check.cjs.
Command (from the repository with stage3/api's locked TypeScript on NODE_PATH):

    node stage3/adapt/92-tty-host-results/proofs/check.cjs --require-repaired /tmp/tty-adaptation/control

Result with the mutant: exit 1. Restored, the same command with the after tree
exits 0 and prints PASS TestTTYHostAxiomTypecheck. Both the ordinary and stricter
Adamic checker options are exercised. Real failing lines from the mutant run:

    src/compiler/sys.ts:1531:13: error TS2322: Type '() => number | undefined' is not assignable to type '() => number'.
      Type 'number | undefined' is not assignable to type 'number'.
        Type 'undefined' is not assignable to type 'number'.
    src/compiler/sys.ts:1534:13: error TS2322: Type '() => boolean | undefined' is not assignable to type '() => boolean'.
      Type 'boolean | undefined' is not assignable to type 'boolean'.
        Type 'undefined' is not assignable to type 'boolean'.

The full assertion and named stack are in logs/undone-checker.log.gz. The restored
run is in logs/restored-checker.log.gz. No mutation is left in either built tree.
