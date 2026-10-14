Built complete apply/lane READS.md inventories backed by cold product and shard access traces.
Apply tip: typescript/stage3-apply-products at 46ca8671e67ed2f47fa5c0bd1b034aef4efeb14e; lane tip is this report's final containing branch tip, recorded in delivery.
All four apply products, three lane products and shard 001 pass cold below 60s; tables give measured walls.
Input-drop, package-byte, source-location, runner-marker, bytecode and shard mutants are caught; identical npm 11.9 gate installs intentionally pass.
Not covered: Planner's future read-line mutants, 150 more Go shards, other platforms, a new full-lane run or another original-producer baseline.

# Planner fixes for Task: #hefhych

The lane now consumes the tree's stage3/typescript/node_modules. It never asks
ADAMIC_TYPESCRIPT_SOURCE for packages. The TypeScript NodeProject manifests were
already on both branches and are unchanged. The package pin is generated with
registry npm 11.11.0 and the gate's ignore-scripts/hoisted policy, using the same
installed_packages and packages_digest helpers as main's API pin. The pin binds
full bytes, names, executable bits and symlink targets, including npm's hidden
lockfile. The runner marker is excluded from both digest and product copy.

Both physical installs are excluded from artifact identity and laneInputs:
their lockfiles and verified pins bind them instead. The TypeScript lockfile
must also equal the adapted predecessor's lockfile. No shard.py check was
relaxed; shard.py is byte unchanged. Its Go caller now sets a private, empty
PYTHONPYCACHEPREFIX because the empirical trace found a checkout check.pyc.

## A and B: cold products and D: one cold shard

nproc=5; cpu.max=400000 100000, a four-core CPU quota. Linux x64, Node24.19.0,
Go1.27.1, Python3.12, PATH npm11.9.0. The reference installer is npm11.11.0.
Setup exit0 and every timing line are archived in applyproducts' setup.log.gz:
17.333s total, Go build17.113s, node0.019s, Go0.022s, submodules0.060s,
markdown0.079s, clang0.156s, shared cache0.879s, cache warm17.308s.

`source /workspace/adamic-tools/env.sh` precedes Go commands. Every final cold
product runs in its own fresh go test process, -count=1 -timeout90s, with no
concurrent builds. The cache initially lacks that product; predecessors are
prepared. ADAMIC_BUILD_CACHE_DIR=/tmp/hefhych-review-cache. Apply uses the local
pinned source /tmp/hefhych-typescript. All lane runs unset
ADAMIC_TYPESCRIPT_SOURCE. Shard results use a fresh t.TempDir.

| Unit | Test wall | Fresh Go process wall | Exit |
|---|---:|---:|---:|
| Apply 10 | 25.16s | not separately recorded | 0 |
| Apply 40 | 48.54s | not separately recorded | 0 |
| Apply 70 | 50.68s | not separately recorded | 0 |
| Apply 99 | 25.53s | not separately recorded | 0 |
| Lane Install | 14.65s | 15.934752s | 0 |
| Lane Compiler | 37.05s | 38.216042s | 0 |
| Lane Harness | 26.09s | 26.802313s | 0 |
| Shard file 001 | 17.57s | 18.383023s | 0 |

Exact product commands (each row expands NN or Phase):
`timeout 90 go test ./stage3/applyproducts -run '^TestProduct_Stage3AdaptedNN$' -count=1 -timeout 90s -v`
for 10/40/70/99, each exit0. For Install/Compiler/Harness:
`go test ./stage3/lane -run '^TestProduct_Stage3LanePhase$' -count=1 -timeout 90s -v`,
each under Python subprocess.run(timeout=90), exit0. A Python monotonic clock
records complete process walls in final-BPhase.log. For the shard:
`go test ./stage3/lane -run '^TestShard_Stage3LaneFile001$' -count=1 -timeout 90s -v`,
also subprocess timeout90, exit0; counts273/0/0, no errors.

The initial attempt to use /usr/bin/time exited127 because it is not installed;
no test ran. missing-time.log retains it. Python's clock replaced it. An earlier
lane chain, before the final parent inventory addendum, passed at15.28/36.81/26.46s.
That parent document is part of the broad stage3 hash, so the old products were
removed from this task's cache and all three were measured cold again above.
The original apply overlay-disk run65.26s and unfiltered strace timeout124 are
explicit failed observations, documented and retained in the apply report;
they are not presented as passing measurements.

