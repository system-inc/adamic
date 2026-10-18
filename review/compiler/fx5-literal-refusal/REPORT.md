Built a narrow destination-layout refusal for nonempty union-array literals and three lowering regression leaves toward task #8ng1j76, item 130.
Implementation commit: 5f466b83; evidence commit: see branch compiler/fx5-literal-refusal HEAD.
Validation: focused fixtures, TestCallTargetReaders, full internal/lower, counts refresh and integration lane checks; outputs are beside this report.
Mutant: remove the refusal; the refusal fixture fails on nil error, and the compiled native oracle fails on empty stdout versus Node's number 1.
Not covered: actual conversion into declared union layout; no verified integration q12 cancellation witness found.

The guard is in elementType's nonempty contextual-array path in internal/lower/object.go. It checks the destination element representation for ir.Union, then refuses only when the initializer's representation differs. The literal diagnostic is main.a:1:36, with exactly "union-typed array built from a narrower literal" and the requested fix. Mixed [1, 'a'] remains accepted and is compared to source Node through both generated JavaScript and native execution.

The number[] value assignment does not take this literal path. The existing mutable-invariance pass refuses it first at main.a:2:36, saying "a value of type number[] seen as (string | number)[], which can write string | number where number is read" and naming adamic/invariant-mutable. Its regression checks this message and path.

The repository console declaration takes one argument, so fixtures print `${typeof items[0]} ${items[0]}`. Source Node prints `number 1`. With the guard removed, lowering succeeds, generated JavaScript agrees, native compilation succeeds, and the native stdout oracle reports `Native backend stdout = "", source Node = "number 1\n"`. This is an observed stdout disagreement, not a clang warning failure. revert-refusal.diff applies to the fixed object.go; restore with its reverse. For the native proof, temporarily replace the refusal test body after t.Parallel with lowersAndAgreesWithNode(t, fx5NumberLiteral) and lowersAndAgreesWithNodeNative(t, fx5NumberLiteral). Restore the test after proving the failure.

Searched local review and oracle files, Git history, and the integration/review-fixtures-stage and integration/review-lane-fxspptb branch trees for q12 and related literal/union evidence. Also inspected array join, indexed-read, map and for-of layout paths. No verified program with two canceling layout mistakes was found; no two-mistake claim is made.

Toolchain setup used GOPROXY='https://proxy.golang.org|direct'. The bounded setup command hit its 300-second limit while warming Go dependency actions; tools and submodules were ready. Continued with source /workspace/adamic-tools/env.sh and bounded package compilation. The first focused command's outer 90-second cold-compilation limit expired; the subsequent 300-second compilation completed. Initial fixture sources used the brief's two-argument console call and the checker rejected them with TS2554; corrected to the repository's single-argument API. nproc: 5. Timing lines are in setup.log: Go 0.079s, Node 0.094s, clang 0.527s, markdown dependencies 2.133s, submodules 36.321s.

Commands (all test output saved directly to files):

- go test ./internal/lower -run '^TestFX5' -v -count=1 -timeout 90s
- go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s
- go test ./internal/lower -run '^TestFX5NumberLiteralRefused$' -v -count=1 -timeout 90s, twice with the removed guard (refusal assertion, then native oracle)
- go test ./internal/lower -count=1 -timeout 5m -json
- go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 9m -args -update-counts (the initial 90s run timed out)
- git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -

The checkout fetch refspec initially populated only main, so explicit remote-tracking refspecs were needed for origin/devtools/fast-gate and origin/cloud/merge-tree. No other worker's branch was merged. Only object.go, the new lowering test, and review evidence were changed. No cohere code was copied.

The first full lowering run failed on missing @types/node declarations. Installed stage3/api locked dependencies with timeout 120 npm ci --ignore-scripts, then reran the full package. Initial counts refresh timed out at 90 seconds while traversing the existing corpus; reran with a bounded 9-minute test timeout. These setup failures are preserved in lower-missing-node-types.json.gz, node-types-setup.log and counts-timeout.log.

Final full internal/lower: PASS, 100.533s. New leaves: TestFX5NumberLiteralRefused 0.16s; TestFX5NumberValueRefused 0.23s; TestFX5MixedLiteralAgrees 1.76s. TestCallTargetReaders: PASS, 44.415s including its package compilation. The mixed control prints number 1 and string a on source Node and both backends.

Counts refresh: PASS, 130.373s; counts.md unchanged because the new fixtures are lowering test sources. Integration lane checks: PASS: lane checks 8.6 s: gofmt and tools on 2 Go files, t.Parallel on 1 test packages; vet 1 packages. Full lowering outputs are in lower-full.json.gz. The revert diff passed git apply --check.
