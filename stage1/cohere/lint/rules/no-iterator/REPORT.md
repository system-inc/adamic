Built: no-iterator is registered on the unified harness with a handed ParseNode listener, .a modules, verbatim messages, precise human-only suggestions and positive raw witnesses.
Commits: based on lint-area d3a37422, then advanced to current main 4e0bfda5 after its runtime merge. Only this rule directory is added on codex/lint-port-no-iterator.
Checks: lint-registry, gofmt, registry tests, vet, TestRulesAgree, all 41 TestMutants and TestOwnedWitnesses pass on current main. The complete lint-area run also passes.
Mutants: changed static-key predicate compiles and exits cleanly on Node, emitted JavaScript and sanitized native; independent Go bytes catch it on both bases. A compiled nil-options adapter mutant hits the unchanged shared guard.
Not covered: the full root/17-input gate is not run. Existing malformed upstream recovery cases retain the harness's explicit refusal checks. Prior wave 06 checker/HIR work is parked outside this branch.

The source is the dedup ledger's batch2:no_iterator.ts, checked against pinned cohere's no_iterator.go and property.AccessedName. The latter unwraps parenthesized subscript expressions; the port does likewise. A computed identifier, declaration, substituting template and similar-but-different template key stay silent. Dotted, quoted, cooked escaped, parenthesized and optional accesses fire. The Unicode witness checks byte conversion at the shared finding serializer. Eight findings fire in the owned witness. Suggestions replace only the tail from the receiver's end through the access's end, preserving the receiver. They do not become automatic fixes.

The descriptor uses node: true and the two named AST access kinds. The driver selects relevance and hands the node to the rule. Its kind comparison distinguishes dotted-name semantics from computed-key semantics. Child lookups classify the key, and the root is not fetched explicitly in the rule. No shared context, registry generator, harness, upstream rule body or compiler file is edited.

upstreamTest is TestNoIterator, capturing every actual test: TestNoIteratorFires, TestNoIteratorStaysSilent, TestNoIteratorDeclinesASubstitutingTemplate and TestNoIteratorSuggestsTheRightSpan. The shared capture reports 2,027 unique source/rule/options combinations. Main's TestRulesAgree compares 13,056,587 bytes on Go/source Node/emitted JavaScript/ASan+UBSan native, passing in 65.45s. TestMutants passes all 41 mutations in 1,128.49s; owned witnesses match 98,430 bytes in 30.24s. Package total is 1,224.195s. The preceding lint-area run passes in 873.994s, with 98,606 owned-witness bytes; temporary filenames differ between runs, so byte equality is asserted within each run rather than across runs.

NoIterator has no upstream options type and ignores options. Its adapter nevertheless decodes and retains supplied JSON, instead of returning nil and triggering the shared guard. Empty/null input preserves upstream defaults. testdata/witness.options.json contains {} for the incoming witness-options support; the tested base does not yet consume sidecars. A manifest explicitly carrying {} in field 5 is separately compared on unchanged unified Go, source Node and freshly rebuilt main-based sanitized native: 4,427 identical bytes with empty stderr. The adapter mutation changes its initial condition to if len(fields) >= 0, making it drop every option bag. The overlay build succeeds; executing that field-5 manifest exits 2 with no-iterator: options {} reached an adapter that decodes none. The guard is unchanged. The default witness also runs on emitted JavaScript through TestOwnedWitnesses.

Toolchain setup completes: Go 0s, clang 0s, Node 0s, submodules 0s, warm cache 49s, total 49s; nproc=5, cgroup CPU 400000/100000 and memory 17.6 GB. Go 1.27.1, clang 20.1.8 and Node 24.19.0. Obsolete compiled parked-wave scratch artifacts were removed to provide build space, preserving their sources and logs. An initial untyped empty edit array was refused as an array of never; an explicit readonly SuggestionEdit[] corrects it. Superseded/aborted logs are retained as such and are not counted as successful complete gates.

Commands, each with stdout/stderr sent to its named log:

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > /tmp/no-iterator-setup.log 2>&1
go run ./cmd/lint-registry > /tmp/no-iterator-main-registry.log 2>&1
gofmt -l stage1/cohere/lint/rules/no-iterator/oracle.go > /tmp/no-iterator-gofmt.log
go vet ./stage1/cohere/lint/... > /tmp/no-iterator-main-vet.log 2>&1
go test ./stage1/cohere/lint/registry -count=1 > /tmp/no-iterator-main-registry-tests.log 2>&1
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestMutants|TestRulesAgree)$' -count=1 -v -timeout=30m > /tmp/no-iterator-main-harness.log 2>&1
go -C cohere test ./internal/lint/rules/core -list '^TestNoIterator' > /tmp/no-iterator-upstream-testnames.log 2>&1
```

The options probe uses the same explicit-file Go overlay as goOracleFrom in lint_test.go, substituting only this adapter for the guard mutant. Build the normal sanitized driver with go run ./cmd/adamic build stage1/cohere/lint/main.ts -o <scanner> --sanitize after registry generation. Its manifest row is <copied witness.ts>, no-iterator, three empty/legacy columns ending in false, and {} in field 5, separated by tabs. Compare oracle, source Node and native stdout, requiring zero exits and empty stderr. The archived options-mutant.go.txt records the sole Go mutation. evidence/index.json records uncompressed byte counts and SHA-256 hashes for gzip logs.

Step 1: all earlier wave 06 rules depend on missing production checker/context/link/oracle wiring or source-to-HIR/SSA/capture analysis, so the permitted all-blocked landing step was skipped. Their names and reproducer are pushed in stage1/cohere/typeaware/claims/wave-06-PARKED.md on codex/typeaware-wave-06. None of that branch's private drivers, checker questions or parked modules is brought into this landing branch.