New apply99 and retained plain-main apply were independently whole-tree hashed:
exit0, both b122537d07680ddad789bf9e543ac807efdeb6fe6748b49d325e83764306901b.
That comparison includes every sorted non-.git path, full mode, symlink and byte.

## Read inventory: observations, reconciliation and fixes

READS.md in each package supplies paths/globs, readers and how observed. Exact
repository paths, including directories, unsuccessful probes and Git metadata,
are retained rather than filtering the trace to an expected allowlist.

| Trace | Command result | Cold traced test wall | Repository path/operation pairs |
|---|---:|---:|---:|
| Apply10 | 0 | 43.06s | 387 |
| Lane Install, final worker | 0 | 36.00s | 28199 |
| Shard001 before bytecode isolation | 0 | 25.37s | 19457 |
| Shard001 final caller | 0 | 24.43s | 19452 |

Compiled binaries avoid confusing compiler access with runtime access.
`timeout 90 go test -c -o /workspace/scratch/hefhych-review/apply.test ./stage3/applyproducts`,
`timeout 90 go test -c -o /workspace/scratch/hefhych-review/lane.test ./stage3/lane`,
and `timeout 90 go test -c -o /workspace/scratch/hefhych-review/lane-final.test ./stage3/lane`
each exit0. From stage3/applyproducts:
`timeout 90 strace -f --seccomp-bpf -yy -s 4096 -e trace=openat,execve,chdir,fchdir -o /workspace/scratch/hefhych-review/apply-fast.trace /workspace/scratch/hefhych-review/apply.test -test.run '^TestProduct_Stage3Adapted10$' -test.count=1 -test.timeout=90s -test.v`
exit0, own fresh trace cache. From stage3/lane:
`timeout 90 strace -f --seccomp-bpf -yy -s 4096 -e trace=openat,execve,chdir,fchdir,connect -o /workspace/scratch/hefhych-review/install-final.trace /workspace/scratch/hefhych-review/lane.test -test.run '^TestProduct_Stage3LaneInstall$' -test.count=1 -test.timeout=90s -test.v`
exit0, apply99 prepared and Install absent. The earlier pre-marker-fix trace
passed33.74s and is retained too. The final trace has a real gate marker present
in the input install and absent in the copied output. Traced products were
removed before the separate cold timings, not reused for them.

Final shard command from stage3/lane:
`timeout 90 strace -f --seccomp-bpf -s 4096 -e trace=openat,execve,chdir,fchdir,clone,clone3,fork,vfork,connect -o /workspace/scratch/hefhych-review/shard-final.trace /workspace/scratch/hefhych-review/lane-final.test -test.run '^TestShard_Stage3LaneFile001$' -test.count=1 -test.timeout=90s -test.v`
exit0. The earlier shard trace uses lane.test and shard.trace, with the same
options. Omitting per-open descriptor decoding reduces overhead; the extractor
tracks process cwd/fd inheritance. Both extractors exit0; final extractor reports
zero unresolved relative dirfd calls. Final trace reads no repository pyc file.

Reconciliation findings and outcomes:

| Difference | Resolution |
|---|---|
| Apply99 product_key reads top-level adapt metadata absent from Inputs.Files | Fixed recipe to include every adaptation entry at99; real metadata omission mutant caught. |
| Physical API/TypeScript installs read but absent from Inputs.Files | Explained and verified: declared lockfiles plus validated full-byte pins; runner marker not an input or output. |
| Install copier included an unkeyed runner marker | Fixed copy exclusion; removal mutant fails copied-payload assertion. |
| Shard manifests and lane/oracle runtime outside product read set | Correct additional Loom unit reads, fully listed including runtime_hash-only reads. Products do not consume those files. |
| Existing ignored check.pyc read outside source key | Go caller now uses private PYTHONPYCACHEPREFIX; forged-cache mutant proves tracked source lookup. |
| .gitattributes read by shared Git key machinery | Explicit in both inventories; current rule only affects unrelated internal/oracle fixtures. |
| Mutable .git bookkeeping, module-root probes, directory-only scans of excluded trees | Retained in exact lists; explained as key infrastructure or nonsemantic enumeration, not hidden compiler inputs. |
| Broad stage3 declarations include more than executed compiler code | artifact.input_key actually reads those bytes; remaining conservative Go/proof declarations are intentional. |

