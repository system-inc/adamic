# Fusion lists

Node's arm64 release is built by GCC 12, which fuses `a * b + c` into one multiply-add by default.
So on arm64, its last bits in Math's transcendentals and in parseInt with a radix other than 10 or a
power of two differ from V8's own build, from x86-64 Node, and from Adamic, which never fuses (#4z1nx1s).
Compiler and Correctness ruled, and Adamic confirmed:
- Adamic stays unfused on every target.
- The oracle's truth for arithmetic is V8's unfused result.
- On arm64, a case where Node differs is held to that result only if a closed, audited list names it.

Each `<name>.txt` holds one case a line, tab-separated:
- the sweep's question, as its harness reads it
- the unfused value native answers on every architecture
- arm64 Node's fused answer

`<name>-derived.txt` names IEEE 754 differences the fused twin doesn't reproduce, in functions ECMAScript
lets an engine approximate. Under Correctness's rule of record, confirmed by Adamic on #4z1nx1s, each one's
pinned value is native's unfused answer and equals x86-64 Node's, which the x86-64 run of the sweep
proves. arm64 Node's answer is one ulp away and logged beside it on every arm64 run. Fusion is inferred
(GCC 12), not reproduced by the clang twin, so #t3rw158's aarch64 GCC 12 may later move these onto the
audited list. TestDerivedFusionIsOneUlp keeps every derived entry exactly one ulp from Node's answer, so
a larger difference in one of these functions stays red.

`<name>-pending.txt` names differences no rule explains. It allows nothing: each one is red until it's
audited onto a list or fixed. The mechanism stays, empty, so a new unexplained difference is red.
fusion_test.go reads these files, and `go test ./internal/native -args -update-fusion`, run on arm64,
rewrites them.

| List | Cases | Audit | Where it runs |
|---|---:|---|---|
| parseInt.txt | 4,598 | v8ParseInt, V8's HandleGenericCase loop in Go, fused and unfused | every architecture, no Node |
| ieee754.txt | 1,406 | the fused twin (fusedTwin, Options.contract) gives arm64 Node's answer | arm64 |
| ieee754-derived.txt | 164 | pinned equals x86-64 Node; arm64 Node one ulp away; fusion inferred, not reproduced: log2 99, atan 28, log1p 17, asinh 12, atanh 6, atan2 2 | x86-64 proves the pin; arm64 logs Node's answer |
| ieee754-pending.txt | 0 | none | |
| parseInt-derived.txt, parseInt-pending.txt | 0 | none needed | |

On x86-64, Node must answer every listed or derived case's pinned value. A fused answer there is red.
That run is the independent proof of the IEEE 754 pinned values, which -update-fusion took from native's
own arm64 answers.

## Mutants

Each of these turns the named test red.

| Mutant | Red test | Where it was run |
|---|---|---|
| fusion_test.go handled: drop the arm64 condition, so a fused answer is explained anywhere | TestFusionListExplainsOnlyOnArm64 (amd64 case explained, no error) | Mac, pure Go |
| parseInt.txt: pin `9007199254740993` radix 36 to `44fa555c722220b4`, not `...b5` | TestParseIntFusionListIsV8sLoop | Mac, pure Go |
| ieee754.txt: pin `cos 1 7feffffffffffffe` to `3fd7ffdfb4c53091`, not `...90` | TestIeee754MatchesNodeBitForBit, twice, since the sweep asks it twice | Sun1's arm64 Lima guest, at 857ec3e4 |
| ieee754-pending.txt: delete `asinh 1 3fb10d67b5aca865` | TestIeee754MatchesNodeBitForBit: a plain mismatch, not a pending one | Sun1's arm64 Lima guest, at 857ec3e4 |
| ieee754-derived.txt: record `asinh 1 3fb10d67b5aca865`'s Node answer as `...3003f`, two ulps from its pin | TestDerivedFusionIsOneUlp | Mac, pure Go |
| fusion_test.go handled: a derived case explained on amd64 too | TestFusionListExplainsOnlyOnArm64 (derived amd64 row) | by its table: the amd64 row expects one error |
| fusion_test.go fusedTwin: build with `contract: false` | TestNumbersParseExactlyAsJavaScriptDoes and TestIeee754MatchesNodeBitForBit (no case explained) | Sun1's arm64 Lima guest, at 5b9965bb, the twin's first form |
