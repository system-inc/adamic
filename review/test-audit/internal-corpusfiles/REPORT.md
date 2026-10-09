Unit u017, origin/main 7b18d0576930caca4e22ce2eef92fcf563af52d0.
Five top-level rows, all enabled; clean baseline passed.
Fifteen fixed-menu mutants: twelve killed, three survivors.
Four sacred rows; Prettier row subsumed and vacuous.
Evidence: test-audit/internal-corpusfiles, review/test-audit/internal-corpusfiles/.

Code under test and oracle, declared before mutations
Repository, Upstream, checked, selectFiles, git, names, matches. Their Go implementation selects corpus files from Git and validates pins, roots, cleanliness and physical membership. No Adamic lowering, native product or stage1 port is involved. Git is run to establish repository state; expected selection, ordering and error substrings are self-written. No external-authority classification is claimed.

Scope
All five names are in the starting list.log. No Your rows section was supplied; the entire package was used. No moved/vanished row can be identified without an earlier list. ConvertedPackageContracts has eleven input subcases sharing one checker, recorded as one top-level family row. No top-level helper or harness witness is present. Planted corpus faults are inputs to the collector under test, not mutants of an independent agreement oracle. The collector is treated as this unit's production code rather than as another suite's shard/coverage construction.

[
  {
    "test": "TestConvertedPackageContracts",
    "package": "internal/corpusfiles",
    "file": "internal/corpusfiles/files_test.go:61",
    "seconds": 0.213,
    "oracle": "Git runs establish actual tracked, dirty, ignored, pinned and sparse states; self-written counts, order, equality and diagnostic substring assertions decide correctness. Diagnostic substrings are weaker than full text.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M03",
      "M04",
      "M06",
      "M07",
      "M08",
      "M14",
      "M15"
    ],
    "unique_kills": [
      "M07",
      "M08"
    ],
    "last_proven_fail": "M08 files_test.go:95: planted corpus fault survived",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 15,
    "probe_kills": [
      "PRepository",
      "PUpstream",
      "PSelect"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestConvertedPackageContracts",
      "TestMissingAndEmptyRoots",
      "TestSparseCheckoutCannotShrinkCorpus",
      "TestDeterministicRootsPatternsAndMissingGit",
      "TestProvisionedPrettierFixtures"
    ],
    "evidence": "GOWORK=off ADAMIC_CSS_FIXTURES=/tmp/u017-prettier ADAMIC_MUTANT=M08 timeout 120 go test -json -count=1 -timeout 90s ./internal/corpusfiles/ -run . > M08.log 2>&1; files_test.go:95: planted corpus fault survived",
    "members": [
      "scanner",
      "scanner/profile-runtime",
      "typeaware",
      "markdowninline",
      "yaml",
      "cssstrings",
      "cssnumbers",
      "graphql/printer",
      "markdownblocks/default",
      "markdownblocks/census",
      "selector"
    ],
    "vacuous_subcases": []
  },
  {
    "test": "TestMissingAndEmptyRoots",
    "package": "internal/corpusfiles",
    "file": "internal/corpusfiles/files_test.go:112",
    "seconds": 0.026,
    "oracle": "Git runs establish actual tracked, dirty, ignored, pinned and sparse states; self-written counts, order, equality and diagnostic substring assertions decide correctness. Diagnostic substrings are weaker than full text.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M04",
      "M05",
      "M06",
      "M10",
      "M14"
    ],
    "unique_kills": [
      "M05",
      "M10"
    ],
    "last_proven_fail": "M10 files_test.go:117: planted corpus fault survived",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 15,
    "probe_kills": [
      "PSelect"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestConvertedPackageContracts",
      "TestMissingAndEmptyRoots",
      "TestSparseCheckoutCannotShrinkCorpus",
      "TestDeterministicRootsPatternsAndMissingGit",
      "TestProvisionedPrettierFixtures"
    ],
    "evidence": "GOWORK=off ADAMIC_CSS_FIXTURES=/tmp/u017-prettier ADAMIC_MUTANT=M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/corpusfiles/ -run . > M10.log 2>&1; files_test.go:117: planted corpus fault survived"
  },
  {
    "test": "TestSparseCheckoutCannotShrinkCorpus",
    "package": "internal/corpusfiles",
    "file": "internal/corpusfiles/files_test.go:121",
    "seconds": 0.032,
    "oracle": "Git runs establish actual tracked, dirty, ignored, pinned and sparse states; self-written counts, order, equality and diagnostic substring assertions decide correctness. Diagnostic substrings are weaker than full text.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M03",
      "M04",
      "M09",
      "M14",
      "M15"
    ],
    "unique_kills": [
      "M09"
    ],
    "last_proven_fail": "M09 files_test.go:143: planted corpus fault survived",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 15,
    "probe_kills": [
      "PSelect"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestConvertedPackageContracts",
      "TestMissingAndEmptyRoots",
      "TestSparseCheckoutCannotShrinkCorpus",
      "TestDeterministicRootsPatternsAndMissingGit",
      "TestProvisionedPrettierFixtures"
    ],
    "evidence": "GOWORK=off ADAMIC_CSS_FIXTURES=/tmp/u017-prettier ADAMIC_MUTANT=M09 timeout 120 go test -json -count=1 -timeout 90s ./internal/corpusfiles/ -run . > M09.log 2>&1; files_test.go:143: planted corpus fault survived"
  },
  {
    "test": "TestDeterministicRootsPatternsAndMissingGit",
    "package": "internal/corpusfiles",
    "file": "internal/corpusfiles/files_test.go:147",
    "seconds": 0.036,
    "oracle": "Git runs establish actual tracked, dirty, ignored, pinned and sparse states; self-written counts, order, equality and diagnostic substring assertions decide correctness. Diagnostic substrings are weaker than full text.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M02",
      "M03",
      "M11",
      "M15"
    ],
    "unique_kills": [
      "M02",
      "M11"
    ],
    "last_proven_fail": "M11 files_test.go:156: selection: [/tmp/adamic-gate/TestDeterministicRootsPatternsAndMissingGit2198410340/001/inputs/z.TS /tmp/adamic-gate/TestDeterministicRootsPatternsAndMissingGit2198410340/001/inputs/tracked file.ts /tmp/adamic-gate/TestDeterministicRootsPatternsAndMissingGit2198410340/001/inputs/a.ts] / [/tmp/adamic-gate/TestDeterministicRootsPatternsAndMissingGit2198410340/001/inputs/z.TS /tmp/adamic-gate/TestDeterministicRootsPatternsAndMissingGit2198410340/001/inputs/tracked file.ts /tmp/adamic-gate/TestDeterministicRootsPatternsAndMissingGit2198410340/001/inputs/a.ts]",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 15,
    "probe_kills": [
      "PRepository",
      "PSelect"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestConvertedPackageContracts",
      "TestMissingAndEmptyRoots",
      "TestSparseCheckoutCannotShrinkCorpus",
      "TestDeterministicRootsPatternsAndMissingGit",
      "TestProvisionedPrettierFixtures"
    ],
    "evidence": "GOWORK=off ADAMIC_CSS_FIXTURES=/tmp/u017-prettier ADAMIC_MUTANT=M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/corpusfiles/ -run . > M11.log 2>&1; files_test.go:156: selection: [/tmp/adamic-gate/TestDeterministicRootsPatternsAndMissingGit2198410340/001/inputs/z.TS /tmp/adamic-gate/TestDeterministicRootsPatternsAndMissingGit2198410340/001/inputs/tracked file.ts /tmp/adamic-gate/TestDeterministicRootsPatternsAndMissingGit2198410340/001/inputs/a.ts] / [/tmp/adamic-gate/TestDeterministicRootsPatternsAndMissingGit2198410340/001/inputs/z.TS /tmp/adamic-gate/TestDeterministicRootsPatternsAndMissingGit2198410340/001/inputs/tracked file.ts /tmp/adamic-gate/TestDeterministicRootsPatternsAndMissingGit2198410340/001/inputs/a.ts]"
  },
  {
    "test": "TestProvisionedPrettierFixtures",
    "package": "internal/corpusfiles",
    "file": "internal/corpusfiles/files_test.go:162",
    "seconds": 0.052,
    "oracle": "Runs Git via Upstream and requires selection to succeed. Counts are only logged, never asserted; nil output passes PUpstream. This is a success-status oracle with no positive result assertion.",
    "oracle_kind": "external-run",
    "kills": [
      "M03",
      "M04",
      "M06",
      "M14",
      "M15"
    ],
    "unique_kills": [],
    "last_proven_fail": "M15 files_test.go:170: corpus-files: /tmp/u017-prettier root tests/format/css: no tracked files match [\"*.css\" \"*.scss\" \"*.less\"]",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestConvertedPackageContracts"
    ],
    "mutants_in_matrix": 15,
    "probe_kills": [],
    "subsumer_seconds": 0.213,
    "vacuous": true,
    "bounded": false,
    "matrix_rows": [
      "TestConvertedPackageContracts",
      "TestMissingAndEmptyRoots",
      "TestSparseCheckoutCannotShrinkCorpus",
      "TestDeterministicRootsPatternsAndMissingGit",
      "TestProvisionedPrettierFixtures"
    ],
    "evidence": "GOWORK=off ADAMIC_CSS_FIXTURES=/tmp/u017-prettier ADAMIC_MUTANT=M15 timeout 120 go test -json -count=1 -timeout 90s ./internal/corpusfiles/ -run . > M15.log 2>&1; files_test.go:170: corpus-files: /tmp/u017-prettier root tests/format/css: no tracked files match [\"*.css\" \"*.scss\" \"*.less\"]"
  }
]

