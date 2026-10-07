# 46: Fix empty single-quoted pragma arguments

This is a fork bug fix, separate from adaptation 45's soundness assertions.
It replaces only the `||` token in parser.ts:10731's argument selection:

```diff
- const value = (matchResult[2] || matchResult[3])!;
+ const value = (matchResult[2] ?? matchResult[3])!;
```

The matcher at parser.ts:10707 has alternative single-quote and double-quote
branches. Exactly one of groups 2 and 3 participates in a successful match,
and either participating capture can be the empty string. Nullish selection
therefore gives a present string, preserving emptiness. The assertion remains
on the combined expression; neither individual optional capture is asserted.

The stock typescript@6.0.3 AST locates getNamedArgRegEx, verifies its reviewed
constructor, locates extractPragmas's value, and requires adaptation 45's
checked boundary. Only its operator token changes. Unrecognized shapes fail
before writes. An existing `??` is a zero-edit pass. This layer is ordered after
45 and is also understood by 45 when both layers are reapplied.

Node counterexample: `/// <reference path='' />` passed to createSourceFile.
Before, value is undefined and parser.ts:10737 throws a TypeError reading
value.length. After, the reference path is `""` and its capture span length is 0.
Nonempty single/double quotes, empty double quotes, and no-match input retain
their previous results. test.cjs reads the actual matcher and selector from
source, checks those cases, confirms an operator-only diff and idempotence,
and kills a put-back `||` mutant with the empty-value assertion.

The real public-API before/after probes and the full default oracle are
recorded with the final evidence. The oracle runs from the requested
`a3ef0dc` base with adaptations 00, 10, 45 and 46, excluding adaptation 20.
Reference baselines are not changed or accepted.

The issue is drafted in [stage3/upstream/LEDGER.md](../../upstream/LEDGER.md).
Kirk chooses whether and when to file it. Nothing has been sent upstream.

Commands (all test output goes to log files):

```sh
source /workspace/adamic-tools/env.sh
NODE_PATH=/tmp/regex-captures-cache/api/node_modules node stage3/adapt/46-fix-pragma-empty-argument/test.cjs /tmp/regex-captures-upstream > /tmp/regex-captures-fix-tests-final.log 2>&1
```

The Node public-API probe runs stock npm 6.0.3 and the built fork compiler:

```sh
node stage3/adapt/46-fix-pragma-empty-argument/public-probe.cjs /tmp/regex-captures-cache/api/node_modules/typescript/lib/typescript.js /tmp/regex-captures-fixed/built/local/typescript.js > /tmp/regex-captures-parser-public-after.log 2>&1
```

It passes: before is the TypeError, after is
`{"referencedFiles":[{"pos":21,"end":21,"fileName":""}],"typeReferenceDirectives":[],"libReferenceDirectives":[],"diagnostics":[]}`.
Ten control inputs retain identical parser observations. The original-parser
put-back mutant fails the empty-path expectation. `logs/public-api.log`
records all of those observations. A separate unsafe-name mutant is rejected
by both adapters before writes, holding the fixed-name regex assumption.

## Final default oracle and idempotence

The complete default oracle passes with 106,367 passing, 0 failing, 0 pending;
all phase exits 0. `evidence/report.json` has runners=all, tests=null, workers=4.
`evidence/baseline.diff` is **0 bytes**. Install/build/tests took
2.924s / 34.124s / 572.645s, total wall 609.787s. The final full suite was run
on 10+45+46 from `a3ef0dc`; no baseline was updated. No reference baseline
covers the prior crash in the exercised default suite.

The upstream-config census on that built tree has 0 checker diagnostics,
compared with the exact 29 on its 10-only control. Reapplying 45 and 46 after
the suite gives 0 files/edits in each adapter and identical SHA256 maps of
82,834 tree files, excluding .git and node_modules. Raw census records,
independent audit, provenance, commands, limits and idempotence evidence are
in adaptation 45's README and evidence directory.

Native tsc and the full Adamic gate are not claimed. The known import-cycle
and multiple-root lowering limitations remain outside these source units.
