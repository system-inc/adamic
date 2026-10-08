Built: integration merge reconciliation and bounded required-undefined presence checks; merge remains uncommitted because original certificates regress.
Commits: starting tip 990007bccb58642c1565bfcb687e43b707518aef; pending integration a2ab63f0ff7783312b4a4a73b760686993b8bac8; no new commit or push.
Commands and outputs: focused merged regression PASS 111.731s; final presence PASS 1.827s; IR/lower units PASS 0.007/0.160s; original certificates FAIL at lowering.
Mutants: bounded required-field omission caught by sanitized native, release native and JavaScript; original pair mutant driver stops on the admission regression and proves no new certificate.
Not covered: remaining lane 7 pairs, lane 4b union overlap recertification, lane 4 intersected Identifier, and complete original recertification after integration.

Working date: October 11, 2026.

The branch resolves the requested abbreviated starting tip exactly. The newest
integration tip was fetched explicitly. No peer branch was merged. Peer branches
were fetched only to read the newly assigned blockers.

The integration tip already combines lane 7's nullish hook and callable kind
check with nullable member selection, Map certificates and callable certificates
in javascript/readiness.go, javascript/view_nullish.go and native/view_nullish.go.
The actual conflicts were docs/checked-views-plan.md and lower/view_lazy.go.
The plan retains both append sections. Producer certification and untagged
recursive completion now precede finishBoundedIntersections.

The automatic merge also mismatched integration's new field-table layout.
The bounded JavaScript and native table builders now carry optional presence
and permitted undefined payload as separate flags. The native bounded runtime
uses adamic_object_optional_view_undefined, matching the recursive walker.
The existing presence oracle adds two bounded witnesses through an array child.
The missing-field control panics, the initialized-undefined control agrees with
Node, and the successful native control passes leak checking. The presence
mutant marks required label fields optional. Each backend then prints true and
exits zero, and all three independent missing-field pins catch it.

The bounded classifier now recognizes the new ViewMap descriptor and its
undefined-only ViewNullable wrapper. Map entries and certificates retain their
own read boundaries. Dictionary descriptors remain deferred to their own lookup.
Null-bearing nullable descriptors remain refused. Temporary diagnostic tracing
has been removed from production files; its log remains evidence.

Blocking observation: integration now admits untagged FlowNode and ConciseBody
unions. These descriptors were deferred while unsupported before this merge;
the bounded classifier instead reaches its existing BoundedRefused case after
their owning adapter clears Unsupported. Original Bindable, emitNode, heritage
and JSDoc fixtures therefore refuse demanded reads at lowering. The final
original rerun ends FAIL in 30.170s. This is a safe refusal regression, not an
observed wrong native output. Tracker controls pass in that run.

Proposed next change, NOT APPLIED: at a supported untagged object-union descendant
inside a bounded walk, retain presence and physical object-kind checks and leave
member selection to that union's owning checked read. Root union admission would
still require its existing selected-arm proof. This follows the lane's bounded
read model, but needs acceptance and regression/mutant evidence before any new
certificate is claimed.

Automatic approval review rejected that proposed code edit, stating:
"The scripted replacement appears to remove a descriptor-type refusal and
replace it with unconditional kind admission, potentially weakening soundness
and causing silent miscompiles; the user authorized preserving checks, not
dropping this guard." The rejected command did not run. The existing untagged
refusal remains. No alternate command applies the rejected edit. Approval is
needed to proceed with that specific proposal; otherwise a deeper union adapter
must preserve the full current descendant obligation.

Coverage accounting:
- Starting tip: 10 original pairs / 56 reads certified, 7 / 10 remaining.
- No new original pair/read certified in this session. The merged tree is not
  recertified; nine historical pairs / 55 reads remain affected by the regression.
- Lane 4b's seven assigned union/intersection rows total 45 reads. Six rows / 44
  reads overlap existing lane 7 certificates: 9474, 7642, 9476, 9485, 9475, 36241.
  The seventh is 7644 / one read. Do not add this coverage twice.
