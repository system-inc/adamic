# Cohere rule inventory

Run from the Adamic root after `bash cloud/setup.sh` and sourcing its environment:

```sh
go run ./stage1/cohere/lint/inventory -compiler /path/to/TypeScript > inventory-run.log 2>&1
```

The compiler checkout must be v6.0.3 commit `050880ce59e30b356b686bd3144efe24f875ebc8`. Fetch the four named remote branch refs before running; absent refs are recorded as unavailable. The command writes `inventory.json`, `inventory.md`, build/generation/test logs and the raw harness-invocation ledger. It builds an overlay inside cohere because Go's internal-package boundary forbids importing its rule registry from Adamic. It does not edit the submodule or apply fixes.

`-capture-tests=false` reuses the existing raw capture ledger in the output directory. `-measure=false` leaves measurements unknown. For a metadata-only refresh use both flags and `-reuse /absolute/path/to/previous/inventory.json`; the generator refuses different cohere commits or rule sets. Reuse records the previous corpus, so changed repository bytes require a new measurement. Outputs describe the pinned scanner checkout, not later main.

The registry is the completeness oracle. Go type resolution binds function identities, receiver methods, sibling calls, callback function values and registration decoders. Reachable cohere declarations are traversed transitively. External upstream methods are recorded at their API boundary. This is a conservative source graph, not observed runtime call coverage; unresolved interface dispatch can hide an edge. Private unreachable helpers are excluded. Same-stem companion Go files count as rule sources; other sibling helpers are separately named dependencies. Line counts include blank and comment lines.

The test overlay records every entry into cohere's common capture harness, with its nearest rule-test caller and line. These are executed invocations, not unique fixtures, subtests or assertions. Passing families are distinguished from partial failing families. Direct Context tests and external corpus checks can exist without a captured invocation; their files remain visible. Inspect the test log before treating a count as coverage.

Frequency runs execute real Go cohere rule callbacks with a shared TypeScript program, default/nil rule options, and no suppressions or project lint configuration. This measures the potential findings of every applicable rule, including currently disabled rules. The synthetic compiler configuration is strict ESNext/Bundler with JSX preserved. Required-options rules have no default measurement. Panics invalidate a rule's entire corpus count. `.a` files are parsed as TypeScript for syntax measurements, but have no checker bridge; counts needing their checker are unknown. Parse diagnostics are explicit exclusions. Zero means a completed measurement returned no diagnostics; it does not certify semantic parity or that every interesting subject shape occurred.

Regenerate the derived pipeline with `python3 stage1/cohere/lint/inventory/testdata/analyze.py` after regenerating the inventory.

The analysis and worker handoff are in [PIPELINE.md](PIPELINE.md) and [PROMPTS.md](PROMPTS.md). The rule-by-rule wave, named blockers, checker API questions and branch evidence are generated, not hand-maintained.

Validation:

```sh
go test -count=1 -v ./stage1/cohere/lint/inventory > inventory-tests.log 2>&1
python3 stage1/cohere/lint/inventory/testdata/mutants.py /tmp/inventory-mutants > inventory-mutants.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v > inventory-oracle.log 2>&1
```

Mutants use disposable overlays and must compile and fail the named test. They never replace the baseline files.
