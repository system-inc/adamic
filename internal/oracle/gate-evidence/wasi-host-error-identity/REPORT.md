# Nominal identity for host errors

Base: codex/host-proof-combined ead7c3727983b134b077e8e723b07bb058489f0c.
Branch: library/wasi-host-error-identity.

The host constructors allocated plain adamic_object values with class == NULL.
The generated instanceof Error check uses nominal class identity, so name,
message and code alone do not satisfy it. This affects native Linux as well as
WASI; the combined report itself records the Linux reproducer returning false.
The final six-case source fixture was also run on the unchanged baseline:
WASI compiled and exited zero, but printed six "not Error"/"other" pairs,
while Node printed the nominal kinds and string code/message. Its Node
comparison failed as expected (identity-baseline.log.gz). No WASI refusal path
is converted to a catchable error.

adamic_error_tag attaches the reserved built-in definition identity used by
lower/error_classes.go. Its descriptor retains the existing three-slot owning
layout, own string message/code and destructor traversal. TypeError and
RangeError keep their identities and Error ancestry. File and directory errors,
process directory/code errors and the legacy runtime Error constructor use this
shared path. The latter also repairs the performance host's guarded SyntaxError
output. No name-based instanceof fallback or lowering relaxation is introduced.

The new .a witness triggers empty-path readFileSync, readdirSync and statSync, NUL-path
TypeErrors on file/directory paths, and closeSync RangeError.
Node, the JavaScript backend, sanitized native Linux and WASI agree on Error
identity and the full string code/message. The native leak check passes. A real
WASI mutant removes only each pending object's class tag after the host call;
it compiles and exits zero with no stderr, printing six "not Error"/"other" pairs.
Only the Node stdout comparison catches it. The witness is registered with the
ordinary oracle through one line in internal/oracle/oracle_test.go.

The two optional-widening refusals are independent of runtime identity. Their
internal oracle witnesses now use the existing checked unknown-to-string code
helper instead of assigning Error to NodeJS.ErrnoException. The compiler's
no-optional-widening refusal remains intact. Both formerly refused witnesses
pass the native and WASI input comparisons and the existing symlink/stat mutants.

Verified commands write directly to logs:

```sh
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestWASIHostErrorIdentity$' -count=1 -v -timeout 30m
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^(TestWASIHostErrorIdentity|TestWASIFileAgreesWithNode|TestWASIInputAgreesWithNode|TestInputAgreesWithNode|TestNodeFSFileAgreesWithNode|TestNodeFSDirectory.*)$' -count=1 -v -timeout 30m
go test ./internal/native -count=1 -timeout 30m
go vet ./internal/native ./internal/oracle
gofmt -l internal/oracle/wasi_host_error_test.go
git diff --check
```

Focused identity/mutant proof passed in 12.805s. The native plus WASI host
selection passed in 125.673s, including source Node, JavaScript backend,
sanitizers/leaks and the existing directory mutants. Existing explicit WASI
platform, argv, mode, timestamp, temporary-directory and special-kind refusals
remain covered rather than converted into agreement.

The full native package completed in 973.671s and failed outside the host
contract: stale two-argument closure/method C harnesses against the three-argument
ABI, missing pre-existing runtime-static inventory entries, and the graph
container boundary's map use-after-free. Representative closure/method failures
were reproduced on ead7c37 (13.615s), as were the static inventory and graph UAF
(11.057s). These failures were not repaired in this host identity unit. Vet,
formatting and diff checks passed.

Whole TestCountsAreRecorded -update-counts was attempted. It fails in unrelated
compiler/graph witnesses (taste/21_truthy_loops, graph_regions_cache, concurrency
closure ABI and virtual target inference). The same taste, concurrency ABI, graph frees and virtual-call panic were
reproduced on an isolated unchanged ead7c37 worktree. The failed run does not
write the table. All 41 host rows were independently measured with the existing counted
runner (94.578s); only those 40 existing rows and the new fixture row were
updated. The saved measuring runner is counts-runner.go.txt; its temporary test
was removed. The new row is 26 allocations/frees, 36 retains, 52 releases,
peak 5, no arena/graph regions or merges. Whole-table freshness is not claimed.

Setup used GOPROXY='https://proxy.golang.org|direct' and --wasi-sdk; sourced
/workspace/adamic-tools/env.sh. Node v24.19.0, Go 1.27.1, clang 20.1.8,
WASI SDK 27; nproc=5, quota=4. Timings: Node 0.033s, Go 0.032s,
submodules 0.082s, markdown 0.096s, clang 0.291s, WASI SDK 0.418s,
Go build 91.021s, cache warm 91.715s, done 91.756s.

Limits: no whole repository, whole oracle, whole flow or macOS gate is claimed.
