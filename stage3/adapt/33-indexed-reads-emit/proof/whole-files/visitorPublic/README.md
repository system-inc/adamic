# visitorPublic.ts at zero

Integration input: 634ef061fc72c061e2de1606d5c8faebec4411f6, merged by fast-forward
into codex/stage3-indexed-reads-emit. TypeScript source pin and stock API: 6.0.3.
Setup: 97 seconds, nproc 5; individual readiness timings are in setup.txt.

The existing U-loop ledger entry changes from decline to assert. The loop bounds
prove parameters[i] is a member of the populated readonly NodeArray returned by
nodesVisitor. The assertion stays on that single read before its callback.

Unfiltered load.Load census: one TS2345 before, zero diagnostics of any code
after. verify.txt records byte-identical stock JavaScript and adapter contract
idempotence for all 30 partition files. The required-read mutant replaces this
! with ?? 0; mutant.json records both the emitted-JavaScript and site-contract
checks failing specifically on that change. The API validator mutant appends an
unlisted interface to the emitted declaration file; exact projection rejects it.
The emitted declaration file was restored byte for byte.

Default oracle: 106366 passing, 1 failing, 0 pending, 313.748 seconds. The only
baseline diff is api/typescript.d.ts: adaptation 20's 189 optional-property lines
against adaptation 40's already projected reference. No other baseline differs.
A read-only stock AST validator proves the complete emitted API equals pristine
plus exactly 189 adaptation-20 owner lines and 28 inherited adaptation-40 lines
(27 brand types and ErrorCallback.arg0). No baseline was accepted. The original
adaptation-20 validator rejects the composed snapshot because it permits only
its own edits against pristine. No adaptation 70 directory is in 634ef06.

The user's exception rule names only adaptations 20 and 70. Whether the inherited
40 reference projection can be retained has been asked explicitly; it remains
pending. This file is checker-clean but is not presented as oracle-green or
pushed. Later files have not been started under the one-file-at-a-time rule.

Meter provenance: 176a496, run 0 on integration 634ef06 plus five pinned compiler
features; meter-input.json contains its one visitorPublic finding. Its private
preparation resolver fails on an unrecognized internal/lower/enums.go conflict
when replayed against 634ef06. No resolution was guessed. Direct loader census
uses the unchanged integration compiler and has stricter style/syntax diagnostics
elsewhere than that latent configuration; global 36-to-37 progress is therefore
not claimed. Both configurations agree on visitorPublic's sole input finding.
