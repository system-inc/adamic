The brief says six rows but lists 94 functions. All 90 decoder wrappers call the same checker with only target and index inputs. The family rule therefore produces five rows. Treating targets as separate rows would split that shared checker. members.json records every member.

The brief cites 8de93800f4; fetched origin/main was b902a0cc. All names were checked through go test -list and remain in the cited files. All diff locations refer to fetched main.

The whole-package clean baseline timed out at 90.024s without an earlier reported failure. The scoped clean baseline passed in 76.942s with WASI enabled. The combined inactive switch baseline also timed out at 90.017s without an earlier reported failure. Splitting decoder execution by target gave clean inactive runs of 58.791s native and 18.792s WASI. Matrix parts observe every member and count as one family. Tests outside the recorded bounded sets remain unknown. No other package was tested.

Warm env.sh lacked WASI SDK. SDK 27 was installed and ADAMIC_TEST_WASI=1 enabled for scoped baseline, costs and matrices. The initial whole-package baseline skipped WASI. No scoped row skipped after installation. Core setup was skipped, zero seconds. SDK installation and npm ci were not timed separately.

ADAMIC_BUILD_CACHE_DIR controls the decoder cache. library.go's runtimeLibrary instead uses a content-hashed user cache. Each selector run has its own decoder cache directory, seeded with hard links to immutable products from the identical switched source. The C selector is read per process. This reuses a valid switched product, not a stale product from before mutation. The emission rows inspect generated text and do not compile it into native products.

The first cost run for each row included coverage instrumentation; all three used -count=1 and the binary's reported package seconds. This limits comparisons to purely uninstrumented timings. The decode family median measures all 90 members, not the near-zero scheduling time of a parallel wrapper.

The Go reach inventory contains 142 covered function-location records. The C inventory names decoder and formatting entries and routines, but transitive allocation, output, startup and string helpers were not exhaustively instrumented. Those paths were not targeted. Twelve code-derived mutants, four per area, fall below the aspirational three per grouped row. The budget favored observing all family members and validating each standalone diff.

No test or oracle was changed. The receiver assignment row passes its own empty-answer probe because zero is its expected negative result. This is recorded as vacuous, separately from its observed production kill. Subsumption is a hint from the recorded mutant set, not grounds for deletion. Expected adapter computation in the devirtualization row shares production exactReceiverMethod, limiting oracle independence.

No package-wide or repository-wide uniqueness is claimed. Central replay must settle that. No PR was opened and main was not pushed. Baseline, timing, matrix, probe, compile and replay-check logs are retained with the standalone diffs.
