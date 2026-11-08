# A–C at 3ac38c67

Replacement head: 3ac38c67f2745703f82644d9116240a03ec48e64. It adds the exit twin to f95deb36's chdir row.

A: gofmt exit 0, empty output; vet exit 0.
B: combined generic-method/NodeWatch pins exit 1. NodeWatch tests pass. Chdir and exit both fail for .a and .ts with Load / TS2591, Cannot find name process.
C: removing exactly ` && l.generic(source)` and rerunning generic-method pins exits 1. The same four rows are red with the same loading failure; they are already red in baseline, so this is not a discriminating green-to-red mutant and does not establish the guard. All other generic-method subtests pass.

Exact patches, full logs, red-row indexes and complete load messages are adjacent. Byte-for-byte source restoration and whole-worktree Git diff both exit 0. No permanent compiler/test changes were made.

The initial f95deb36 run had already started when the replacement arrived; its logs are retained separately under evidence/f95deb36/. Both chdir extensions fail with the same TS2591 there, in baseline and mutant. Steps D/E remain at the originally requested f95deb36, as stated before those runs began.

Results were reported immediately after C, including the full chdir failure messages.
