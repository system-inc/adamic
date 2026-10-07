Implemented all six pending claimed rules in .a, including complete owned edits and suggestions; the shared bridge still explicitly refuses four unsupported repair shapes.
Commits: c5a6f4cb constraint; 679c1c91 as-const; 0cbffc40 enum; 7bc069c8 namespace; 6bed7076 cache; 1e90906c query; final evidence commit follows these.
Commands: owned validate.py passed four-way fixture/corpus comparison, six semantic mutants and throughput; final supplemental fixtures were 83,165 identical bytes and corpus 74,551,560 identical bytes.
Mutants: unknown-constraint detection, literal equality, enum suggestion value, Node alias expansion, invalidate callee name and GraphQL binding resolution all compiled and were caught solely by output comparison on all three port executions.
Not covered: default shared-driver integration/all-rule dispatch, its report formatting and converging fixer, full gate, stage1 CLI self-lint and the earlier Tailwind JSX gap; no new rules claimed while that claimed-rule completion bar remains open.

## What changed

This completes the owned implementation work for the three original slot-15 claims and the three last claimed TypeScript rules. The two network rules can publish through the existing finding bridge. Node namespace prefix fixes, unnecessary constraints, as-const annotations and enum suggestions retain complete proposals in owned immutable Result/Edit/Suggestion records and refuse explicitly when the current shared bridge cannot carry them. No repair is silently dropped. Source, registration descriptors, unchanged Go adapters, messages, raw witnesses and compiling mutant specifications are directory-owned. All new Adamic implementation files are .a. Complete-proposal records are reused across this unit's owned directories.

- No-unnecessary-type-constraint preserves bare any/unknown syntax, parameter-name ranges, independent suggestion edit intervals, and filename-sensitive generic-arrow comma repairs, including comments, defaults and existing trailing commas.
- Prefer-as-const preserves variable/property/as-expression anchors, literal kinds and cooked/canonical values, postfix-token exclusions, non-repairable binding patterns, single as-expression fixes and both annotation edits.
- Prefer-enum-initializers preserves member positions rather than computed enum values, whole member text, quoted/computed names and all three suggestion messages and replacements.
- Node namespace imports preserve all built-ins, three subpath aliases, module expansions, exact resolved policy messages, two independent findings, specifier-only prefix fixes and the absence of binding-rename fixes.
- Cache invalidation preserves grouping removal, property-chain matching without a receiver-name restriction, first arguments and per-element literal/template findings with their upstream spans.
- GraphQL strings preserve four method names, first arguments, grouping/assertions, direct variable-statement initializer search outward from each reference, and cycle termination. The scope search follows Go exactly; it is not a general checker/binding bridge.

## Validation contract

The shared harness branch codex/lint-harness-dot-a was not present on the origin fetched for this run. No shared generator, harness, formatter, parser, compiler or submodule source was edited, including scratch copies. validate.py builds an entirely owned comparison entry point and an independently written serializer of unchanged Go Diagnostic objects. The Go oracle selects unchanged upstream rule values and walks the upstream AST; the source port uses the stage1 parser. The owned protocol compares every finding's byte range, id and exact message; every fix interval and replacement; every suggestion id/message and every edit; applied suggestions on fixtures; and combined safe-fixed source. This is complete-proposal parity, not a claim that the current shared formatter/fixer integration works.

For the large corpus, --proposals-only suppresses repeated whole-source copies after each suggestion, retaining every proposal and the combined safe-fixed source. The first version repeated whole compiler files and produced 1.4 GB; that superseded run was stopped and is not counted as a pass. Corpus fixtures have all raw edit ranges and replacements checked; fixture comparisons also check full applied suggestion source. Original invalid computed enum-key input is retained, matching Go's recovered AST rather than deleting the case.

