# #cyjqp5f: stopped at the lane-oracle counting gate

Base: 8e3a784aa5a601f57342c07b68f145fe56252c21. Node v24.19.0, Linux x64. Upstream v6.0.3: 050880ce59e30b356b686bd3144efe24f875ebc8. Work was in a detached scratch worktree; recursive shallow submodules initialized. No Loom submission.

The complete unfiltered counting oracle recorded six calls to getNodeSystem's public watchFile, all with interval 10. Histogram: 250=0, 500=0, 2000=0, undefined=0, 10=6; every other value=0. This fails the explicit stop gate: "If any value outside the members appears, stop and report."

The source caller is src/testRunner/unittests/sys/symlinkWatching.ts:50-66, passing 10 at :65. Its three real ts.sys cases at :696, :709 and :722 each watch both a file/directory and its symlink. They use PriorityPollingInterval and UseFsEvents, and all pass in the complete oracle. Thus an enum-only public guard would reject actual passing lane inputs. The earlier dynamic-polling counterexample remains valid, but is not representative of all watcher modes.

## Counting method and positive control

Official stage3/apply.sh completed with exit 0. Its adapted sys.ts was backed up byte for byte. In this private tree only, the exact nodeSystem object's watchFile shorthand was replaced in place with an arrow logging String(pollingInterval) plus typeof to a synchronous append-only file and forwarding the original four arguments. watchDirectory and property order are unchanged. A type assertion in this private measurement copy erases during emission; it is not a proposed adaptation. The full before/after source and unified instrumentation diff are retained.

count-run.py records the exact instrumentation and oracle command. stage3/oracle/run.py is the unmodified base script. It ran npm ci, npm run build and npm test, eight workers, all runners, no filter, light=false. Raw logs and the full report are compressed alongside the six raw interval records. The oracle completed with 106,366 passing, one failing, zero pending. The only failing title and baseline diff are the sanctioned public API row. This is the private counting oracle, not an assertion that a candidate passed stage3/lane/run.sh.

The separate Node control loads this same private tree's built/local/typescript.js. Each of 250, 500, 2000, undefined and 10 produces exactly one log entry in a separate control file, checked by assertions. This proves the site counter fires, including for undefined; the control never writes the lane file. The three enum values and 10 register and close successful priority-polling watches. Undefined throws Node ERR_INVALID_ARG_TYPE in that mode. Undefined was never observed in the complete lane. The control is supplemental evidence, not a main/change comparison witness.

## Ownership and remaining proofs

The initial grep covered stage3/adapt owners.json, targets.json and every adapt.cjs. 93-builtin-imports/proofs/targets.json and 99-wasm-crypto-import/proofs/targets.json whole-pin getNodeSystem. No owner pins or adapters were amended because the prerequisite gate failed. No candidate branch, implementation commit, Debug.fail check or ledger line was created. watchDirectory remains untouched.

Candidate Node equality/message witness, restoring mutant, full main/change census, stock tsc over 79 roots, adamic types, candidate apply, emitted JavaScript comparisons, upstream npm ci plus hereby dts declaration comparisons/empty diff, and candidate whole lane were not run. Baseline apply and the full private counting oracle are the completed proofs. No empty declaration diff is claimed.

Reproduction: start from the exact base and upstream pin, install stage3/api, export ADAMIC_TYPESCRIPT_SOURCE and scratch TMPDIR/GOTMPDIR/GOCACHE/STAGE3_CACHE/npm_config_cache, and apply into main-tree. count-run.py uses the retained scratch paths; adjust the root and remove its producer-PID wait for replay. It applies only the retained counting diff and invokes the official oracle. Run count-control.cjs separately with ADAMIC_WATCH_COUNT pointing at a fresh control file. Source and emitted bundle hashes are in sha256.json.
