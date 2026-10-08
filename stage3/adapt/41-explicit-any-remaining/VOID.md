# Historical void proposal

Superseded on October 8: void is a compiler lesson, admitted on compiler area-next
with Node agreement. All rewrites are dropped; tsc source stays as written. Any
void site still stopping after area-next lands is a compiler bug. See README.md.

The superseded proposal adapted all 15 syntax sites: eight call statements and seven arrow bodies
in checker.ts, moduleNameResolver.ts, utilities.ts and transformers/declarations.ts.
Calls in expression statements keep the call and discard its result. Arrow
bodies become statement blocks, so callbacks still return undefined, including
when push returns an array length. The constant void 0 arrow becomes an empty
body. The removed void-rules.json recorded every original line, expression and AST parent.
The adapter requires those exact owners before planning any writes.

The Node proof runs each original and adapted expression with effect-recording
callees, compares evaluation order, object writes and callback results, and
checks second applications leave source bytes unchanged. A mutant returning
`diagnostics.push(diag)` fails the undefined-result comparison.

The stable latent census sees four void refusals before and zero after. The
real source mutant restores only moduleNameResolver.ts:1838:35; its census has
exactly that one void refusal. All 15 syntactic expressions are independently
accounted for, including those in checker-diagnosed bodies skipped by the census.

The full final oracle matches main: 106,366 passing, one sanctioned API
acknowledgement failure, zero pending, and exactly the same api/typescript.d.ts
baseline diff bytes. No upstream baseline is accepted or changed by this unit.
The separate landing lane is recorded in REPORT.md when complete.

One earlier census raced with tree preparation and saw the pre-void source.
It was discarded; only the stable rerun is retained as final evidence.

The removed adapter and expression proof can be inspected in Git history before
the batch 4 ruling; the old receipt remains evidence/void-proof.log.txt.


The local expression proof does not replace the full upstream oracle. Native
compiler execution and unobserved lowering beyond diagnosed bodies are not claimed.
