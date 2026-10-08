# Item 18 stopped: wrong output with exit 0

Pin: codex/library-error-value 6469f37bc02f5ce41b5735a966e68e974377c388, found with git branch -r --contains. Candidate base: 7a1fd0a95cee3059533fa7a7f5adbe9a7787c014. Last green compiler integration: 58270ec655f020bc7e3d20e37469fc3b942af56b.

The unmodified candidate compiler silently loses the observable failure of Error.captureStackTrace on a frozen target. The Go overlay adds a test only; it does not mutate compiler source. This triggers the unit's instruction: a wrong output with exit 0 is the top finding; stop and report it.

| Observation | stdout | stderr | exit |
| --- | --- | --- | --- |
| Node 24.19.0 | caught\n | empty | 0 |
| Native, repository sanitizer harness | completed\n | empty | 0 |
| JavaScript IR backend | completed\n | empty | 0 |

The source is capture-frozen.a and the actual uncached run is capture-frozen.log. Source SHA-256: 1776db68dfd9affdc6a62a29efe356244f8467b9b366ff7009fd551b35bee27e. Both compiler backends disagree with Node and exit successfully. No release-binary observation is claimed for this probe.

The incoming library capture function has an empty body. Node attempts to define the stack property and throws for this frozen target; the candidate's empty function completes instead. Stack reads remaining NotYet does not remove the observability of that exception through catch.

Item 18 was not merged. The active candidate was saved before git merge --abort, restoring the green compiler source tree. resolution.patch is an artifact, not applied code. Patch SHA-256: 9225402ee3e942b918968f87a6f8280cf6855260c9ab7d9875aaacd42c733a2d. It includes the integration resolutions and Linux-generated counts from the candidate, plus the subsequent small fixes. It is not a green merge.

The integration resolutions preserve rooted loader paths, all shared call-target queries, captured and namespace readiness storage, explicit readiness stop messages, the optional-field Error-to-ErrnoException refusal, and named unsupported Node members. They restrict the as-any receiver exception to the exact library capture intrinsic. Accessor targets survive repeated class-dispatch registration. The closed-input proof excludes only fresh nominal built-in Error initializers, retains ordinary mutation refusals, and traverses pure absence checks and throws. Later fixes classify the deliberately refused stack probe with flow's other refusal fixtures and teach freshness that a phantom member retains receiver effects but holds no receiver. These fixes are preserved only in the excluded candidate patch.

Earlier Linux counts regeneration passed in 184.935s, before the later freshness fix. Earlier uncached lowering rerun passed in 269.558s. The first full gate failed a stale fake-Error diagnostic, deliberate-refusal fixture classification in flow, and disk exhaustion in native/IR. The 31GB generated Go build cache was cleared with go clean -cache, freeing 28GB. The unmodified write-site baseline separately found the missing PhantomMember handling. A fresh full gate and baseline were started after those fixes, then interrupted at the exit-zero disagreement. Item 18 has no successful final full gate and no complete final baseline claim.

Source overlay mutants and their exact catchers are in mutants/ and the combined MUTANTS.md report. They are candidate-only evidence. The original throwing-capture attempt failed to build and was excluded; its adapted nominal throw was caught. The original accessor anchor no longer existed after centralizing dispatch preservation, and the adapted erasure was caught. A closed-input control initially used the wrong diagnostic matcher, then was caught by the actual cycle refusal. A stack-read control survived the other stack guard. The nominal fake-Error control survived the existing optional-field refusal; its earlier stale-diagnostic failure is not a valid kill. Interrupted controls and an invocation without the toolchain environment are not successful proof.

To reproduce in a disposable checkout containing these artifacts, apply resolution.patch to the compiler tree at the candidate base, then oracle-probe.patch. Source the printed toolchain environment and run:

```sh
export GOPROXY='https://proxy.golang.org|direct'
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestFront3CaptureFrozenProbe$' -count=1 -timeout 10m -v > /tmp/front3-capture-frozen-repro.log 2>&1
```

The expected result is FAIL with the three observations above. This is a real candidate disagreement, not an intentional compiler mutant. Items 19, 20, 22 and current-main landing were not attempted after it. Item 17 remains held for its 22 carried host lowering failures, recorded in ../item17/REPORT.md. Item 21 remains held pending clearance of its carried views-integration pin.

Current-main check after the user authorized continuation: freshly fetched origin/main bfa0bfec9ab6f49c81e22984d3fca4ee21a472e8 does not print completed. With compiler sources unchanged in a detached worktree and a test-only Go overlay, independent Node prints caught\n with empty stderr and exit0. The main checker rejects this exact source with TS2591 for node:fs and TS2339 for Error.captureStackTrace. No main native/JavaScript program was produced. The direct-node harness avoids cachedNode's source-loading prerequisite; the failed Go test reports the checker rejection, not a wrong-output main binary. Raw evidence: current-main-capture-frozen.log. Source branch remains codex/library-error-value, exact pinned commit 6469f37bc02f5ce41b5735a966e68e974377c388, still excluded.
