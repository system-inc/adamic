Built: hash-verified hidden-02 comparison for roadmap step 30; no compiler representation or any contract changed.
Commits: delivery starts at dcdbb909; existing generic-return dependency measured separately at 32b7eed5.
Commands: exact historical selector still mismatches; the replacement Object.entries selector reproduces; focused existing oracle and real-any contract pass.
Mutants: none added or credited by this evidence-only stop; no implementation check changed.
Not covered: Object.entries reference-value support, source adaptations, native compilation of the assigned region, or a full repository gate.

The check-first dependency is compiler/any-returns-next 32b7eed56c293e4c1cec049254db54ec0b38f0c9. It was measured in an isolated detached worktree and was not merged into this delivery branch. Both compiler worktrees use cohere 7945d102a6c18dd36adf9114a758ce646e8b2359. All 82 adapted source byte lengths and SHA-256 hashes match census pin 388096e6a83a4e9d287fb827f793c599ba1bf0ad.

The precise historical selector core.ts:1891:17 does not reproduce on either current compiler. It selects memoize alone at core.ts:1891:1 and reports `a value of type T` at 1892:9 and `a function returning T` at 1893:12. This is a signature mismatch, not a successful reproduction of the original caller-context stop. The old brief already records that mismatch. The full area census does reproduce the exact caller-context stop: utilities.ts:1384:1 has Boundary [48518, 64484), with diagnostic core.ts:1891:17 and `a function returning any`. This is the assigned head reproduction. It is not credited as a pass of the isolated selector.

On the existing fix the actual caller, utilities.ts:1384:1, stops at utilities.ts:1386:24:

```
Adamic 0.1 refuses Object.entries; tsc's result must have one homogeneous number, string or boolean value type; any and widened field views are unsound
```

Replay with `-kind Refused -reason Object.entries` succeeds. The explanatory fix text is not part of the reason selector. Its Boundary is exactly [48518, 64484). The hash-pinned stock AST contains only the getScriptTargetFeatures declaration unit intersecting this interval; it contains no independently attempted nested function declaration that could expose these bytes. Clearing the invented-any diagnostic therefore earns no revealed-byte credit for this assigned region.

The new first stop is a different contract, Object.entries on the feature table's array and Map values. This unit keeps that refusal and real any blocked. It does not turn either into an unchecked runtime any. Per the check-first instruction, the generic-return work is named as an existing dependency; real-any contracts belong to TypeScript's source adaptations. No new positive or negative oracle fixture was added, and counts.md is unchanged.

Validation observations:

- The existing any_returns_concrete.a fixture passes the focused uncached Node oracle, sanitized native build, JavaScript backend, release build and leak check. Only this fixture was selected.
- TestGenericResolvedReturnContracts passes and retains its named refusal for an explicit any return.
- The brief's exact explicit-any identity witness prints 7 on source Node, exit 0, and remains blocked by the area compiler with `a function returning any`, exit 1.
- No whole package test or full gate was run for this unit. Setup itself runs its prescribed cache warmup.

Setup was rerun on the stable area base with `GOPROXY=https://proxy.golang.org|direct`. It succeeded: Go ready 0.017s, Node ready 0.021s, submodules ready 0.071s, clang ready 0.161s, markdown dependencies ready 0.824s, Go build ready 47.796s, cache warm 47.992s, done 48.016s; nproc 5, cgroup quota four CPUs. The initial attempt overlapped checkout and failed cache warmup with missing ir.AllocateEnvironment, ir.Labeled and test helpers; that attempt is not claimed as validation. No module-fetch 403 occurred. Every shell sourced /workspace/adamic-tools/env.sh.

Commands and complete observations are retained in evidence. Both historical replay commands exit 1. The replacement replay exits 0. No push or merge into main or any area branch was performed.

Regional measurement (UTF-8 bytes; the source hash is ef43309e71a5bff946d868de2753b73de6b3cae07e5f215e2a2f1281ef0406ef):

| Compiler / evidence | Hidden intersection | Hidden bytes |
| --- | --- | ---: |
| Original census, ed6e2975 at pin 388096e6 | [48518, 64484) | 15,966 |
| Area base dcdbb909, completed fresh census | [48518, 64484) | 15,966 |
| Existing fix 32b7eed5, fresh exact-head replay | [48518, 64484) | 15,966 |

The difference from both the original and the area base is **0 bytes**. The existing fix removes this head's invented-any diagnostic but does not reveal its source. Its archived pre-merge full census independently has the same regional intersection. The fresh full comparison census did not complete: it left 68 records and no recoverable exit status. Its partial log is retained; no fresh corpus-wide remaining-any count is claimed. The archived comparison has 15 remaining boundaries: ten at explicit-any callees and five involving unresolved legacy Node imports. Those are prior evidence, not a new certification of the comparison tip. RESULT.json keeps archived and fresh evidence separate. Its combined any-reason count includes both ordinary and resolved-generic any diagnostics; it is not the brief's single-reason count.

This is an evidence-only delivery toward roadmap step 30. The requested generic-return implementation already exists on the named dependency. Object.entries reference-value support is the next distinct implementation task; it was not implemented here. No new check was added, so no new mutant result is claimed. No fixture or counts row changed, and TestCountsAreRecorded was not run. The exact original any-bearing witness remains refused. The requested new fixture and mutant are not delivered by this stop report.

Reproduce the arithmetic with `python3 stage3/hidden-02-any-return/measure.py /tmp/hidden-02-census-pin /tmp/hidden-adapted/src/compiler`, after preparing the frozen worktree and adapted source. The complete fresh area census and the archived comparison ledger are compressed in evidence; the fresh successful next-head replay is retained separately.
