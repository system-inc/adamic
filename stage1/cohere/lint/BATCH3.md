# Remaining frequency-ranked syntax rules

Branch `codex/stage1-lint-batch3` starts at scanner tip
`0090256e607c3f2de7d5b67cef680ec010f95c1d`. The twenty selected names and their
frequency counts are in [VOLUME.md](VOLUME.md); REPORT.md on that tip links the
continuation but does not itself contain the pending list. This batch implements
its remaining ten names. Forty rules now have implementations on this branch.
Eight additions have full captured own-fixture coverage; the two identifier
rules have the JSX parser dependency below.

The other worker's `origin/codex/stage1-lint-batch2` was initially unpublished.
Its subsequently fetched tip `c4373c05259c7235cf1202d5cd4138c16d498caa` implements
twenty unrelated rules and explicitly excludes the frequency-selected twenty.
Its registry and BATCH2.md were checked again before this commit. No overlap
was found, so none of these ten was skipped and that branch was not merged.
No compiler, runtime, scanner, parser or submodule source changed.

## Implementation and coverage

Each rule is in its own module under [rules/](rules/), with its own activation
and exported visitor. [registry.ts](rules/registry.ts) has one call per family.
The shared acyclic RuleContext supplies the parser, scanner, parent indexes,
findings and decoded settings. Parent links are numeric, not owning cycles.

| Rule | Unique captured Go cases | Ported behavior |
| --- | ---: | --- |
| `one-var` | 340 | scope modes, initialization, consecutive blocks, loop and export guards, require groups, multiple repairs |
| `nexus/consistency-no-ambiguous-identifier` | 33 | identifier roles, catch/sort/handler exemptions, exact Unicode simple folding; four JSX cases excluded |
| `nexus/consistency-no-abbreviated-identifier` | 183 | ordered whole/prefix/suffix/segment policy, imported and declared names, rest-args exemption; five JSX cases excluded |
| `@typescript-eslint/no-non-null-assertion` | 36 | assertion positions, optional-chain suggestions, all suggestion edits |
| `@typescript-eslint/prefer-enum-initializers` | 21 | initializer sequence and all three alternatives with complete edits |
| `nexus/consistency-no-long-line-comment` | 16 | reachable consecutive comment groups, option/directive exemptions, safe block rewrites |
| `nexus/consistency-no-multiline-arrow-function` | 40 | arrow syntax, callback/hook/handler exemptions, safe function-expression rewrites |
| `prefer-destructuring` | 113 | declarations and assignments, nested decoded options, receiver/member cases, preserved source |
| `nexus/consistency-no-single-line-jsdoc` | 17 | reachable JSDoc trivia, line/URL exemptions, safe line-comment rewrites |
| `nexus/consistency-no-shouting` | 57 | comment tokens, code/quoted spans, acronym/currency/user allowlists |

These are 856 new captured combinations. The total capture is 2,129 unique
source/rule/decoded-options/filename combinations. Filename is part of the
capture key and preserved during fix reparsing, so TSX cannot silently become
ordinary TypeScript. Seventeen combinations are explicit parser-limit checks:
nine new JSX fixtures and eight inherited malformed method-signature fixtures.
They contribute no bytes to the successful parity totals. Seven admitted
malformed fixtures remain findings-only, as documented in VOLUME.md.
The existing baseline and volume generated corpus is retained. Twenty-eight
new controls exercise all ten families, nested scopes and options, false token
punctuation in comments, Unicode folding/quotes, every ECMA line ending,
multi-edit suggestions, and a real upstream fix cycle.

The oracle calls unmodified pinned Go rules, real `report.Write`, and real
`edit.FixText`; the test overlay only captures inputs. Original Go assertions
still run during capture. The port is compared with that oracle both on Node
and as native code under ASan/UBSan with Linux leak checking. Node here runs
this same TypeScript port; it is not an independent ESLint oracle.

Finding descriptions, IDs, human positions, UTF-8 ranges, first and additional
automatic edits, and complete suggestion payloads for the two new TypeScript
suggestion rules are compared. Suggestions remain unapplied. One-var's real
Go fix cycle exposed exhausted-budget behavior: Go returns the original input,
not the partially fixed result. The port now prints the exact rejection and
unconverged rules and returns the original input after ten passes too.