The first native build refused nested mutable proposal arrays as cycle-capable. Rewriting proposal fields and result collections as immutable arrays removed that refusal. Empty typed-array constructors work around stage 0's array-of-never limitation. Final native and emitted JavaScript builds and executions succeeded. Native correctness uses ASan/UBSan/LeakSanitizer and requires empty stderr and exit 0. Source Node and emitted JavaScript are also required to exit 0 with empty stderr for successful comparisons and mutants.

Pinned cohere: 715ba94f3608a6500086b1076ce5cb7e51b836db. Pinned TypeScript v6.0.3: 050880ce59e30b356b686bd3144efe24f875ebc8. Toolchain Go 1.27.1, Node 24.19.0, clang 20.1.8; nproc=5.

Reproduce from the repository root after sourcing /workspace/adamic-tools/env.sh:

```
python3 stage1/cohere/lint/rules/typescript-no-unnecessary-type-constraint/validate.py --scratch /tmp/wave15-owned --typescript /path/to/pinned-TypeScript-v6.0.3 --mutants --throughput > /tmp/wave15-validation.log 2>&1
```

The final run used /tmp/lint-wave1-15-typescript. Additional exact filename-sensitive/manual-assertion vectors were added to the driver afterward and checked with --skip-build --fixtures-only; the semantic implementation was unchanged and the passed corpus was not needlessly repeated.

All selected upstream Go test families passed without modifying their harness, using its existing COHERE_DOCS_CAPTURE environment hook. Distinct captured original cases: constraint 37, as-const 69, enum 21, namespace 15, cache 16, query 23, total 181. With 19 explicit filename/scope/Unicode/numeric/manual-assertion probes and six owned witnesses, final fixture comparison covers 206 rows and 83,165 bytes. Corpus comparison covers 77 compiler and 150 stage1 .ts/.a files: 227 files, 1,362 file/rule pairs and 74,551,560 bytes. Capturing only asserted cases omits several upstream manual assertion calls; the exact missing filename-sensitive and message/span vectors are included explicitly in the supplementary probes.

## Mutants

Each owned mutant specification changes a real rule computation. The driver copies the stage1 module tree unchanged to scratch, mutates exactly the owned rule.a, builds successfully, executes source Node/emitted JavaScript/sanitized native successfully, requires empty stderr, and requires only the output bytes to differ from Go. All eighteen comparisons caught their mutant:

| Rule | Mutation | Compared effect |
| --- | --- | --- |
| no-unnecessary-type-constraint | UnknownKeyword replaced by NeverKeyword in the predicate | Missed unknown constraints and false never findings |
| prefer-as-const | Literal text equality inverted | Missing restatement findings and wrong literal findings/fixes |
| prefer-enum-initializers | Second suggestion position+1 changed to position+2 | Wrong suggestion message, edit and applied source |
| import-require-node-namespace | NodeFileSystem expansion changed | Wrong alias findings/messages and expected alias wording |
| network-no-invalidate-cache-literal-key | Callee invalidate changed to invalidateNever | Missed findings |
| network-no-string-literal-query | Initializer resolution replaced with -1 | Missed hoisted-string findings |

These supersede the earlier Go-only blocker-probe mutants for this implementation's validation. The older probe files and reports remain historical evidence. The independent oracle sentinel `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m` passed in 0.396s with one native/Node miss each; its compiling one-byte mutant proves the base comparator can fail.

## Shared integration gaps

Default `go run ./cmd/lint-registry` exits 1 requesting next-no-assign-module-variable/rule.ts. Default lint package compilation fails at profile_test.go:32 because portFiles is a function. The registry test package also fails when it discovers those .a descriptors through the unchanged .ts-only generator; it is not reported green. Four rules' positive bridge probes explicitly exit 70 on source Node, emitted JavaScript and sanitized native:

- Constraint suggestions edit name.End():constraint.End() rather than the finding name span.
- Annotation fixes require two edits.
- Enum diagnostics require three alternatives.
- Node-prefix fixes replace the module specifier rather than the whole import finding.

