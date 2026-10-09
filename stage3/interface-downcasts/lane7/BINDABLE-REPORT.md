Built: bounded intersection reads with tagged arm selection; four original Bindable pairs certified against complete tsc declarations.
Commit: this batch follows 9e4e13fb; its sha is recorded by the branch tip. Not pushed from this session (proxy 403); delivered as a format-patch.
Commands: original pairs PASS 94.1s; uncached ^TestCheckedView PASS 149.7s; ir and javascript PASS; vet and gofmt clean; lower and native full packages fail 12 tests that fail identically at 9e4e13fb.
Mutants: skip, shape and nested for each of the left, expression and access reads (12 runs), each caught by three pins; coalescing mutant caught by a unit test and a refusal.
Not covered: 9454 argumentExpression (its read is always dominated), 11 more candidate pairs / 31 reads, Weak parent cycles, integration merge (code conflicts).

Working date: October 11, 2026. Static candidates: 17 object pairs / 66 reads.
Certified with complete pinned declarations: 5 pairs / 34 reads (10236 from the
previous group; 9476 left 9, 9474 expression 11, 9485 left 9, 9475 expression 4).
Remaining: 12 pairs / 32 reads, listed in lazy-pair-progress.json. Exact
allocation reachability remains unmeasured.

What I observed first. With complete declarations, each read of these pairs was
refused as "union intersection": every arm is recursive, because Node.parent is a
required Node and every Declaration needs a Symbol. The existing recursive runtime
walks data until an (object, contract) pair repeats, so an honest acyclic value
can never satisfy it: some parent chain ends. A cycle needs a Weak parent, and a
probe showed a Weak slot read through a view refuses as "representation
conversion". The original binder code itself says why the declarations lie here:
"Fix up parent pointers since we're going to use these nodes before we bind into
them" (binder.ts bindSpecialPropertyAssignment).

What I built. Contracts lane 7 refused (union intersection, recursive, mixed or
compound payloads) are reconsidered once every descriptor is complete. A read is
admitted when the bounded walk validates its whole reachable graph. The walk
checks every descriptor once per path, as the existing inline field walk does:
presence, kind and literals at each level; tagged object unions by their
selected arm; plain unions by their discriminant; arrays, Maps and callables by
kind. A descriptor already entered on the path, array elements, mixed scalar
unions and refused families keep their own checked reads. Nullable, nominal and
tuple members refuse the read. The pure object recursive walk is unchanged.

Two things the first attempt got wrong, both caught by fixtures. The checker
gives Identifier and LeftHandSideExpression & Identifier separate descriptors
whose fields differ only by optional copies of one canonical EmitNode and an
incomplete JSDocArray reservation; arms now coalesce when their obligations are
the same to the walk. Lowering and the emitters could also have disagreed about
a refused descriptor reachable from an admitted root; admission now runs to a
fixpoint and walks a refused descriptor only when it is admitted too.

Fixtures (stage3/interface-downcasts/lane7/original/bindable-*.a) import the
complete emitted declarations. Good: property access, a three-name chain, the
identity/stored/forEach helper path, element access with a string literal, and
computed element access. Wrong, each with an exact exit 70 message in sanitized
native, release native and JavaScript: missing Symbol on the left's name, a left
kind of 999, a missing Symbol one level into the chain, the leftmost expression
being this (kind 110), a non-literal static element argument, this as an element
receiver, and a missing receiver Symbol. tsc's own isBindableStaticAccessExpression
admits this.x, whose expression is no EntityNameExpression: Adamic stops it at the
read. That comes from reading upstream source, not from running tsc.

The chain fixtures separate the two reads of one path. The left read enters the
property access descriptor once, so it checks the prefix's tag and leaves the
prefix's fields; only the expression read reaches them. Each mutant's fixture is
wrong only where its read is the first to look.

Mutants, run by original/run-bindable-mutants.py. Each fixture then runs to
"true" and exit 0 in all three modes, failing its exit 70 pin three times:
- left-skip / access-skip: the read falls back to the plain union discriminant.
- left-shape / access-shape: the selected arm accepts kind 999.
- expression-shape (twice): the Identifier arm accepts kind 110.
- left-nested / access-nested / expression-nested (twice): Identifier forgets symbol.
- expression-skip (twice): the expression read falls back to its discriminant.
- Coalescing by id only: TestIntersectionUnionArmsCoalesceCopies fails, and
  bindable-element-good is refused at compile time (safe direction).

9454 BindableStaticElementAccessExpression.argumentExpression is not counted. Its
contract is held: the non-literal argument fixture stops at the left read. But a
direct cast to that intersection, and the type predicate the original site uses,
are both refused, so no admitted program reaches its read without an earlier read
checking the same argument. A skip mutant there could not be caught.

Commands, all output captured to logs in bindable-evidence:
- bash cloud/setup.sh: first run failed when a download hit "Connection reset by
  peer"; the retry finished in 498.221s on 4 processors (clang 213.017s, go build
  497.999s). nproc 4.
- Pristine microsoft/TypeScript fetched at 050880ce; 00-setup/adapt.cjs, lane4b and
  lane7 prepare.cjs: 78 declarations, 77 original spans, pairs 1+4+9+11+9+4+1.
- go test ./internal/oracle -run '^TestCheckedViewIntersectionOriginal' -count=1 -v:
  PASS 94.125s (8 tracker and 17 Bindable subtests).
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedView' -count=1: PASS 149.670s.
- go test ./internal/lower ./internal/ir ./internal/javascript ./internal/native -count=1:
  ir PASS 29.688s, javascript PASS 1.707s; lower FAIL (7 tests) and native FAIL (5
  tests). The same 12 tests fail on a clean 9e4e13fb worktree (baseline-failures.log).
- go vet ./...: clean. gofmt -l cmd internal: empty.
- The full repository gate was not run.

Integration: merging codex/views-integration 6a7f1bf3 conflicted in
javascript/readiness.go, javascript/view_nullish.go and native/view_nullish.go,
where lane 7's nullish intersection hook and callable kind meet integration's
nullable member selection, Map and callable certificates. As briefed, the merge
was aborted unresolved and this batch is built on 9e4e13fb alone.
