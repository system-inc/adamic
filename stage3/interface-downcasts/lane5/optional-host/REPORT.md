Built: optional host method view admission with the original realpath bind and directory condition contexts.
Commit: 61d77dbeefd069c1865c48a8122be0525444c4f1; base 926a1d39d1a0d6b4cf49bbccf521a4cc02d10f56; branch codex/views-optional-host.
Commands: scoped lower and oracle tests PASS; lane counts update and verification PASS; global counts has the same 39 failures as the unchanged base.
Mutants: native-presence and javascript-presence caught by Node stdout; native-callable and javascript-callable caught by exact exit-70 pins; plain-error-invention caught by compile-refusal pins.
Uncovered: general bind, receiver-bearing producers, mutable view aliases, parameter-carried context adapters, aggregate/rest signatures, and inherited global counts/errno runtime failures.

The two existing good.a files are unchanged. Their declarations remain optional methods, not function properties. The admission relaxation requires an existing readonly unknown source slot with the same name, a supported optional method signature, and the original reverse relation for every other source member. It refuses nominal class relations, dictionaries, and invented target fields. Writable-slot proofs still run. Cast admission supplies no callable certificate: every reached read retains the existing presence, initialization, physical-kind, and independently recorded producer-signature checks.

The context adapters follow const local views and const aliases of those views. The bind form is exactly host.method?.bind(host), with the same identifier receiver and this argument, one supported scalar parameter, and a supported scalar result. It checks the original read before creating a fresh closure wrapper; the identity control matches Node's distinct bind identities. Only receiver-free implementation producers have usable existing certificates. Receiver-bearing literal methods retain their existing compile refusal. The condition form keeps host.method && a declared function. Reading that hoisted function identity has no effects; its use remains conditional on the checked optional method being present. Ordinary optional methods and detached methods retain their refusals.

Assumption: the conservative supported subset above is this unit's admission boundary. Views retain their existing transitive checks, and the new bind/condition contexts work through local const aliases; general helper-parameter or mutable alias context admission is not claimed. No other lane was merged. The delivery branch starts at the specified callable area tip and contains only this unit's commits. The four protected files are unchanged. Nothing was copied from cohere.

Plain Error assertions that invent optional host fields are refused before deferred cast admission, including both data and method controls. An inherited lower test expected admission for new Error as NodeJS.ErrnoException; its expectation now follows the October 8 ruling. Its independently typed NodeJS.ErrnoException producer remains admitted. The unchanged-base filesystem runtime defect described below is separate from this assertion refusal.

Observed controls (Node, sanitized native, release native, JavaScript):

| Probe | Present | Missing field | Initialized undefined |
|---|---|---|---|
| realpath-binding | abc/3 | none | none |
| directory-condition | true | false | false |

All positive controls, alias controls, and the bind identity control match Node's stdout, stderr, and exit status. Positive native runs pass the leak check. Fifteen fixtures are registered and measured in counts.md; thirteen are new .a files. The absent controls omit the actual own slot, separately from the initialized-undefined controls.

Wrong-type and wrong-signature controls deliberately exercise Adamic's stronger read contract. Node observations are logged separately from Adamic's required failures. Adamic emits no stdout and exits 70, with full stderr equality across sanitized native, release native, and JavaScript. The wrong-type pins are:

```
adamic: panic: field read failed: host.realpath matches no member of ((path: string) => string) | undefined; expected ((path: string) => string) | undefined, found number
adamic: panic: field read failed: host.directoryExists matches no member of ((path: string) => boolean) | undefined; expected ((path: string) => boolean) | undefined, found number
```

Wrong-signature pins name the same field and declared signature and end with "found function with incompatible result representation". The non-invoking signature controls preserve the original bind/condition read but observe the wrapped value's presence, making omission mutants safe to execute without calling an incompatible ABI.

Every mutant runs independently and restores its source in finally. The final runner rejects build failures and requires semantic assertion failures. The four runtime mutants compile and execute with exit 0 and empty stderr, then lose a pinned oracle expectation:

| Mutant | Mutation | Catch |
|---|---|---|
| native-presence | Skip the slot presence predicate and treat every optional read as absent | Both good probes become none/false; Node stdout comparison fails |
| javascript-presence | Skip Object.hasOwn presence predicate and treat every optional read as absent | Same Node stdout failure |
| native-callable | Return present closures without recorded-signature validation | Both wrong-signature-read controls print present instead of exit 70 |
| javascript-callable | Return functions without recorded-signature validation | Same exact exit-70 failure |
| plain-error-invention | Disable the plain Error assertion boundary | Both invented-field controls lower successfully; compile-refusal pins fail |