- Still pending: 7644, 97923, 98493, 92175, 9454 and 9657. 68704 remains excluded
  from certification because its original SymbolTrackerImpl declaration is not
  exported. Lane 4's 9477 escapedText receiver remains pending.
- Exact production allocation reachability and whole-tsc compilation unmeasured.

Commands, each with direct log-file output under integration-evidence:
1. export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh.
   Initial setup fetched submodules in 182.127s, then failed during go list on
   the pre-reconciliation recursiveIntersectionField arity. Retry after repair
   PASS: Go .021s, Node .030s, submodules .075s, markdown .092s, clang .166s,
   build 22.706s, total 22.850s. nproc 5, cgroup quota 4 CPUs. Environment sourced
   from /workspace/adamic-tools/env.sh.
2. node stage3/adapt/00-setup/adapt.cjs /tmp/lane7-upstream;
   node stage3/interface-downcasts/lane4b/original/prepare.cjs
   /tmp/lane7-upstream /tmp/lane7-declarations;
   node stage3/interface-downcasts/lane7/original/prepare.cjs
   /tmp/lane7-upstream /tmp/lane7-declarations.
   PASS: 78 declarations; 77 exact original spans, pinned 050880ce59e30b356b686bd3144efe24f875ebc8.
3. go test ./internal/lower ./internal/ir ./internal/javascript ./internal/native
   ./internal/oracle -run '^(TestViewIntersection.*|TestIntersection.*|TestUntaggedView.*|TestCheckedView(Intersection.*|Nullish.*|MapCertificate.*|MapPhantom.*|MapEntry.*|Untagged.*))$' -count=1.
   Initial oracle FAIL 210.346s with a pre-runtime-repair binary; lower/IR pass;
   JavaScript/native report no matching tests. This is superseded by run 4.
4. go test ./internal/lower ./internal/ir ./internal/oracle -run
   '^(TestViewIntersection.*|TestIntersection.*|TestBoundedIntersection.*|TestUntaggedView.*|TestCheckedView(Intersection(Required.*|Source|ProductionMutants|RecursiveDemand|SelectedArms)|Nullish.*|MapCertificate.*|MapPhantom.*|MapEntry.*|Untagged.*))$' -count=1.
   PASS lower .600s, IR .009s, oracle 111.731s.
5. ADAMIC_INTERSECTION_ORIGINAL_DECLS=/tmp/lane7-declarations go test
   ./internal/oracle -run '^TestCheckedViewIntersectionOriginal' -count=1 -v.
   Original rerun FAIL 30.170s as described above. First uncached run FAIL 50.550s.
6. go test ./internal/oracle -run
   '^TestCheckedViewIntersectionRequiredUndefinedPresence$' -count=1 -v.
   Final PASS 1.827s; first repaired run PASS 25.225s, including leak checks.
7. ADAMIC_INTERSECTION_PRESENCE_MUTANT=1 go test ./internal/oracle -run
   '^TestCheckedViewIntersectionRequiredUndefinedPresence/bounded-missing$'
   -count=1 -v. Expected FAIL .528s, three successful mutant outputs caught.
8. python3 stage3/interface-downcasts/lane7/original/run-pair-mutants.py
   /tmp/lane7-declarations /tmp/lane7-merge-mutants.
   FAIL at the first left-skip run because lowering refuses before the mutant.
   No claim that those 28 mutants have been proved on the merged tree.
9. go test ./internal/ir ./internal/lower -run
   '^(TestBoundedIntersectionContracts|TestIntersectionUnionArmsCoalesceCopies|TestViewIntersectionContracts|TestUntaggedView.*)$' -count=1.
   PASS .007s / .160s. Changed Go files have empty gofmt -l output;
   git diff --check clean. No whole package tests or full gate run.

No new standalone .a fixture was added, so no new counts.md row is introduced.
No protected emission/oracle orchestration file was edited. No cohere source was
copied. Merge is pending in the worktree, with conflicts resolved but original
regressions unsolved. No failing merge was committed or pushed.