`timeout 5 unshare -n true` exit1, Operation not permitted. Network namespaces
cannot be disabled here. Install and both shard traces show zero connect calls;
products invoke no git fetch, npm ci or package download. Source cloning in apply
is local file://. Compiler/harness npm commands run offline with empty private
user/global configuration and a product-local cache, not HOME. No product does
network provisioning. Registry bootstrap/install below happens before products.

## npm pin provenance and version mutant

`timeout 90 curl -fsSL https://registry.npmjs.org/npm/-/npm-11.11.0.tgz -o /workspace/scratch/hefhych-review/npm-11.11.0.tgz`
exit0; tar extraction exit0. Archive sha256:
cbcf4cc03148ccdb586a8bf2093c952f093fb43d5cbc97593c98b67ef8c003b0.
`node /workspace/scratch/hefhych-review/package/bin/npm-cli.js --version` exit0,
11.11.0. `timeout 90 node /workspace/scratch/hefhych-review/package/bin/npm-cli.js ci --prefix stage3/typescript --ignore-scripts --no-audit --no-fund --cache /tmp/hefhych-upstream-npm`
exit0,338 packages in3s. npm's default hoisted strategy matches the gate's explicit
--install-strategy=hoisted. Tree lockfile sha256:
e85be648176ba25eaa38f9c1b7b7ca9bab782b0c448fb15fc44c355c40ef9a4b.
Full package digest and committed lane pin both equal:
830a100752031492fa8283014487f1b075ecadf03f4e5fb1e31cc20a7e5e3d9d.
The install's hidden lock entries and package versions also match the lockfile.

`timeout 90 npm ci --prefix /workspace/scratch/hefhych-review/npm119-project --ignore-scripts --no-audit --no-fund --install-strategy=hoisted --offline --cache /tmp/hefhych-upstream-npm`
with PATH npm11.9.0, same package/lock files, exit0. packages() accepts it: its
full digest is byte-identical, including .package-lock.json. This is the requested
npm-version mutant's deliberate pass, justified by exact bytes, not version-only
trust. The earlier npm11.9 install used different installed bytes and is refused
with the actual digest87a89f... and a message naming stage3/lane/node_modules.sha256
and its npm11.11 reference. For example, esbuild/bin/esbuild is9350 bytes in the
gate-policy install versus11108536 bytes in the prior install. Version metadata
alone would not detect that difference; the byte pin does.

## C: artifact size and reuse