Exact final commands, with output written directly to log files:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/views-optional-host-setup-final.log 2>&1
source /workspace/adamic-tools/env.sh
npm ci --prefix stage3/api > /tmp/views-optional-host-node-types.log 2>&1
go test ./internal/lower -run '^(TestOptionalHostMethodBoundaries|TestNodeFSFileQualifiedErrorType|TestNodeFSFileKeepsDetachedMethodRefusal|TestUncheckableCastsStayRefused|TestCheckedCastProofAndElision)$' -count=1 -v > /tmp/views-optional-host-restored-lower.log 2>&1
go test ./internal/oracle -run '^(TestCheckedViewOptionalHostMethods|TestCheckedViewNullishCallableArity)$' -count=1 -v > /tmp/views-optional-host-restored.log 2>&1
python3 stage3/interface-downcasts/lane5/optional-host/run-mutants.py > /tmp/views-optional-host-mutants.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts > /tmp/views-optional-host-counts-global.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -args -update-counts > /tmp/views-optional-host-counts-scoped-final.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 > /tmp/views-optional-host-counts-verify.log 2>&1
go test ./internal/oracle -run '^TestNodeFSFileAgreesWithNode$/close$' -count=1 -v > /tmp/views-optional-host-certified-host.log 2>&1
git diff --check
```

Final lower output: PASS, 1.623s. Final unit oracle plus the inherited nullish callable arity test: PASS, 6.654s. Lane counts update: PASS, 182.819s; verification: PASS, 129.825s. Only these named tests were run; no whole-package test or full gate was run. Earlier focused runs exposed an incorrect ir.Conditional field name, a Go fixture quoting error, missing Node type dependencies, and the original invented Error admission; those were corrected. A restored test run overlapped a mutant run and was discarded; the final passing tests ran after confirmed mutant completion. Final measured count rows were independently verified against restored production sources. New rows were placed in the runtime table without changing any measured values.

The mandatory global updater fails on 39 fixture names, including pre-existing lowering/ABI boundaries and filesystem controls without their host input setup. An isolated detached worktree at base 926a1d39 ran the identical command with the same pinned cohere submodule referenced through a symlink and the same locked Node dependencies. It failed on exactly the same 39 names: zero new or removed names. Current duration 50.984s, base duration 89.929s. See logs/counts-comparison.json and both full logs. This is observed failure-name equality, not a claim that the full gate is green.

The scoped certified filesystem close oracle fails in native with:

```
adamic: panic: field read failed: errno.code matches no member of string | undefined; expected string | undefined, found unsupported representation
```

It has identical stdout, stderr, and exit behavior on the unchanged base, with current duration 0.650s and base duration 0.643s. The qualified producer lower test passes. No change to the host producer/runtime certificate mechanism is made here. This inherited runtime defect remains for integration.

Setup timing lines (successful final setup):

```
setup: go ready (0.017s)
setup: node ready (0.019s)
setup: submodules ready (0.052s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.006s
setup: markdown dependencies ready (0.063s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.163s)
setup: go build ready (27.389s)
setup: test binaries deferred (use --warm-tests) (27.488s)
setup: build cache warm (27.489s)
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (27.516s)
```

nproc returned 5. Versions: Go 1.27.1, Node v24.19.0, clang 20.1.8. The first setup fetched pinned submodules in 232.162s, then its warm build failed because source enumeration overlapped creation of the new helper: optionalHostViewRelation undefined in view_cast_preflight.go and view_objects.go. The retry failed on unknown ir.Conditional fields Then and Else. The helper was completed, fields corrected to WhenTrue/WhenNot, and setup rerun successfully after sources stabilized. Full initial/retry/final logs are retained. npm ci installed the locked three packages in 867ms, resolving missing @types/node 25.3.3 in the scoped lower tests.

Run the mutant script only when no other compiler/test process is reading this worktree, since it temporarily edits producer/check sources. It uses /tmp logs by default or ADAMIC_OPTIONAL_HOST_LOGS. The checked-in logs contain the final independent results and baseline comparisons. Delivery uses the requested branch only; integration remains responsible for merging and the shared fast gate.
