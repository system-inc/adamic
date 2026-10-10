Eight deterministic, sorted-name, round-robin shards cover all 933 registered oracle fixtures. Each gate selects one named subtest; the original test name selects the whole grid. The fixed count is eight because all measured isolated runtime-cold commands, including setup, finished below the 45-second target. Fixtures and compiler sources are unchanged. ADAMIC_GATE_UNCACHED is still handled by the existing result-cache helper; runtime archives retain their existing cache policy.

Measurement machine: Linux amd64, cgroup cpu.max `400000 100000` (four CPU quota), GOMAXPROCS=4, -parallel=4; Go 1.27.1, Node 24.19.0, native clang 20.1.8, WASI SDK 29 / clang 21.1.4. Base: `7a10c877667582a15326acb10fa22fa7a0c45fb8` from origin/main. Protocol reference: stage1-split/markdownblocks `ad97e448`, events_shards_test.go.

Each command gets a separate initially empty XDG_CACHE_HOME: both oracle observations and all native/WASI runtime archives start cold. Go dependencies use the shared GOCACHE populated by the bounded startup attempts. Every command uses -count=1 and ADAMIC_GATE_UNCACHED=1; the first shard includes compilation of the changed Go test package. Wall means the entire go command, not the Go JSON parent-test Elapsed field (which excludes parallel children). SDK installation, submodule acquisition, and npm ci are environment provisioning, performed before the final measurements.

| Gate subtest | Fixtures | Command wall | Native identity setup | Result |
| --- | ---: | ---: | ---: | --- |
| shard-000 | 117 | 40.226 s | 11.369s | pass; no kill |
| shard-001 | 117 | 33.826 s | 11.463s | pass; no kill |
| shard-002 | 117 | 33.956 s | 11.077s | pass; no kill |
| shard-003 | 117 | 35.491 s | 10.667s | pass; no kill |
| shard-004 | 117 | 34.651 s | 11.400s | pass; no kill |
| shard-005 | 116 | 35.692 s | 10.940s | pass; no kill |
| shard-006 | 116 | 34.938 s | 11.593s | pass; no kill |
| shard-007 | 116 | 34.342 s | 11.395s | pass; no kill |

Before (exact origin/main test, all required dependencies installed, runtime-cold): killed at 90.010 s by the hard process-tree deadline; this is a failure under the replacement rule. See before-installed.json.gz and before-installed.time.json. It was not allowed to finish for a final duration. After: slowest shard-000, 40.226 s. The measured Node identity / three native-runtime archive setup is 10.667–11.593 s per isolated shard. WASI feature-specific archives are built lazily by the existing native.Build path and included in the command walls. The original single-fixture cold-runtime probe took 17.673 s including its WASI archive. Setup is not hidden or shared across these separate shard commands.

The fully cold Go dependency build is still over budget: the initial empty GOCACHE invocation was killed at 75.008 s before any test started. A subsequent single-fixture startup attempt was also killed at 75.003 s after remaining compilation and beginning setup. This change does not solve a gate that recompiles all Go dependencies from an empty GOCACHE per unit; that stricter requirement remains an honest partial. No compiler changes were made to address it.

The user changed deadlines during measurement. Historical baseline/startup runs and shards 000–004 used the then-current 75-second kill rule; shards 005–007 and the proofs used the replacement 90-second rule. None of the final shards hit either deadline. All future gate invocations should use -timeout 90s plus a 90-second process-tree deadline if setup compilation also needs bounding; reaching that deadline is a failure. The budget remains below 60 s, with 45 s as this split's target.

Initial pilot shards 000 and 001 failed because the environment lacked the pinned @types/node 25.3.3. After npm ci --ignore-scripts in stage3/api, both were rerun with new empty caches and passed. Their pilot evidence is retained separately; failed observations are not counted as completed fixtures.

The union test checks every registered row and fixture name exactly once, and rejects empty/missing shards, omitted fixtures, and duplicates. The planted-disagreement test uses the same comparison as the live shard runner. The real planted-fixture subprocess changes one lowered dedication string in memory, then requires an actual test failure naming shard-001/dedication/dedication.a and stdout differs. It does not edit a fixture. The proof command passed in 22.543 s, including cold runtime setup.

The receipt verifier requires a successful package, successful selected shard, and one run/pass pair per registered fixture. All eight actual logs prove 933 successful fixture completions exactly once. Omitting each shard in turn is rejected as missing (completion-proof.txt); skipped, duplicate, and killed/truncated receipts are rejected too. A declared union alone is never treated as completed execution.

Run one gate unit from the repository root (with WASI_SYSROOT and the provisioned tools on PATH):

```sh
ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 GOMAXPROCS=4   go test ./internal/oracle -run '^TestWASIAgreesWithNode$/^shard-000$'   -parallel 4 -count=1 -timeout 90s -json > shard-000.json
```

Verify the saved, real completed union from the repository root:

```sh
python3 review/split-oracle-wasi/verify_wasi_shards.py   review/split-oracle-wasi/proof.json.gz   review/split-oracle-wasi/shard-*.json.gz
python3 review/split-oracle-wasi/verify_wasi_shards.py --self-test
```

Raw Go JSON receipts are compressed without filtering. Timing JSON includes the exact command, exit status, wall, and deadline status. The saved measure.py and prove-completion.py reproduce the commands on this machine; their paths describe this environment. The baseline overlay executes origin/main's original wasi_test.go and removes only the newly added shard-proof test file; it does not alter compiler or fixture sources.
