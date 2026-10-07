# Nexus no utils folder port

Current status: filename transport is fixed and the full unified checks pass.
See LANDING.md and evidence/filename-fix. The following record preserves the
earlier blocked run and its reproducer.

Built the unified SourceFile listener, node:true descriptor, exact messages,
JSON-preserving Go adapter and a clean-running mutant. Source:
origin/codex/stage1-lint-batch4:nexus_consistency_no_utils_folder.ts. Base:
origin/area/stage1-lint d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898.

Upstream test prefix TestConsistencyNoUtilsFolder captures both actual tests:
TestConsistencyNoUtilsFolderFires (three paths) and
TestConsistencyNoUtilsFolderStaysSilent (five paths). The rule accepts any
and ignores all options; no upstream options struct exists. The adapter
decodes field 5 as JSON and passes the decoded value without dropping data.
No option is invented to simulate a filename.

Commands and observations:

- go run ./cmd/lint-registry: exit 0.
- gofmt -w rules/nexus-consistency-no-utils-folder/oracle.go: complete.
- go vet ./...: exit 0, empty output.
- go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v
  -timeout 20m: FAIL, witness reports no findings, 6.522s.
- go test ./stage1/cohere/lint -run
  '^TestMutants$/^no-utils-folder-utils-segment-omitted$' -count=1 -v
  -timeout 20m: FAIL, mutant survived on Node, 19.730s. The complete
  unrelated registry mutant suite was not run after this blocker.
- go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -v
  -timeout 20m: PASS, 13,053,823 bytes, 97.732s. This does not certify
  this path-only rule, because the captured paths have lost their directories.

Blocker: ownedWitnesses generates a new temporary basename and upstream
capture writes case-NNN/<basename>. These drop utils and _utils directory
segments. The witness-options branch adds only JSON options, which this
upstream rule ignores. The smallest path reproducer and required shared
filename transport are documented in BLOCKER_HISTORY.md. No shared harness changed.

Independent real-path proof uses the same unified driver and unchanged Go
registry oracle, with 21 selected/all/options rows and 12 findings. Go, source
Node, emitted JavaScript and ASan/UBSan native match on 3,167 bytes. The
utils-segment-omitted mutant compiles, exits 0 with empty stderr on each side,
and only the Go byte comparison catches it (2,078 bytes instead). Sanitizers
pass on both normal and mutant runs. The initial manual emitted-JavaScript
command omitted the Adamic runtime loader and failed with package-not-found;
its log is retained. The corrected runner uses oracle/node.mjs, as the shared
harness does. Exact script, manifest and stdout/stderr are in evidence/.

This is a blocked port, not landing-ready. Full repository gate and the full
registry mutant suite remain unrun. Existing TestRulesAgree explicitly
records unsupported parser-recovery cases; those were not relaxed or hidden.
No main or area ref was pushed. Stop here pending filename transport support.
