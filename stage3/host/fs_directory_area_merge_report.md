Built: merged the library area into the directory/path landing branch, preserving both units' intent and the stronger QualifiedName proof checks.
Commits: merge parents d447586ce3c121b748e270bb8dcd28b12f9ad9c1 and 2faf682bbf365dcc0afde3768832a2fdffc7b965; the pushed merge SHA is reported separately.
Checks: touched packages, complete Linux counts regeneration, nine input fixtures on both backends, directory/path mutants, flow and freshness passed; detailed retries below.
Mutants: all 21 directory/path compiled mutants and five owned upstream source mutants were caught only by Node stdout comparison; the fs-file close failure mutant also passed after its fixture adaptation.
Uncovered: owned upstream 07/08/09/24 remain NotYet on both backends; 25 remains Checker on both backends. No full repository gate or macOS execution.

The latest landing instruction is acknowledged: merge origin/area/library into codex/host-fs-directory-land with a merge commit. No rebase, force-push, main push or PR.

## Conflict resolution

stage3/fixtures/NOTICE and stage3/fixtures/host/status.json exactly match area/library. The area QualifiedName positive and negative proof tests are retained. The counts registry retains both the area's slow regex fixtures and this branch's fs-file fixture preparation and counts.

Two fs-file test sources used an assertion from Error to NodeJS.ErrnoException that the area's proof now correctly refuses. They now assign the Error to a typed NodeJS.ErrnoException variable, preserving the same code property observations. The lowerer refusal test remains intact. This changes no runtime implementation. The complete Linux counts regeneration measured the resulting close fixture reference-count change and the System directory fixture change on the merged tree.

## Validation

All commands use `source /workspace/adamic-tools/env.sh`. Test output was written directly to log files. Logs, including initial failures and retries, are in [logs/fs_directory_area_merge](logs/fs_directory_area_merge/).

- `bash cloud/setup.sh`: PASS; Go 0s, clang 1s, Node 1s, submodules 1s, cache warm 303s, total 303s; nproc 5, quota 4 CPUs. Node v24.19.0, Go 1.27.1, clang 20.1.8.
- `npm ci --prefix stage3/api`: PASS, three packages in four seconds; official pinned @types/node 25.3.3.
- `go test ./internal/lower -run 'TestNode(Library|FSFileQualified)' -count=1`: PASS 53.594s after the qualified assignment adaptation.
- `go test ./internal/load ./internal/lower ./internal/native ./internal/javascript ./internal/flow ./internal/fresh ./internal/regexp ./internal/unicodeproperties -count=1 -timeout 30m`: load PASS 18.260s, lower PASS 198.661s, native PASS 686.650s, JavaScript no tests, fresh PASS 269.483s, regexp PASS 25.316s. Flow initially rejected the old close fixture assertion; Unicode's Node batch timed out under contention. All other Unicode comparisons had zero disagreements.
- `go test ./internal/flow ./internal/fresh -count=1 -timeout 30m -parallel 3`: PASS flow 198.032s, fresh 103.186s with the corrected close fixture.
- `go test ./internal/unicodeproperties -run '^TestCanonicalizeUnicodeNode$' -count=1 -timeout 30m -args -unicode-node-workers=3`: PASS 312.988s, isolated from other CPU-heavy tests.
- `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -parallel 3 -args -update-counts`: complete Linux regeneration PASS 60.963s. The initial attempt failed on the old close assertion and a contended long-regex timeout; it did not update the ledger.
- `go test ./internal/oracle -run 'TestInputAgreesWithNode/internal/oracle/testdata/(node_|realpath)' -count=1 -timeout 30m -v`: PASS 15.859s, nine input fixtures, including all seven directory/path inputs, original realpath and Buffer input; both backends, sanitizer and leak checks.
- `go test ./internal/oracle -run 'Test(NodeFSDirectory|NodePathRelative|NodeFSFile|NodeBuffer|RealPath|TypeOf|MapSmall|RegExpLongBacktrack)' -count=1 -timeout 30m -v`: all directory/path runtime checks and 21 mutants, map-small and typeof mutants, and realpath checks passed. Initial close assertion and long regex sanitizer failures are retained in the log.
- `go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(node_buffer_|regexp_native_|library_map_small_|library_map_set_keys|proven_|typeof_|inherited_static_field_read)' -count=1 -timeout 30m -v`: area fixtures and Buffer passed except a contended quadratic_matchall sanitizer timeout. Release output agreed with Node. The first isolated retry reused the cached timeout; fresh uncached retry evidence follows below.
- `go test ./internal/oracle -run 'Test(RegExpLongBacktrackNode|RegExpNativeTiming|NodeFSFile(AgreesWithNode|Mutants))/(close|closeSync)?' -count=1 -timeout 30m -parallel 1 -v`: PASS 259.158s. All fs-file comparisons and mutants ran, including close and closeSync_swallowed_failure. Long regex sanitizer, release and leak checks PASS 202.96s; all four native timing probes PASS 18.68s. Original deadlines are unchanged.
- `gofmt -l cmd internal` and `go vet ./...`: PASS, empty logs.

The earlier close-only retry matched no tests; its retained log is not treated as validation. Timeout evidence is retained and no test deadline was weakened.

## Owned acceptance fixtures

`python3 stage3/host/check_fs_directory.py --mutants --logs /tmp/fs-directory-area-acceptance` exits 1 because the compiler stages remain blocked. All five source Node runs agree with the area's audited status.json, and all five source mutants fail only stdout comparison.

| Fixture | Native | JavaScript | Node vs area status.json |
|---|---|---|---|
| 07_directoryExists | NotYet: node:fs.mkdtempSync | NotYet: node:fs.mkdtempSync | Agrees |
| 08_getDirectories | NotYet: node:fs.mkdtempSync | NotYet: node:fs.mkdtempSync | Agrees |
| 09_realpath | NotYet: node:fs.mkdtempSync | NotYet: node:fs.mkdtempSync | Agrees |
| 24_useCaseSensitiveFileNames | NotYet: node:fs.mkdtempSync | NotYet: node:fs.mkdtempSync | Agrees |
| 25_readDirectory | Checker | Checker | Agrees |

A one-line reproducer for the first blocker is `adamic build stage3/fixtures/host/07_directoryExists.a`; the lowerer reports it cannot lower node:fs.mkdtempSync. The readDirectory fixture has checker errors TS2345, TS2532, TS2322, TS2775 and TS7030. Full diagnostics and observations are retained in the acceptance evidence. No driver APIs outside this unit's census were added and no acceptance sources or audited status were rewritten.

Fresh retry: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/regexp_native_quadratic_matchall[.]a$' -count=1 -timeout 15m -parallel 1 -v`: PASS 16.982s, no cache hits, source Node and both backends agree, sanitizer and leak checks pass.

Whitespace verification is restricted to edited source and this report: the complete staged merge includes inherited stock TypeScript CRLF diagnostic baselines and literal diagnostic logs with trailing whitespace. Those oracle observations are preserved byte-for-byte.
