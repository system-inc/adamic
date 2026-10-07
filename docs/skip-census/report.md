Built: AST census for 51 skip sites, classified table, scanning test and failing log checker.
Commits: implementation 451669af74ef5d2b2f478012a6addb86b2d29db5; base bf35c1cbf0cfba856abb081d094da7a1946c8e4d.
Commands: package suite and go vet pass; historical checker exits 1 with skips=33 required-input=17 unknown=0.
Mutants: added t.Skip, allow-added, allow-removed and allow-required all caught; condition-change probes also caught.
Not covered: full gate, cohere-owned tests, indirect helper calls, or provisioning the missing correctness corpora.

The branch is devtools/skip-census, created from origin/area/developer-tools as the unit requested. Only a new package and new docs files are changed. The default fetch refspec did not fetch the area branch, so it was fetched explicitly.

The table has 33 required-input sites, 15 measurement sites and 3 not-applicable sites. Runtime events and source sites have different counts: one site can skip several subtests. The JSON choice, identities and AST limitations are described in [skip-census.md](../skip-census.md). The setup-worker list is [gate-inputs.md](../gate-inputs.md).

Setup printed Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 125s, done 125s on 5 processors. Its env file was /workspace/adamic-tools/env.sh. Setup overlapped the checkout from b8fb957a to bf35c1cb; load before/after was not captured for setup, so these are operational timing lines, not a controlled comparison. No setup failure occurred.

Validation commands and observed verdicts:

```sh
go test -v -count=1 ./internal/skipcensus/... > /tmp/skip-census-tests.log 2>&1
# exit 0; seven test functions passed, including the exact historical skip list
go vet ./... > /tmp/skip-census-vet.log 2>&1
# exit 0; no diagnostics
go vet ./internal/skipcensus/... > /tmp/skip-census-vet-final.log 2>&1
# exit 0 after the final test additions
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle -run '^TestNativeAgreesWithNode$/^dedication$/^dedication.a$' > /tmp/skip-census-oracle.log 2>&1
# exit 0; dedication oracle passed; native hits=0, node hits=0, probe hits=0
gofmt -l internal/skipcensus
git diff --check
# no output
```

Proofs used the complete original gate-out/test.jsonl from the plain log branch, not just the reduced regression fixture. Origin gate-logs/2e165469ec95/plain tip was ae33bd09df0ebdf7363a203cee2dda6c804d483b. SHA-256 plain.tgz: ee51ba7b85f619ead2205bb86257ab0e33b6ec8391e98ba5876dd1644caf9150. SHA-256 extracted test.jsonl: 2970c041f6a38dc5c78b2ece350be84f3b14cd06faa2196b8423cbad22ea66be.

Reproduce the full-log proof and implementation mutants:

```sh
git fetch origin gate-logs/2e165469ec95/plain:refs/remotes/origin/gate-logs/2e165469ec95/plain
git show origin/gate-logs/2e165469ec95/plain:plain.tgz > /tmp/skip-census-plain.tgz
mkdir -p /tmp/skip-census-plain
tar xzf /tmp/skip-census-plain.tgz -C /tmp/skip-census-plain
go build -o /tmp/skip-census-command ./internal/skipcensus/cmd
go test -c -o /tmp/skip-census.test ./internal/skipcensus
python3 internal/skipcensus/testdata/proofs.py > /tmp/skip-census-proofs.log 2>&1
```

The proof script copies owned test source into /tmp/skip-census-source-copy, adds TestUndeclaredWitness containing a new t.Skip, and runs the real compiled TestCensus with ADAMIC_SKIP_CENSUS_ROOT naming that copy. It requires exit 1 and the undeclared witness name. It also builds independent source mutants in scratch packages:

| Mutant | Check that caught it |
| --- | --- |
| Add a new t.Skip in a scratch copy | TestCensus: undeclared TestUndeclaredWitness, exit 1 |
| Allow required-input runtime skips | TestLogClassesAndMutants: required-input skip survived, exit 1 |
| Ignore new source skips | TestASTAndDriftMutants: new-skip mutant survived, exit 1 |
| Ignore removed source skips | TestASTAndDriftMutants: deleted-skip mutant survived, exit 1 |
| Change an existing guard in the source fixture | TestASTAndDriftMutants: old identity removed and new identity undeclared |