Policy JSON vocabulary and shouting allowlists are generated static data via
[testdata/generate_batch3_data.py](testdata/generate_batch3_data.py); one-var
messages are copied static Go messages. Neither contains computed verdicts.
The policy-data generator was rerun in an isolated scratch tree, formatted
with pinned Cohere, and compared with `diff -u`: no differences. Its formatting
output is saved in `data-generation.log`.
To regenerate that data from the repository root:

```bash
python3 stage1/cohere/lint/testdata/generate_batch3_data.py
/workspace/scratch/cohere --format-only --no-cache stage1/cohere/lint/rules/policy_data.ts
```

Fixed-shape Go regex checks use literal scans and existing Unicode simple-fold
tables on this baseline. Arbitrary configurable regex matching is not claimed.
Abbreviation binding sets are constructed once per file, rather than searching
every declaration for every identifier. Comment positions use one ECMA line
index and binary search. Fix application joins accepted source spans and edits
once per pass, preserving the existing overlap selection and ordering.

## Final verification

Successful raw logs are in [batch3_evidence/](batch3_evidence/). The commands
below were run in `/workspace/adamic` after
`source /workspace/adamic-tools/env.sh`; test output was redirected directly to
files, never piped. The TypeScript corpus is v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`, verified by the test. Cohere is pinned
at `715ba94f3608a6500086b1076ce5cb7e51b836db` throughout.

```bash
bash cloud/setup.sh > /tmp/lint-batch3-setup.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestCompilerAndStage1Agree|TestBatch3Controls|TestJsxParserGap|TestNestedConstructorGap|TestOptionAndComparatorGaps)$' -count=1 -v -timeout 30m > /tmp/lint-batch3-release.log 2>&1
go test ./stage1/cohere/lint -run '^(TestBatch3Mutants|TestBatch3RepairMutants|TestMutants|TestDecorationOptionMutant|TestCountGuardMutant|TestCommentFoldMutant|TestPositionIndexMutant)$' -count=1 -v -timeout 30m > /tmp/lint-batch3-final-mutants.log 2>&1
go test ./internal/oracle -run '^(TestNativeAgreesWithNode|TestTheOracleCatchesOneByte)$/^internal$/^oracle$/^testdata$/^(strings|maps_and_text|indexing|visits|sorting)\.a$' -count=1 -v -timeout 10m > /tmp/lint-batch3-core-oracle.log 2>&1
go vet ./... > /tmp/lint-batch3-final-vet.log 2>&1
/workspace/scratch/cohere --no-fix --no-cache stage1/cohere/lint/rules stage1/cohere/lint/{finding,lint,main,settings,repair_edit,repair_suggestion}.ts stage1/cohere/lint/gaps/{6_regex_literal,7_jsx_parse}.ts > /tmp/lint-batch3-source-gate.log 2>&1
gofmt -l stage1/cohere/lint
git diff --check > /tmp/lint-batch3-diff-check.log 2>&1
```

Final release: **32,052,531 identical canonical bytes** across all three runners:

- 1,321,368 bytes for captured fixtures and the combined generated corpus.
- 30,689,034 bytes for all 77 pinned compiler files and 112 current stage1
  `.ts` files: 189 files total.
- 42,129 bytes in the independent 28-control gate.

The complete release selection passed in 479.052s; its large corpus passed in
385.99s. The stricter final standalone upstream rerun passed in 60.994s with
1,321,368 identical bytes and all seventeen excluded cases explicitly refusing.
Final mutation selection: **PASS, 904.004s**, all **34 mutants** caught on
both Node and sanitized native: ten new family checks, three new repair checks,
and twenty-one inherited checks. No compile/refusal/sanitizer failure was counted
as a killed mutant. Final vet, gofmt and staged whitespace checks exit0 with
no output.
The filtered core oracle passed in 26.414s: five Node/native/JavaScript fixtures
and the one-byte negative control. Vet and whitespace checks print nothing.
The source and format gate checks 276 rules, 29 checked, 100% Adamic-ready
(27 of 27 selected files), no findings. No full repository `go test ./...`
gate, throughput benchmark, allocation profile or instruction profile is
claimed. The untouched ten-family `TestVolumeMutants` suite is not repeated;
the original seventeen plus decoration/count/fold/position checks are repeated.
Optional historical profiling/snapshot jobs are not run.

## Mutants

Each final mutant must compile and complete normally on Node and sanitized
native, then produce wrong output compared with Go. Compilation failures,
panics and sanitizer failures are not counted as caught mutants.

| Family | Mutation | What catches it |
| --- | --- | --- |
| one-var | disable its own activation | missing scope findings, edits and fixed source |
| ambiguous identifier | disable its own activation | missing identifier messages and ranges |
| abbreviated identifier | disable its own activation | missing policy findings and ranges |
| non-null assertion | disable its own activation | missing assertion findings and suggestion payloads |
| enum initializers | disable its own activation | missing findings and alternative suggestions |
| long line comment | disable its own activation | missing group findings, ranges and fixed source |
| multiline arrow | disable its own activation | missing arrow findings, ranges and fixed source |
| prefer-destructuring | disable its own activation | missing findings and source repair |
| single-line JSDoc | disable its own activation | missing trivia findings, ranges and fixed source |
| shouting | disable its own activation | missing uppercase-token findings and ranges |
| complete suggestion edits | change the second optional-chain edit from `?.` to `!!` | additional edit record; legacy first repair alone is insufficient |
| complete suggestions | drop the enum's third alternative | full suggestion list; legacy first repair alone is insufficient |
| fix-cycle rollback | return partially fixed source after exhaustion | original-source comparison against the actual upstream cycle |

The thirteen new mutants are accompanied by the inherited seventeen finding
and fixer mutants listed in REPORT.md, plus decoration range, count-only,
Unicode fold and comment position mutants described in PERFORMANCE.md and
VOLUME.md. Every changed answer is saved in the raw mutation log. The count-only
mutant changes the count while ordinary output remains identical.

## Observed limits and initial failures

- Eight inherited malformed method-signature combinations remain bounded
  parser failures. Go recovers; the port panics or fails to terminate within
  two seconds. The existing EOF-loop reduction is unchanged.
- Nine JSX identifier fixtures require missing parser coverage. The standalone
  [JSX proving program](gaps/7_jsx_parse.ts) finds a JsxElement under Go TSX,
  but neither Node nor native stage1 finds one. The lint driver explicitly
  refuses the type-assertion-shaped TSX tree. This dependency is not hidden as
  clean output, and parser sources remain untouched.
- [The regex proving program](gaps/6_regex_literal.ts) prints `true` on Node;
  stage0 refuses RegularExpressionLiteral on this scanner baseline. Fixed
  literal matching is the rule-local workaround. GAPS.md separates these
  observations from claims about current main.
- A first standalone Go JSX proof used a relative filename, which Cohere's
  parser rejects. It was corrected to an absolute TSX path.
- Early large-corpus runs exposed quadratic comment-position and repeated
  full-source repair copying. One native execution hit the harness's ten-minute
  limit; later stale Node/native executions were explicitly stopped. They are
  failed or cancelled runs, not parity evidence. The final run uses the line
  table and one join per repair pass; no speedup ratio is claimed.
- The first chunk-join implementation used a two-value `push`, refused by
  stage0. Two single-value pushes compile. An inherited position mutant's
  original anchor became nonunique after adding the comment collector; its
  anchor now identifies the original warning-comment path specifically.
- A first fix-cycle mutant survived because the seed's initialized and
  uninitialized options were reversed. The seed was replaced with actual
  upstream invalid161 options (`Uninitialized: Always`, `Initialized: Never`),
  and only its subsequently caught result counts. Early command flag and Go
  variable-binding errors were corrected before final verification.
- General binder/type-checker rules, config validation, suppression, JSX,
  general parser recovery, non-UTF-8 source and filesystem integration are
  outside this unit. Ordinary JSDoc trivia is covered, not a JSDoc node parser.

Toolchain setup succeeded and printed:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (1s)
setup: build cache warm (25s)
setup: done in 25s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Direct `nproc` prints **5**. Versions: Go1.27.1, clang20.1.8, Node24.19.0.