Mutant table, every location against starting origin/main
M01 internal/corpusfiles/files.go:68: drop trailing NUL trim; 'strings.TrimSuffix(string(data), "\\x00")' -> 'string(data)'; failed []
M02 internal/corpusfiles/files.go:73: drop filename case folding; 'strings.ToLower(path.Base(name))' -> 'path.Base(name)'; failed ['TestDeterministicRootsPatternsAndMissingGit']
M03 internal/corpusfiles/files.go:74: change matching result constant; 'return true' -> 'return false'; failed ['TestDeterministicRootsPatternsAndMissingGit', 'TestSparseCheckoutCannotShrinkCorpus', 'TestProvisionedPrettierFixtures', 'TestConvertedPackageContracts']
M04 internal/corpusfiles/files.go:98: flip pin comparison; 'actual != pin' -> 'actual == pin'; failed ['TestProvisionedPrettierFixtures', 'TestMissingAndEmptyRoots', 'TestSparseCheckoutCannotShrinkCorpus', 'TestConvertedPackageContracts']
M05 internal/corpusfiles/files.go:107: disable missing-root guard; 'err != nil {\n\t\t\treturn nil, actual, nil, fmt.Errorf("%s: missing named root' -> 'false && err != nil {\n\t\t\treturn nil, actual, nil, fmt.Errorf("%s: missing named root'; failed ['TestMissingAndEmptyRoots']
M06 internal/corpusfiles/files.go:111: flip whole-repository pathspec condition; 'root == "."' -> 'root != "."'; failed ['TestProvisionedPrettierFixtures', 'TestMissingAndEmptyRoots', 'TestConvertedPackageContracts']
M07 internal/corpusfiles/files.go:119: disable upstream status guard; 'len(status) != 0' -> 'false && len(status) != 0'; failed ['TestConvertedPackageContracts']
M08 internal/corpusfiles/files.go:141: disable dirty-file guard; 'len(dirty) != 0' -> 'false && len(dirty) != 0'; failed ['TestConvertedPackageContracts']
M09 internal/corpusfiles/files.go:163: disable sparse missing-file guard; 'len(missing) != 0' -> 'false && len(missing) != 0'; failed ['TestSparseCheckoutCannotShrinkCorpus']
M10 internal/corpusfiles/files.go:166: disable empty-root guard; 'counts[i] == 0' -> 'false && counts[i] == 0'; failed ['TestMissingAndEmptyRoots']
M11 internal/corpusfiles/files.go:174: change sorting option; 'sort.Strings(files)' -> 'sort.Sort(sort.Reverse(sort.StringSlice(files)))'; failed ['TestDeterministicRootsPatternsAndMissingGit']
M12 internal/corpusfiles/files.go:160: change map value constant; 'selected[absolute] = true' -> 'selected[absolute] = false'; failed []
M13 internal/corpusfiles/files.go:161: off-by-one count increment; 'counts[i]++' -> 'counts[i] += 2'; failed []
M14 internal/corpusfiles/files.go:97: drop HEAD whitespace trim; 'strings.TrimSpace(string(head))' -> 'string(head)'; failed ['TestProvisionedPrettierFixtures', 'TestSparseCheckoutCannotShrinkCorpus', 'TestMissingAndEmptyRoots', 'TestConvertedPackageContracts']
M15 internal/corpusfiles/files.go:73: swap glob arguments; 'path.Match(strings.ToLower(pattern), strings.ToLower(path.Base(name)))' -> 'path.Match(strings.ToLower(path.Base(name)), strings.ToLower(pattern))'; failed ['TestDeterministicRootsPatternsAndMissingGit', 'TestProvisionedPrettierFixtures', 'TestSparseCheckoutCannotShrinkCorpus', 'TestConvertedPackageContracts']