| Measurement | Exact observed value |
|---|---:|
| ready.json + shards-ready.json + tree/** + api/** logical bytes | 720961310 |
| Regular files in those four entries | 92205 |
| artifact.tgz, the buildcache-stored archive blob | 182089130 bytes |
| Whole product including archive and phase/config metadata | 903050533 bytes, 92208 files |
| Second final-product fetch | 0.11s product lookup; 0.49s top-level test; 1.270720416s fresh Go process |

All payload entries are regular files; zero symlinks. Local buildcache retains
both materialized files and the explicit archive blob; no remote-store upload
is claimed. Size walker and fetch command both exit0. Reuse command:
`go test ./stage3/lane -run '^TestProduct_Stage3LaneHarness$' -count=1 -timeout 90s -v`
under subprocess timeout90, predecessor/compiler hit0.25s and Harness hit0.11s.

Final buildcache Harness directory key: a2f57b0aabc03aee37827cfddc45c7560d54e3ded5b51e01429a339a56eb158b.
Artifact input identity: ea363c16f8a4a7ba016a77a62522ec2aef15569a93fe0f3df8f43c31f90cf04e.
Runtime identity: c11db10d59521181432ea6bc97ce72323c8bd7975330cb8e79c596d6740d29c5.
`timeout 90 python3 -B stage3/lane/bind_products.py ARTIFACT` exit0,
151 existing units rebound; no task/count repartition or additional Go shards.

## Mutants and checks

| Mutation | Executed check | Observed rejection |
|---|---|---|
| Drop source.json independently for apply10/40/70/99 | prove_inputs.py invokes each TracksInputs test with -count1 -timeout90 | Four exit1, changed source pin reused product key |
| Drop top-level metadata from apply99 | TestStage3Adapted99TracksMetadata, real recipe/private file changed | Exit1, same key-change assertion |
| Drop products.py independently for lane Install/Compiler/Harness | prove_products.py invokes each TracksInputs test | Three exit1, changed lane worker reused product key |
| Omit upstream pin field from artifact identity | test_upstream_installed_bytes_invalidate_key | Exit1 AssertionError |
| Reinclude physical API install | test_installed_parser_is_keyed_by_its_byte_pin | Exit1 AssertionError |
| Disable installed-byte pin comparison | test_changed_installed_bytes_are_refused_before_copy | Exit1 ValueError not raised |
| Reinclude physical tree TypeScript install | test_tree_install_is_keyed_by_pin_and_lockfile | Exit1 AssertionError |
| Restore external checkout package sourcing | test_changed_installed_bytes_are_refused_before_copy | Exit1 external checkout must not supply packages |
| Remove root runner-marker copy exclusion | test_runner_marker_is_not_published | Exit1 copied marker exists |
| Remove private bytecode prefix with a valid forged checkout check.pyc present | Actual TestShard_Stage3LaneFile001 | Exit1 hefhych checkout bytecode mutant; positive ignored forged cache and passed |
| Break owned projectReferencesSourcemap test in private artifact copy | prove_shard.py builds private harness and runs actual manifest argv | 001 exit1 with planted error;002 exit0,6/0/0; product hash unchanged |
| Delete shards.json | Actual Go shard | Exit1 no such file |
| Empty manifest units | Actual Go shard | Exit1 missing version1 units |
| Exit0 but report272 passes | Actual Go shard | Exit1 differs from manifest |
| Correct counts but command exit1 | Actual Go shard | Exit1 manifest command |

Positive/key commands and coordinator exits:
`timeout 90 go test ./stage3/applyproducts -run 'Tracks|Deadline' -count=1 -timeout 90s` exit0;
`timeout 150 python3 -B stage3/applyproducts/prove_inputs.py` exit0;
`timeout 90 go test ./stage3/lane -run 'TracksInputs|Rejects' -count=1 -timeout 90s` exit0;
`timeout 150 python3 -B stage3/lane/prove_products.py` exit0.
Additional fixture/recipe mutants run under timeout90 with restoration in finally;
exact scripts and logs are archived. `timeout 90 python3 -B prove_bytecode.py`
exit0; positive fresh Go process18.843733s, negative1.318060s. The forged cache
and Go source are restored. `timeout 180 python3 -B stage3/lane/prove_shard.py ARTIFACT`
exit0 on the final artifact; every child has timeout90. Complete argv, exit and
wall are in D-mutants-final.log.gz. The private-source mutant results preserve
the planted error and unaffected shard002; no product was modified.

Final package/readers validation:
`timeout 90 go test ./stage3/applyproducts ./stage3/applyproducts/fetch ./stage3/lane -count=1 -timeout 90s`
exit0, apply1.918s and lane18.610s. Import search finds only fetch and lane as
Go consumers of applyproducts and no additional Go consumer of lane.
`timeout 90 python3 -B -m unittest discover -s stage3/lane -p 'test_*.py'`
exit0,73 fixtures in11.079s; `timeout 90 python3 -B stage3/test_apply.py` exit0,
25 fixtures. `timeout 90 go vet ./stage3/applyproducts ./stage3/applyproducts/fetch ./stage3/lane`
exit0. All logs are archived without piping test output.

The final apply tip's prescribed lane checks exit0,2.6s, three Go files and one
test package. Command on each committed branch before push:
`timeout 90 bash -c 'git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -'`
after sourcing the toolchain environment. Final lane result and actual commit
and push outputs are recorded in delivery after the commit is created. The
branch carries the apply fix by normal merges, with no history rewriting.
Every new commit ends Task: #hefhych; all net file changes are under stage3.

No new original-producer timing is claimed: wave1's143.365293s reference remains
historical. No new full151-shard union or full lane run, other-platform proof,
remote-store measurement, Loom submission or Planner-generated read-line mutant
is claimed. Only shard001 is declared in Go. No cloud/fast-gate file changed.
