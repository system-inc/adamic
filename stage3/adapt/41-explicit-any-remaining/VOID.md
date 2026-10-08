# Discarded void expressions

All 15 syntax sites are adapted: eight call statements and seven arrow bodies
in checker.ts, moduleNameResolver.ts, utilities.ts and transformers/declarations.ts.
Calls in expression statements keep the call and discard its result. Arrow
bodies become statement blocks, so callbacks still return undefined, including
when push returns an array length. The constant void 0 arrow becomes an empty
body. void-rules.json records every original line, expression and AST parent.
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

```sh
source /workspace/adamic-tools/env.sh
NODE_PATH="$HOME/.cache/adamic-stage3/api/node_modules" node stage3/adapt/41-explicit-any-remaining/void-proof.cjs FINAL_TREE > void-proof.log 2>&1
```

The local expression proof does not replace the full upstream oracle. Native
compiler execution and unobserved lowering beyond diagnosed bodies are not claimed.