The complete information exists in owned Result objects for integration to consume. The current registered visit refuses those proposals through publish. Both network rules' bridge probes exit 0 with empty stderr on all three executions. The existing Finding has a single interval/replacement and suggestion string; its formatter/serializer and fixer need a complete repair contract before these four rules can run through the standard driver. Inference: changing only a rule directory cannot certify default combined dispatch and serialization under that contract.

The earlier Tailwind candidate still has one explicit JSX parser refusal before listener dispatch. It was already pushed and documented in its directory and the previous continuation report; this unit cannot repair the shared parser. No further rules are claimed because all previous claims have not reached the full integrated completion bar.

Setup reported Go, clang, Node and submodules ready at 0s each, then failed test-cache warming at the inherited profile compilation error; no final warm/done timing was emitted. Sourcing the actual environment path permitted all owned builds and isolated checks. Full gate, default shared report/fix parity, all-rule dispatch and stage1 CLI self-lint were not certified. `git diff --check` passed. No PR.

## Findings per second

Best of three interleaved end-to-end count-only runs over 77 compiler files and one 1,000-construct positive fixture per rule. Startup, reads, parsing and complete proposal construction are included; builds and rendering/applying repairs are excluded. Native timing uses the same program without sanitizers at the compiler's default optimized build; correctness above uses sanitized native. All three counts matched. These are corpus-plus-positive-fixture rates, not isolated visitor benchmarks.

| Rule | Findings | Native /s | Node /s | Go /s |
| --- | ---: | ---: | ---: | ---: |
| no-unnecessary-type-constraint | 1000 | 931.55 | 1143.95 | 5600.16 |
| prefer-as-const | 1000 | 923.43 | 1180.08 | 5479.11 |
| prefer-enum-initializers | 1680 | 1561.24 | 1963.43 | 8059.21 |
| import-require-node-namespace | 2000 | 1772.95 | 2346.52 | 11511.35 |
| network-no-invalidate-cache-literal-key | 1000 | 903.32 | 1156.60 | 5693.89 |
| network-no-string-literal-query | 1000 | 846.00 | 1119.35 | 5186.13 |

Evidence is under completion-evidence/. Fixture outputs and mutant outputs are retained; the large Go corpus output is gzip-compressed, with byte counts and SHA-256 hashes for all compared corpus outputs. validation.log and supplemental-fixtures.log record the completed checks and rates. Setup/default failures, bridge refusals, upstream test logs and the sentinel log are included. Test output was written directly to files, never piped.

## Published harness update after the implementation push

After pushing 03d9d672, origin exposed codex/lint-harness-dot-a at
2650ad595b82220c368631ea13139fad4b306ed6. This supersedes the earlier observation
that the branch was not yet published. Its unmodified source adds .a discovery,
emitted JavaScript comparisons, profile compilation fixes, independent fix
ranges and complete suggestion serialization. No shared source was adopted or
edited on this worker branch. The default failure logs above are observations
of this branch's existing foundation, not assertions about the new branch.

The owned probe-next-harness.py builds the exact published serializer through
a Go overlay, changing only its entry-point name, and runs unchanged Go rules.
The constraint and enum witnesses now serialize successfully with exit 0 and
empty stderr. The as-const annotation witness still exits 2 with
`panic: unexpected fix shape`: the published serializer permits exactly one
automatic fix, while Go supplies the two separate annotation edits. That is
the concrete remaining repair-model gap on the published next harness. The
probe logs and stderr are in completion-evidence/next-harness-*.

Reproduce after fetching that commit:

```
python3 stage1/cohere/lint/rules/typescript-no-unnecessary-type-constraint/probe-next-harness.py --scratch /tmp/wave15-next-harness > /tmp/wave15-next-harness.log 2>&1
```

Integrating the new shared suggestion/range model requires adapting the owned
publish bridge on that foundation. Complete owned proposals and their byte
comparisons are already pushed; integrated all-rule parity is still not
certified. No more rules were claimed while the current claimed-rule bar,
including the independent Tailwind JSX parser gap, remains open.
