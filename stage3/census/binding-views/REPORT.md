# STOP: binding views changes upstream declaration output

Base: `7534c1d7575db02bb7b6512119e0685987b7e333`.
Scout: `4de50731d74928af5f4cf1e35e191efc13b31da0`.
TypeScript: `050880ce59e30b356b686bd3144efe24f875ebc8` (v6.0.3).

The user requires stopping if either upstream declaration output differs. Both
adapted checkouts ran `npm ci` and the unmodified `node_modules/.bin/hereby dts`
successfully. **typescript.d.ts differs**, so no implementation commit or branch
was pushed. No Loom submission occurred.

The [complete exact diff](typescript.d.ts.diff) is 5,691 lines. The output has
11,530 lines on each side; exactly 1,302 lines change, each solely by inserting
`export ` after its indentation. No other line changes. This is a byte-level
public declaration gate failure; no claim is made about semantic equivalence.

| Output | Baseline SHA-256 | Candidate SHA-256 |
| --- | --- | --- |
| typescript.d.ts | f03adcf915df559d616fcafd75181d3423eaa0f272237482cedbd025525c9516 | bfe13d0ab8055d651cff6bf4fe322d8fc8c7c923892108cf062a822c9ae8ab12 |
| tsserverlibrary.d.ts | cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd | cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd |

Both outputs are retained under baseline/ and candidate/ as gzip files. The plain
exact diff has SHA-256 `0087ce6cfae5b0397a59fd23b8e11ac916d26822f01c2bf3cb8a082c822b6c8c`.

## Proof status

| Proof | Result |
| --- | --- |
| Isolated base worktree, recursive shallow submodules, pinned upstream clone | Complete |
| Base stage3/apply.sh | Pass; apply-baseline.log |
| Declaration builds | Both succeed; byte identity fails for typescript.d.ts, passes for tsserverlibrary.d.ts |
| Type-only edited-file JavaScript, both comment modes | Pass inside contract controls |
| Owner drift, duplicate, partial adaptation, mixed line endings | Refused with zero writes; contracts.json |
| Adapter replay | Zero writes |
| CRLF resolution preservation | Pass |
| Mutant exact two-file restoration and replay | Pass; six returning refusals not measured |
| Full resolution census | Both runs stopped with SIGTERM after declaration failure, incomplete streams retained and labeled; no delta claimed |
| Program census (scan only) | Not run; wrapper provenance verified against promise-discard-sites 2dc0593b937f8fbaf6c32f8107a899f78c223966 |
| Stock tsc 79 roots | Not run |
| Adamic types | Not run; unmodified executable build succeeds |
| Candidate stage3/apply.sh | Not run |
| All-file emitted JavaScript comparison | Not run |
| Whole unfiltered lane | Not run |
| Generated files | Not regenerated |
| Ledger | None |

The first owner grep is retained. The diagnostics overloads collide with 95h's
whole createProgram pin and were prepared as an amendment there. Resolution's
edited owners have no whole-owner collision. The candidate applies the exact
scout patches, preserving CRLF; normalized LF strings are used for the final
file pins. Its original interface in program.ts is unchanged.

`candidate-adaptation-WIP.patch` preserves the unfinished adapter work for
inspection only. It has not completed all requested integration requirements
or proofs, and is not an accepted adaptation. The standalone 95h amendment was
not replayed end-to-end before the declaration stop. `RESULTS.json` gives
machine-readable proof status, including all unrun checks.