The suite additionally rejects unknown/ambiguous skips, malformed JSON logs, duplicate declarations, invalid classes, absent providers and stale metadata. Moving source lines passes; testing import aliases and local receiver aliases are recognized; comments, string literals and non-testing Skip methods do not become rows.

Historical checker command: `/tmp/skip-census-command /tmp/skip-census-plain/gate-out/test.jsonl`. It exits 1 and prints exactly `skips=33 required-input=17 unknown=0`. The required-input test names are:

- `TestSplitTSGoAgrees`
- `TestThePortParsesAsGoCohereDoes/PostCSS`
- `TestThePortAnswersAsGoCohereAndGitDo/catches_R2_the_size_limit_one_byte_lower`
- `TestThePortParsesAsGoCohereDoes/as_graphql-js`
- `TestUpstreamNumericSeparatorGap`
- `TestExternalComparisonCatchesThreePrinterMutants`
- `TestUpstreamRepositoryCorpusParity`
- `TestCompilerAndStage1Agree`
- `TestCSSPrinterAgreesWithGo/default`
- `TestCSSPrinterAgreesWithGo/narrow`
- `TestCSSPrinterBoundaryProofs`
- `TestThePortParsesAsGoCohereDoes/as_postcss-media-query-parser`
- `TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars`
- `TestThePortParsesAsGoCohereDoes/as_postcss-selector-parser`
- `TestThePortParsesAsGoCohereDoes/as_postcss-values-parser`
- `TestCompilerExpressionsAgree`
- `TestWholeCompilerAgrees`

The full classified output is [skip-census-plain-check.log](skip-census-plain-check.log); scratch and implementation evidence is in [skip-census-proofs.log](skip-census-proofs.log) and the individual mutant logs in this directory. The checked-in testdata/plain-skips.jsonl preserves the last two output events and terminal event for each of the 33 named skips, retaining original event values. TestHistoricalPlainSkips asserts the exact required-input set.

The following is a mode comparison, not an implementation speedup claim: no census command existed before this unit. Both modes use the same implementation commit, same box and prebuilt executable, with three interleaved rounds, alternating order. Before sets ADAMIC_GATE_UNCACHED=0; after sets it to 1. There is no census cache in either mode, so no cache-key mutant applies. Every stdout byte, stderr byte and exit status matched across all six runs for each loop.

| Loop | Before, best of 3 | After, best of 3 | Instrument |
| --- | ---: | ---: | --- |
| AST scan | 0.079793 s | 0.079197 s | `/tmp/skip-census-command -scan` |
| plain log check | 0.132312 s | 0.132440 s | `/tmp/skip-census-command /tmp/skip-census-plain/gate-out/test.jsonl` |

Build-flags line for both timing rows: commit 451669af74ef5d2b2f478012a6addb86b2d29db5; nproc=5; cgroup cpu.max=400000 100000; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node v24.19.0; ordinary Go build with no additional build flags; no census result cache in either mode; prebuilt command.

| Loop and mode | Load before (1/5/15 min) | Load after (1/5/15 min) |
| --- | --- | --- |
| AST scan before | 0.18 0.60 0.74 | 0.18 0.60 0.74 |
| AST scan after | 0.18 0.60 0.74 | 0.18 0.60 0.74 |
| plain log check before | 0.25 0.61 0.74 | 0.25 0.61 0.74 |
| plain log check after | 0.25 0.61 0.74 | 0.25 0.61 0.74 |

All samples and flags are in [skip-census-paired.json](skip-census-paired.json). The skip checker is an additive callable command; wiring it into existing gate code would leave this unit's territory. This unit does not provision inputs, repair the existing skips, run the full gate or assert that optional checks without Skip calls were exercised.