Survivors
M01 unguarded helper parsing behavior: witness-clean.log names=["tracked.ts"], witness-M01.log names=["tracked.ts" ""]. Existing selection ignores that empty entry on the tested patterns; no end-to-end corpus-selection difference was demonstrated.
M12 equivalent candidate: selected[absolute] true becomes false, but only map keys are ranged. No differing observable output demonstrated.
M13 unguarded counts: witness-clean.log files=1 counts=[1], witness-M13.log files=1 counts=[2]. Full-package logs double the logged counts without failing.
Witness command: GOWORK=off go test -v -count=1 ./internal/corpusfiles/ -run '^TestU017Observation$'. New temporary observation harness is preserved in survivor-witness.go.txt and was removed after running; existing tests were untouched.

Probes
PRepository returns nil at Repository entry, PUpstream returns nil at Upstream entry, PSelect returns nil files, empty HEAD, zero per-root counts and nil error at selectFiles entry. Returning zero counts sized to roots keeps wrapper logging valid. Own entries: Converted and Deterministic call Repository and selectFiles (Converted also Upstream); Missing and Sparse call selectFiles; Prettier calls Upstream. The Prettier probe passes. No other own-entry probe passes its row. No positive converted subcase completed successfully under its probe. Probes never contribute kills or uniqueness.

