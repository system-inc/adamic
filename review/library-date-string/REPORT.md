Built shared Date string conversion for String(d), templates, string addition and d.toString, including Date-or-undefined and Date-or-null.
Branch codex/library-date-string starts at f133f7c7631ed4da3b7bb24ee865ce1bfa1e3cca; compiler and combined-proof merges are scratch-only.
Checks: touched packages and full flow PASS; owned uncached Node oracles, UTC WASI, Linux counts, vet and formatting PASS; host fixture 10 agrees on both backends.
Mutants: flipped GMT sign, weekday shifted by one, and Invalid Date capitalization each compiled and finished leak-clean; only Node stdout caught them.
Remaining: named refusal outside admitted zones/ranges, Denver refused on WASI; no macOS execution or full repository oracle run. No main or force push.

The Date string conversion

One named library conversion consumes each operand once and uses the existing Date internal slot supplied by the fs file worker. String conversion and templates pass through that function; string addition's helper evaluates both operands before ToPrimitive, including the right operand when the Date is on the left. The selected conversion uses Date's string/default primitive hint. General user-object conversions, Date subclasses and prototype mutation are not implemented here.

Reused from codex/library-date/date.c: Gregorian floor division and civil calendar splitting, English weekday/month labels, and V8's signed/padded year format. Reused from the fs landing on the requested base: numeric Date construction, TimeClip and getTime/valueOf storage. No broad Date branch was merged and none of library-date-3's compile-time TZ checks was taken. Numeric Date inputs and the existing fs-created dates are this unit's scope; new Date() clock access and string/calendar construction remain outside it.

Runtime timezone contract and assumption

Assumption recorded in the commit: support the exact zone IDs UTC and Etc/UTC throughout the Date range, and America/Denver for Unix seconds 0 through 2147483647 inclusive. Other zones and wider Denver dates throw a catchable Error named DateStringNotYet. UTC's long name is Coordinated Universal Time; Denver uses Mountain Standard Time or Mountain Daylight Time. These exact English names are held to Node 24.19.0 rather than inferred from libc abbreviations. The runtime calls tzset/localtime_r against the system zone database and verifies the resulting Denver calendar against its supported -0700/-0600 offsets before emitting either name. No timezone is read by lowering or embedded from the compiler environment.

When TZ is unset, the native and JavaScript sides resolve a canonical /etc/localtime symlink under /zoneinfo/; an unidentifiable system zone refuses. Copies rather than identifiable symlinks also refuse. Every unsupported zone must take the named runtime path rather than guess an ICU long name. NaN dates always render Invalid Date even in an unsupported zone. The JavaScript backend enforces the same named zone/range boundary, then uses V8's intrinsic Date.prototype.toString. Native and JavaScript are tested with the same artifact executed under UTC and America/Denver. The two-zone test changes only each child's environment, never the compiler/test parent's environment.

The oracle covers all three conversion forms and nullable cases, d.toString, clipped invalid dates, argument evaluation counts, 500 deterministic timestamps, both Denver DST transitions and winter/summer names. UTC also covers negative years, year zero, negative milliseconds and both TimeClip endpoints. A separate Paris probe proves both backends throw the named refusal, catch it, preserve right-operand evaluation order and subsequently format an invalid Date; stock Node successfully formats Paris, so this is an explicit supported-surface boundary, not a claimed Node agreement for Paris.

WASI

All new POSIX calls (readlink, tzset and localtime_r) are behind ADAMIC_TARGET_WASI. UTC calendar formatting needs none of them. ADAMIC_ORACLE_WASI=1 built the entire runtime archive and passed both Date fixtures against Node, emission checks and the WASI runner mutants (11.101s). TestDateStringWASIRefusal proves the Denver conversion throws DateStringNotYet with the precise reason that the WASI system zone database is unavailable (0.950s). WASI uses explicit runtime TZ=UTC; an unidentifiable default zone refuses. Darwin uses _DARWIN_C_SOURCE after _POSIX_C_SOURCE, and LeakSanitizer is selected only on Linux. macOS was not executed.

Compiler proof and host fixture 10

Scratch /workspace/scratch/date-string-proof includes combined proof 2011c9193123ffc80cb23901ac6b5cc32950d422, compiler 2606e8b444f57c59dfa38b9e483d33f94a0cbd2f, and requested base f133f7c. Scratch merge commits are 554e16fc (compiler), f57d0854 (library proof plus closure ABI reconciliation) and 16709de8 (base). Compiler merge conflicts preserve the host namespace/buffer hooks and readiness checks alongside the compiler's callable and optional-slot work. The scratch process runtime closures were adapted to the compiler's argument-count ABI; that is a reproduction dependency and is not part of this Date landing. The base's audited status.json wins its conflict. GOFLAGS=-buildvcs=false accommodates the scratch dependency symlinks.

Before the Date patch, stock 10_getModifiedTime.a stopped at line 32:24 at String conversion of an object. With the patch, native and JavaScript builds exit zero, and both executions agree byte for byte with Node and the audited status.json: 946684800000, undefined, 946684800000 on successive lines. Each exits zero and prints no stderr. No fixture source or recording was edited on the landing branch. See host10.json and preserved command logs.

Commands and observations

After export GOPROXY='https://proxy.golang.org|direct' and source /workspace/adamic-tools/env.sh:

- bash cloud/setup.sh: first overlapped the creation of the new helper and failed during cache warming with undefined helper methods. Reran successfully after the helper was complete: Go and Node ready 0.095s, submodules 0.207s, markdown 0.238s, clang 0.407s, build 62.173s, cache warm 62.459s, total 62.548s; nproc=5, cgroup quota=4. Go 1.27.1, clang 20.1.8, Node 24.19.0.
- TZ=UTC go test ./internal/lower ./internal/javascript ./internal/native ./internal/flow ./internal/load -count=1 -timeout 30m: PASS lower 59.698s, native 264.176s, flow 185.697s, load 3.825s; JavaScript package has no test files and is exercised by both-backend oracles.
- TZ=UTC ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestDateString|TestNativeAgreesWithNode/internal/oracle/testdata/library_date_string' -count=1 -v -timeout 30m: final PASS 1.137s, zero result-cache hits. Native sanitized/release and JavaScript comparisons pass; UTC and Denver runtime tests pass.
- TZ=America/Denver ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestDateString|TestNativeAgreesWithNode/internal/oracle/testdata/library_date_string.a$' -count=1 -v -timeout 30m: PASS 1.730s, zero result-cache hits.
- TZ=UTC go test ./internal/oracle -run TestDateStringRuntimeMutants -count=1 -v -timeout 30m: PASS 23.811s. Three runtime copies mutate real behavior; each uses ASan/UBSan and detect_leaks=1 on Linux, exits zero with no stderr, and disagrees with source Node only in stdout. Production files are never mutated.
- TZ=UTC go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts: PASS 70.283s; verification without update PASS 35.141s. Two rows added; directory enumeration counts move with the three new fixture filenames.
- TZ=UTC WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run 'TestWASIAgreesWithNode/internal/oracle/testdata/library_date_string|TestWASIEmission/internal/oracle/testdata/library_date_string|TestWASIRunnerCatchesMutants$' -count=1 -v -timeout 30m: PASS 11.101s. The same environment with -run TestDateStringWASIRefusal passes 0.950s.
- go vet ./internal/lower ./internal/javascript ./internal/native ./internal/oracle and gofmt -l cmd internal: PASS, empty logs. git diff --check: empty.

The full repository/oracle gate was not run; the full touched package gate, full flow, owned differential fixtures, Linux counts and WASI legs are the worker gate. Evidence is stored beside this report.
