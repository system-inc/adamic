# Library Error facts, namespace merge and truthful casts

Compiler scratch 2d7c31aca3a5dcda0ed82a868431a9f3af336777 starts from main
ffe6efc1 and cleanly merges area/library b05a9306. No compiler merge is pushed.
Namespace-read e1ebac45 conflicts in 16 files, including lowering and native;
merge was aborted as instructed. Three isolated namespace refusals persist.

Both native modes on the final audited slice stop at core.ts:11:52 before enum
initialization. Error's checker diagnostic is gone; its isolated probe now needs
a runtime descriptor for ambient method presence. C and native acceptance are
unreached. Fifteen ordered source stops, exact diagnostics, throwing placeholders
and discovery-only patch are retained. The discovery walk uses the prior validated
slice; the final 76 slice is rebuilt separately. Nothing behind a stub is a native
acceptance claim. The 12 isolated probes all run on Node and all refuse natively.

The baseline lane's adaptation, apply, oracle, lane and API paths have no diff
against origin/main; library's additional host evidence is immaterial to that
pipeline. Both full lanes pass with 106366 passing, one identical sanctioned API
failure, zero pending. Their API baseline diff and ten built JavaScript artifacts
are byte-identical; no API sanction is added. Local lane tests: 32 pass.

76 applies only the finite six-key assertion-cache alias. Its writer and
getOwnKeys mutants are rejected before writing; it is idempotent and both comment
modes erase identically. sameMap's truthful T|U proposal is not applied: stock
checker reports builder.ts:564 TS2322. Widening its private helper exposes three
DiagnosticMessageChain.next assignments at 545, 551 and 555. The negative string
assignment rejects the truthful union; replacing it with the old lying signature
makes that assignment compile, caught by the contract check. The proposal's
function JavaScript is byte-identical. No hiding cast is added.

Final slice: 27 declaration files, 1993 code declarations, 41857 copied-span lines,
2082 byte-audited spans, 79 ordered module import lists. Full and slice Node dumps
are identical: 36429231 bytes, SHA256
686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615.
Node-end and both JSDoc mutants are caught on both trees. This unit does not rerun
the existing 10406-case recovery manifest; no native binary exists to run it.

Commands: stage3/lane/run.sh for each full lane; check.cjs against lane-76's
adapted tree; stage3/slice/run.sh with the seven driver roots, verify.cjs; parser
run.sh against /tmp/parser-adapted10; measure-builds.py with jobs=5, timeout=60;
parser-front32-discover.py, parser-front32-minimals.py. Exact invocations and
results are in reports and compressed logs. nproc=5. Two Go compiler builds ran
out of disk; after deleting only disposable failed trees and cleaning Go cache,
GOWORK=off GOTMPDIR=... go build -p 1 -mod=mod succeeded. No compiler source edit.
The scratch go.mod only redirects cohere SDK paths to an existing pinned copy.

An accidental proof-checkout library merge was quarantined on local branch
codex/stage3-parser-front32-library-scratch (7493fa715). The proof branch was
recreated from its published 6c34bb145 tip and main was merged normally; the
library commit is not its ancestor. No pushed history was rewritten.