Ambiguities and costs
The brief refers to Your rows, but contains no list. Used whole package.
The brief says whole families count as one row; this family is already subtests inside a single top-level Test, so its original top-level name is retained.
The setup-check boundary is ambiguous for a package whose reusable product is itself a test-input collector. These rows directly test collector behavior; normal mutant verdicts are used, with the interpretation disclosed above.
Warm Go works, but go.work references absent cohere/TypeScript/tsc. Initial go test -list failed before any test. GOWORK=off resolves this standard-library-only unit without provisioning unrelated submodules.
/usr/bin/time is absent; used Bash time and Python monotonic timing.
The first scratch switch generator embedded a return as an expression and overlapped match replacements. Its compilation failed; no results from that attempt count. All standalone mutants had passed vet. The corrected switch and valid matrix logs are saved.
Subsystem docs are large; initial combined reads were truncated by output limits. Unit tests and collector implementation were subsequently read whole.
No oracle is weakened because no row is an independent agreement witness. No existing test, Git executable or upstream source was mutated.

Timing and validation
Warm toolchain check timing was not captured; no setup.sh; Go 1.27.1, nproc=5.
Node npm ci: installer reports 709ms, three packages installed. No package test loads node_modules from another directory.
Prettier sparse provisioning completed between tool calls, approximately 19 seconds; no exact monotonic setup timer was captured.
Clean baseline Bash wall 0.448 seconds. Binary package time is recorded in baseline.log. No skips, panic, timeout or bounded narrowing.
Standalone mutants each passed GOWORK=off go vet ./internal/corpusfiles/, logs per id. Corrected switched binary compiled once; Go test runs reuse its build artifacts. No native rebuild needed. Individual timings and wall durations are in runs.json.
Subsumption rests on five caught mutants out of fifteen planted, not a deletion recommendation. Repo-wide uniqueness, other packages, malformed pins/root patterns and exhaustive glob semantics were not covered. Central replay can apply each M*.diff against the starting origin/main. Production source restored; final-baseline.log records the final green run.
Measured commands total 11.613s; timing commands 3.399s; standalone vet 1.860s; corrected switch build 0.275s; matrix/probes 6.074s.
