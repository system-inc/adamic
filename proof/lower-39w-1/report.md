# Compiler proof: #39wtqdj

HEAD: `81525c8b690b9af1662d8275b059eca547faf86c`. Base main: `16429a30452d027809f0ea85bc395479423e493e`. Measured 2026-11-11 UTC; all checks completed by 11:14 UTC.

## Decisive results

The proof is **red**. HEAD’s selected lowering pins pass, but its oracle pins fail: `TestPredicateFalseRead/{spread,again,getter}` refuse during lowering and `TestCheckedPredicateOracle/true` disagrees with its expected counts. Recorded counts pass at HEAD. Both full suites fail; admission passes.

HEAD fixes the observed `no_read` behavior: all three backends agree with Node. Main checks its false side and panics (exit 70) before printing either version line. Each requested mutant is caught by a lowering pin: M1 by the four no-read/replacement/merge cases and the runtime false branch, M2 by the capture of a let assigned later, M3 by the merged read.

`admission exit=0 wall_seconds=1397`

Admission: verdict `pass`, void `false`, diff_count `10`, admitted `0`, newly refused `0`.

| Step | Result | Exit / wall seconds |
|---|---|---|
| Formatting / vet | PASS | vet: 0 / 225 |
| HEAD lower pins | PASS | 0 / 53 |
| HEAD oracle pins | FAIL | 1 / 318 |
| Main copied false-read tests | FAIL (expected no_read regression plus three shared refusals) | 1 / 279 |
| M1 lower / oracle | FAIL / FAIL | 1 / 37; 1 / 92 |
| M2 lower / oracle | FAIL / FAIL | 1 / 35; 1 / 86 |
| M3 lower / oracle | FAIL / FAIL | 1 / 31; 1 / 94 |
| Full lower | FAIL: six leaf tests | 1 / 1515 |
| Full oracle | FAIL: twenty leaf tests | 1 / 3738 |
| Admission | PASS, no admission delta | 0 / 1397 |

## Setup and isolation

Loaded `/workspace/adamic-tools/env.sh` before repository work. Created its configured `/tmp/adamic-gate` outside the repository. Fetched `origin compiler/predicate-false-read-1`, checked out the requested HEAD detached, and ran `git submodule update --init --recursive`. Initial tracked/untracked status was clean. No compiler changes were committed or pushed. Heavy commands ran sequentially; the admission gate ran alone with `--workers 2`.

Provisioned upstream TypeScript with a depth-one fetch of `050880ce59e30b356b686bd3144efe24f875ebc8` from `https://github.com/microsoft/TypeScript.git` into `/tmp/adamic-typescript-050880`, checked out detached, and exported `ADAMIC_TYPESCRIPT_SOURCE` to that directory. Copied HEAD’s `stage3/api/package.json` and lockfile into `/tmp/adamic-stage3-api` and ran `npm ci` there. Its `node_modules/.bin` was placed on PATH.

````text
81525c8b690b9af1662d8275b059eca547faf86c
 5360753e19e4417675a90e92a1823bea66b13918 cohere (5360753e1)
 d92d9bfee114c80be2c375d72edae966176e3a4f cohere/TypeScript (remotes/origin/adamic)
050880ce59e30b356b686bd3144efe24f875ebc8
go version go1.27.1 linux/amd64
v24.19.0
11.9.0
Version 6.0.3
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
````

The proof command environment:

````text
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-typescript-050880
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export npm_config_strict_ssl=true
export PATH=/tmp/adamic-stage3-api/node_modules/.bin:$PATH
````

The repository’s `toolchain.Env`/`GitEnv` filter environment variables. For admission, external wrappers in `/tmp/adamic-proof-tools` re-exported the three CA settings and exec’d the original tools (node, npm, git, go, tsc, python3, clang, clang++, llvm-symbolizer); `ADAMIC_TOOLS` pointed to the wrapper tree, its bin directories led PATH, and its `llvm` link pointed to the original LLVM directory. This preserved the pinned binaries and TLS verification without editing repository toolchain code.

````text
source /tmp/adamic-proof/env.sh
export ADAMIC_TOOLS=/tmp/adamic-proof-tools
export PATH=/tmp/adamic-proof-tools/bin:/tmp/adamic-proof-tools/go/bin:$PATH
````

`gofmt -l internal/lower internal/oracle` printed nothing. `go vet ./internal/lower ./internal/oracle` exited 0:

`vet exit=0 wall_seconds=225`

````text

````

## Exact pinned commands

````sh
go test ./internal/lower -run 'TestCheckedPredicateFalseReadDirections|TestCheckedPredicateBranches|TestCheckedPredicatePendingContracts|TestPredicateUseRegions|TestPredicateUsesBelongToEachCall' -count=1 -v -timeout 30m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestPredicateFalseRead|TestCheckedPredicateOracle|TestCheckedPredicateCountsAreRecorded' -count=1 -v -timeout 40m
````

### Expectations omitted by some fatal diagnostics

Some tests print the entire measured object but no explicit want, or fail during lowering before a runtime got exists. The following fixture definitions are copied verbatim from HEAD so the whole expected count/output contract remains reviewable alongside the verbatim logs. For a lowering refusal the got is the diagnostic, the immediate want is successful lowering (`err == nil`), and no backend output was measured by that failing subtest.

````go
var checkedPredicateFixtures = []struct {
	name, nodeOut, checkedOut, branch, source, target, guard string
	checked, unobservable                                    int
}{
	{"true", "claimed\n", "", "true", "string | number", "number", "", 2, 0},
	{"true_unused", "continued\n", "", "true", "\"text\"", "number", "", 1, 1},
	// isNumber lies on its false side, and nothing reads value there, so the false check is dropped and
	// counted unobservable, and the program prints as Node does (#39wtqdj, Adamic's ruling of Nov 11).
	{"false", "continued\n", "", "", "", "", "", 1, 1},
	{"false_alias", "continued\n", "", "", "", "", "", 1, 1},
	{"false_compound_unchanged", "continued\n", "", "", "", "", "", 1, 1},
	{"asserts", "called\nasserted\n", "called\n", "asserts", "string | number", "number", "", 1, 0},
	{"valid", "argument\ncalled\nnumber\nargument\ncalled\nstring\n", "", "", "", "", "", 1, 1},
	{"asserts_valid", "called\n7\n", "", "", "", "", "", 1, 0},
	{"condition_valid", "continued\n", "", "", "\"truthy\"", "truthy", "", 1, 0},
	{"condition_lie", "continued\n", "", "asserts", "0", "truthy", "", 1, 0},
	{"condition_once", "argument\ncontinued\n", "", "", "number", "truthy", "", 1, 0},
	{"condition_once_lie", "argument\ncontinued\n", "argument\n", "asserts", "number", "truthy", "", 1, 0},
	{"false_unchanged", "continued\n", "", "", "", "", "", 1, 1},
	// tsc's Debug.assertNotNode (debug.ts:307): the bodyless overload claims Exclude<T, U> from the
	// call's resolved signature, and the implementation, gated by an assertion level as tsc's is,
	// never runs the test. Its own claim is checked at the call, so a Keyword stops it there.
	{"assert_not_node", "asserting\nnarrowed\n", "asserting\n", "asserts", "Node", "Identifier", "", 1, 0},
	// Language's false-branch forms (#7v4rjr7): the checker narrows each one's false side, and isS
	// says false to a Statement, so the false check must stop it before n.e is read.
	{"false_switch_true", "undefined\n", "", "false", "Node", "Statement", "adamic: panic: compiler bug: a field the checker proved is there is missing\n", 2, 0},
	{"false_eq_false", "undefined\n", "", "false", "Node", "Statement", "adamic: panic: compiler bug: a field the checker proved is there is missing\n", 2, 0},
	{"false_ne_true", "undefined\n", "", "false", "Node", "Statement", "adamic: panic: compiler bug: a field the checker proved is there is missing\n", 2, 0},
	{"false_false_eq", "undefined\n", "", "false", "Node", "Statement", "adamic: panic: compiler bug: a field the checker proved is there is missing\n", 2, 0},
	{"false_comma", "undefined\n", "", "false", "Node", "Statement", "adamic: panic: compiler bug: a field the checker proved is there is missing\n", 2, 0},
}
````

````go
for _, probe := range []struct{ name, nodeOut, where, checkedOut, source string }{
		{"no_read", "1.2.3\n1.2.3+build.5\n", "32:7", "", ""},
		{"false_read", "some\nobject\n", "23:7", "some\n", "readonly string[] | undefined"},
		{"else_read", "some 1\nnone object\n", "23:6", "some 1\n", "readonly string[] | undefined"},
		{"early_exit", "1\nobject\n0\n", "23:6", "1\n", "readonly string[] | undefined"},
		{"capture", "some\nobject\n", "23:7", "some\n", "readonly string[] | undefined"},
		{"renarrowed", "0\n", "6:6", "", "readonly string[] | undefined"},
		{"getter", "some\nobject\n", "36:7", "some\n", "readonly string[] | undefined"},
		{"spread", "some\n0\n", "23:7", "some\n", "readonly string[]"},
		{"again", "some\nnone\n", "24:7", "some\n", "readonly string[] | undefined"},
	} {
````

False-read cases want false direction `checked` except `no_read`, which wants `unobservable`. For `no_read`, backend output must equal Node (exit 0, empty stderr, the two version lines). For each other case, the expected backend exit is 70; stdout is `checkedOut` above, and stderr is exactly the following format, with the absolute fixture path, pinned location and source type:

````text
adamic: panic: predicate some at <absolute fixture path>:<where> false branch: source <source>, target readonly string[]
````

For `TestCheckedPredicateOracle/true`, the whole scalar expectation is `Proven=0, Checked=2, Unobservable=0, len(Sites)=1`; HEAD, M2 and M3 measure `Proven=0, Checked=1, Unobservable=1`. M1’s failing `false` and `false_alias` instead want `Proven=0, Checked=1, Unobservable=1, len(Sites)=1` and measure `Proven=0, Checked=2, Unobservable=0`. Full measured site objects appear verbatim below.

## HEAD pins

### head-lower

`head-lower exit=0 wall_seconds=53`

Per-test/subtest result:

````text
--- PASS: TestPredicateUsesBelongToEachCall (0.16s)
--- PASS: TestCheckedPredicateFalseReadDirections (0.01s)
    --- PASS: TestCheckedPredicateFalseReadDirections/empty_branches (0.16s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.16s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.13s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.13s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.27s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.23s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.23s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/semver's_toString (0.10s)
--- PASS: TestCheckedPredicatePendingContracts (1.03s)
    --- PASS: TestCheckedPredicatePendingContracts/untagged (0.16s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag (0.16s)
    --- PASS: TestCheckedPredicatePendingContracts/overload_in_.a (0.27s)
    --- PASS: TestCheckedPredicatePendingContracts/loose_false_comparison (0.12s)
    --- PASS: TestCheckedPredicatePendingContracts/open_enum_union_tag (0.11s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_read (0.06s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_pass (0.08s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_domain (0.07s)
--- PASS: TestPredicateUseRegions (0.01s)
    --- PASS: TestPredicateUseRegions/unused (0.19s)
    --- PASS: TestPredicateUseRegions/reversed_false_comparison (0.14s)
    --- PASS: TestPredicateUseRegions/comma_continuation (0.07s)
    --- PASS: TestPredicateUseRegions/true_inequality (0.05s)
    --- PASS: TestPredicateUseRegions/closure_read (0.10s)
    --- PASS: TestPredicateUseRegions/switch_true_default (0.13s)
    --- PASS: TestPredicateUseRegions/replacement (0.06s)
    --- PASS: TestPredicateUseRegions/loose_false_comparison (0.09s)
    --- PASS: TestPredicateUseRegions/false_comparison (0.10s)
    --- PASS: TestPredicateUseRegions/later_read (0.08s)
    --- PASS: TestPredicateUseRegions/loop_read (0.15s)
    --- PASS: TestPredicateUseRegions/computed_index (0.14s)
    --- PASS: TestPredicateUseRegions/indexed_spelling (0.12s)
    --- PASS: TestPredicateUseRegions/indexed_argument (0.17s)
    --- PASS: TestPredicateUseRegions/short_circuit (0.06s)
    --- PASS: TestPredicateUseRegions/compound_result_alias (0.07s)
    --- PASS: TestPredicateUseRegions/throw_continuation (0.12s)
    --- PASS: TestPredicateUseRegions/const_result_alias (0.08s)
    --- PASS: TestPredicateUseRegions/false_continuation (0.08s)
    --- PASS: TestPredicateUseRegions/both (0.09s)
    --- PASS: TestPredicateUseRegions/return_continuation (0.08s)
    --- PASS: TestPredicateUseRegions/true (0.06s)
    --- PASS: TestPredicateUseRegions/false (0.07s)
--- PASS: TestCheckedPredicateBranches (22.65s)
    --- PASS: TestCheckedPredicateBranches/true (20.37s)
    --- PASS: TestCheckedPredicateBranches/false (0.56s)
    --- PASS: TestCheckedPredicateBranches/false_read (0.61s)
    --- PASS: TestCheckedPredicateBranches/asserts (0.58s)
    --- PASS: TestCheckedPredicateBranches/valid (0.53s)
````

Complete output (including every whole got/want diagnostic):

````text
=== RUN   TestCheckedPredicateBranches
=== PAUSE TestCheckedPredicateBranches
=== RUN   TestCheckedPredicatePendingContracts
=== PAUSE TestCheckedPredicatePendingContracts
=== RUN   TestCheckedPredicateFalseReadDirections
=== PAUSE TestCheckedPredicateFalseReadDirections
=== RUN   TestPredicateUseRegions
=== PAUSE TestPredicateUseRegions
=== RUN   TestPredicateUsesBelongToEachCall
=== PAUSE TestPredicateUsesBelongToEachCall
=== CONT  TestCheckedPredicateBranches
=== RUN   TestCheckedPredicateBranches/true
=== CONT  TestPredicateUseRegions
=== CONT  TestPredicateUsesBelongToEachCall
=== CONT  TestCheckedPredicateFalseReadDirections
=== CONT  TestCheckedPredicatePendingContracts
=== RUN   TestPredicateUseRegions/unused
=== PAUSE TestPredicateUseRegions/unused
=== RUN   TestCheckedPredicateFalseReadDirections/empty_branches
=== PAUSE TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestCheckedPredicateFalseReadDirections/semver's_toString
=== PAUSE TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== PAUSE TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== RUN   TestCheckedPredicatePendingContracts/untagged
=== RUN   TestPredicateUseRegions/true
=== PAUSE TestPredicateUseRegions/true
=== RUN   TestPredicateUseRegions/false
=== PAUSE TestPredicateUseRegions/false
=== RUN   TestPredicateUseRegions/both
=== RUN   TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== PAUSE TestPredicateUseRegions/both
=== RUN   TestPredicateUseRegions/return_continuation
=== PAUSE TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== PAUSE TestPredicateUseRegions/return_continuation
=== RUN   TestPredicateUseRegions/throw_continuation
=== PAUSE TestPredicateUseRegions/throw_continuation
=== RUN   TestCheckedPredicateFalseReadDirections/false_branch
=== PAUSE TestCheckedPredicateFalseReadDirections/false_branch
=== RUN   TestCheckedPredicateFalseReadDirections/else
=== PAUSE TestCheckedPredicateFalseReadDirections/else
=== RUN   TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== PAUSE TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== RUN   TestCheckedPredicateFalseReadDirections/capture
=== PAUSE TestCheckedPredicateFalseReadDirections/capture
=== RUN   TestCheckedPredicateFalseReadDirections/renarrowed
=== PAUSE TestCheckedPredicateFalseReadDirections/renarrowed
=== RUN   TestPredicateUseRegions/false_continuation
=== PAUSE TestPredicateUseRegions/false_continuation
=== RUN   TestPredicateUseRegions/const_result_alias
=== PAUSE TestPredicateUseRegions/const_result_alias
=== RUN   TestPredicateUseRegions/compound_result_alias
=== PAUSE TestPredicateUseRegions/compound_result_alias
=== RUN   TestPredicateUseRegions/short_circuit
=== PAUSE TestPredicateUseRegions/short_circuit
=== RUN   TestPredicateUseRegions/closure_read
=== PAUSE TestPredicateUseRegions/closure_read
=== RUN   TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== PAUSE TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== RUN   TestCheckedPredicateFalseReadDirections/getter
=== PAUSE TestCheckedPredicateFalseReadDirections/getter
=== RUN   TestCheckedPredicateFalseReadDirections/spread
=== PAUSE TestCheckedPredicateFalseReadDirections/spread
=== RUN   TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== PAUSE TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== RUN   TestCheckedPredicateFalseReadDirections/conditional
=== PAUSE TestCheckedPredicateFalseReadDirections/conditional
=== RUN   TestCheckedPredicateFalseReadDirections/||_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/||_operand
=== RUN   TestPredicateUseRegions/indexed_spelling
=== PAUSE TestPredicateUseRegions/indexed_spelling
=== RUN   TestPredicateUseRegions/indexed_argument
=== PAUSE TestPredicateUseRegions/indexed_argument
=== RUN   TestPredicateUseRegions/computed_index
=== PAUSE TestPredicateUseRegions/computed_index
=== RUN   TestPredicateUseRegions/loop_read
=== PAUSE TestPredicateUseRegions/loop_read
=== RUN   TestPredicateUseRegions/later_read
=== PAUSE TestPredicateUseRegions/later_read
=== RUN   TestPredicateUseRegions/replacement
=== PAUSE TestPredicateUseRegions/replacement
=== RUN   TestPredicateUseRegions/false_comparison
=== PAUSE TestPredicateUseRegions/false_comparison
=== RUN   TestPredicateUseRegions/loose_false_comparison
=== PAUSE TestPredicateUseRegions/loose_false_comparison
=== RUN   TestCheckedPredicateFalseReadDirections/&&_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/&&_operand
=== RUN   TestCheckedPredicateFalseReadDirections/const_alias
=== PAUSE TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestPredicateUseRegions/true_inequality
=== PAUSE TestPredicateUseRegions/true_inequality
=== RUN   TestPredicateUseRegions/reversed_false_comparison
=== PAUSE TestPredicateUseRegions/reversed_false_comparison
=== RUN   TestPredicateUseRegions/switch_true_default
=== PAUSE TestPredicateUseRegions/switch_true_default
=== RUN   TestPredicateUseRegions/comma_continuation
=== PAUSE TestPredicateUseRegions/comma_continuation
=== CONT  TestPredicateUseRegions/unused
--- PASS: TestPredicateUsesBelongToEachCall (0.16s)
=== CONT  TestPredicateUseRegions/reversed_false_comparison
=== RUN   TestCheckedPredicatePendingContracts/open_tag
=== CONT  TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestCheckedPredicateFalseReadDirections/&&_operand
=== CONT  TestCheckedPredicateFalseReadDirections/||_operand
=== CONT  TestCheckedPredicateFalseReadDirections/conditional
=== RUN   TestCheckedPredicatePendingContracts/overload_in_.a
=== CONT  TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== CONT  TestCheckedPredicateFalseReadDirections/spread
=== CONT  TestCheckedPredicateFalseReadDirections/getter
=== RUN   TestCheckedPredicatePendingContracts/loose_false_comparison
=== CONT  TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== CONT  TestCheckedPredicateFalseReadDirections/renarrowed
=== CONT  TestCheckedPredicateFalseReadDirections/capture
=== CONT  TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== RUN   TestCheckedPredicatePendingContracts/open_enum_union_tag
=== CONT  TestCheckedPredicateFalseReadDirections/else
=== CONT  TestCheckedPredicateFalseReadDirections/false_branch
=== CONT  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_read
=== CONT  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== CONT  TestCheckedPredicateFalseReadDirections/semver's_toString
=== CONT  TestPredicateUseRegions/comma_continuation
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_pass
=== CONT  TestPredicateUseRegions/switch_true_default
--- PASS: TestCheckedPredicateFalseReadDirections (0.01s)
    --- PASS: TestCheckedPredicateFalseReadDirections/empty_branches (0.16s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.16s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.13s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.13s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.27s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.23s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.23s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/semver's_toString (0.10s)
=== CONT  TestPredicateUseRegions/closure_read
=== CONT  TestPredicateUseRegions/true_inequality
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_domain
=== CONT  TestPredicateUseRegions/loose_false_comparison
--- PASS: TestCheckedPredicatePendingContracts (1.03s)
    --- PASS: TestCheckedPredicatePendingContracts/untagged (0.16s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag (0.16s)
    --- PASS: TestCheckedPredicatePendingContracts/overload_in_.a (0.27s)
    --- PASS: TestCheckedPredicatePendingContracts/loose_false_comparison (0.12s)
    --- PASS: TestCheckedPredicatePendingContracts/open_enum_union_tag (0.11s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_read (0.06s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_pass (0.08s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_domain (0.07s)
=== CONT  TestPredicateUseRegions/false_comparison
=== CONT  TestPredicateUseRegions/replacement
=== CONT  TestPredicateUseRegions/later_read
=== CONT  TestPredicateUseRegions/loop_read
=== CONT  TestPredicateUseRegions/computed_index
=== CONT  TestPredicateUseRegions/indexed_argument
=== CONT  TestPredicateUseRegions/indexed_spelling
=== CONT  TestPredicateUseRegions/throw_continuation
=== CONT  TestPredicateUseRegions/short_circuit
=== CONT  TestPredicateUseRegions/compound_result_alias
=== CONT  TestPredicateUseRegions/const_result_alias
=== CONT  TestPredicateUseRegions/false_continuation
=== CONT  TestPredicateUseRegions/both
=== CONT  TestPredicateUseRegions/return_continuation
=== CONT  TestPredicateUseRegions/false
=== CONT  TestPredicateUseRegions/true
--- PASS: TestPredicateUseRegions (0.01s)
    --- PASS: TestPredicateUseRegions/unused (0.19s)
    --- PASS: TestPredicateUseRegions/reversed_false_comparison (0.14s)
    --- PASS: TestPredicateUseRegions/comma_continuation (0.07s)
    --- PASS: TestPredicateUseRegions/true_inequality (0.05s)
    --- PASS: TestPredicateUseRegions/closure_read (0.10s)
    --- PASS: TestPredicateUseRegions/switch_true_default (0.13s)
    --- PASS: TestPredicateUseRegions/replacement (0.06s)
    --- PASS: TestPredicateUseRegions/loose_false_comparison (0.09s)
    --- PASS: TestPredicateUseRegions/false_comparison (0.10s)
    --- PASS: TestPredicateUseRegions/later_read (0.08s)
    --- PASS: TestPredicateUseRegions/loop_read (0.15s)
    --- PASS: TestPredicateUseRegions/computed_index (0.14s)
    --- PASS: TestPredicateUseRegions/indexed_spelling (0.12s)
    --- PASS: TestPredicateUseRegions/indexed_argument (0.17s)
    --- PASS: TestPredicateUseRegions/short_circuit (0.06s)
    --- PASS: TestPredicateUseRegions/compound_result_alias (0.07s)
    --- PASS: TestPredicateUseRegions/throw_continuation (0.12s)
    --- PASS: TestPredicateUseRegions/const_result_alias (0.08s)
    --- PASS: TestPredicateUseRegions/false_continuation (0.08s)
    --- PASS: TestPredicateUseRegions/both (0.09s)
    --- PASS: TestPredicateUseRegions/return_continuation (0.08s)
    --- PASS: TestPredicateUseRegions/true (0.06s)
    --- PASS: TestPredicateUseRegions/false (0.07s)
=== RUN   TestCheckedPredicateBranches/false
=== RUN   TestCheckedPredicateBranches/false_read
=== RUN   TestCheckedPredicateBranches/asserts
=== RUN   TestCheckedPredicateBranches/valid
--- PASS: TestCheckedPredicateBranches (22.65s)
    --- PASS: TestCheckedPredicateBranches/true (20.37s)
    --- PASS: TestCheckedPredicateBranches/false (0.56s)
    --- PASS: TestCheckedPredicateBranches/false_read (0.61s)
    --- PASS: TestCheckedPredicateBranches/asserts (0.58s)
    --- PASS: TestCheckedPredicateBranches/valid (0.53s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	22.712s
````

### head-oracle

`head-oracle exit=1 wall_seconds=318`

Per-test/subtest result:

````text
--- PASS: TestCheckedPredicateCountsAreRecorded (48.21s)
--- FAIL: TestPredicateFalseRead (0.00s)
    --- FAIL: TestPredicateFalseRead/spread (42.71s)
    --- FAIL: TestPredicateFalseRead/again (42.71s)
    --- PASS: TestPredicateFalseRead/early_exit (233.35s)
    --- PASS: TestPredicateFalseRead/false_read (227.86s)
    --- PASS: TestPredicateFalseRead/else_read (233.37s)
    --- PASS: TestPredicateFalseRead/no_read (276.19s)
    --- FAIL: TestPredicateFalseRead/getter (0.40s)
    --- PASS: TestPredicateFalseRead/renarrowed (1.30s)
    --- PASS: TestPredicateFalseRead/capture (1.32s)
--- FAIL: TestCheckedPredicateOracle (297.01s)
    --- FAIL: TestCheckedPredicateOracle/true (42.66s)
    --- PASS: TestCheckedPredicateOracle/true_unused (234.33s)
    --- PASS: TestCheckedPredicateOracle/false (0.93s)
    --- PASS: TestCheckedPredicateOracle/false_alias (0.82s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (1.02s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.02s)
    --- PASS: TestCheckedPredicateOracle/valid (1.06s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (0.87s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (1.05s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.13s)
    --- PASS: TestCheckedPredicateOracle/condition_once (1.20s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.13s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (1.03s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.21s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.39s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.99s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.62s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.25s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.28s)
````

Complete output (including every whole got/want diagnostic):

````text
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
=== RUN   TestCheckedPredicateOracle
=== PAUSE TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateCountsAreRecorded
=== PAUSE TestCheckedPredicateCountsAreRecorded
=== CONT  TestPredicateFalseRead
=== RUN   TestPredicateFalseRead/no_read
=== CONT  TestCheckedPredicateCountsAreRecorded
=== CONT  TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateOracle/true
=== PAUSE TestPredicateFalseRead/no_read
=== RUN   TestPredicateFalseRead/false_read
=== PAUSE TestPredicateFalseRead/false_read
=== RUN   TestPredicateFalseRead/else_read
=== PAUSE TestPredicateFalseRead/else_read
=== RUN   TestPredicateFalseRead/early_exit
=== PAUSE TestPredicateFalseRead/early_exit
=== RUN   TestPredicateFalseRead/capture
=== PAUSE TestPredicateFalseRead/capture
=== RUN   TestPredicateFalseRead/renarrowed
=== PAUSE TestPredicateFalseRead/renarrowed
=== RUN   TestPredicateFalseRead/getter
=== PAUSE TestPredicateFalseRead/getter
=== RUN   TestPredicateFalseRead/spread
=== PAUSE TestPredicateFalseRead/spread
=== RUN   TestPredicateFalseRead/again
=== PAUSE TestPredicateFalseRead/again
=== CONT  TestPredicateFalseRead/no_read
=== CONT  TestPredicateFalseRead/again
=== CONT  TestPredicateFalseRead/spread
=== NAME  TestCheckedPredicateOracle/true
    predicates_checked_test.go:99: direction counts {Proven:0 Checked:1 Unobservable:1 Sites:[{Where:/tmp/adamic-gate/TestCheckedPredicateOracletrue726820244/001/true.ts:2:50 Function:isNumber Overload:0 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:unobservable Reason:string | number -> number}]}]}
=== RUN   TestCheckedPredicateOracle/true_unused
=== NAME  TestPredicateFalseRead/spread
    predicate_false_read_test.go:46: /workspace/adamic/internal/oracle/testdata/predicate_false_read/spread.ts:24:17: stage 0 can't lower an array of never yet
=== CONT  TestPredicateFalseRead/early_exit
=== NAME  TestPredicateFalseRead/again
    predicate_false_read_test.go:46: /workspace/adamic/internal/oracle/testdata/predicate_false_read/again.ts:11:19: stage 0 can't lower an array of unknown with erased element storage (retain its declared element type before reading elements) yet
=== CONT  TestPredicateFalseRead/else_read
--- PASS: TestCheckedPredicateCountsAreRecorded (48.21s)
=== CONT  TestPredicateFalseRead/false_read
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 231.68
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 227.06
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 miss 231.68
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 231.68
=== NAME  TestPredicateFalseRead/no_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 231.69
=== CONT  TestPredicateFalseRead/getter
=== CONT  TestPredicateFalseRead/renarrowed
=== CONT  TestPredicateFalseRead/capture
=== NAME  TestPredicateFalseRead/getter
    predicate_false_read_test.go:46: /workspace/adamic/internal/oracle/testdata/predicate_false_read/getter.ts:37:17: stage 0 can't lower a read narrowed to undefined of a property a call may write yet
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.02
    predicates_checked_test.go:164: true_unused lying body exits 70; removed-true-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
--- FAIL: TestPredicateFalseRead (0.00s)
    --- FAIL: TestPredicateFalseRead/spread (42.71s)
    --- FAIL: TestPredicateFalseRead/again (42.71s)
    --- PASS: TestPredicateFalseRead/early_exit (233.35s)
    --- PASS: TestPredicateFalseRead/false_read (227.86s)
    --- PASS: TestPredicateFalseRead/else_read (233.37s)
    --- PASS: TestPredicateFalseRead/no_read (276.19s)
    --- FAIL: TestPredicateFalseRead/getter (0.40s)
    --- PASS: TestPredicateFalseRead/renarrowed (1.30s)
    --- PASS: TestPredicateFalseRead/capture (1.32s)
=== NAME  TestCheckedPredicateOracle/false
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
=== RUN   TestCheckedPredicateOracle/false_alias
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
=== RUN   TestCheckedPredicateOracle/false_compound_unchanged
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
    predicates_checked_test.go:164: asserts lying body exits 70; removed-asserts-check mutant executes Node result "called\nasserted\n" and is caught
=== RUN   TestCheckedPredicateOracle/valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts_valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_lie
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
    predicates_checked_test.go:164: condition_lie lying body exits 70; removed-asserts-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/condition_once
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_once_lie
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
    predicates_checked_test.go:164: condition_once_lie lying body exits 70; removed-asserts-check mutant executes Node result "argument\ncontinued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_unchanged
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
=== RUN   TestCheckedPredicateOracle/assert_not_node
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
    predicates_checked_test.go:164: assert_not_node lying body exits 70; removed-asserts-check mutant executes Node result "asserting\nnarrowed\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_switch_true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.02
    predicates_checked_test.go:164: false_switch_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_eq_false
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.02
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
    predicates_checked_test.go:164: false_eq_false lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_ne_true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
    predicates_checked_test.go:164: false_ne_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_false_eq
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
    predicates_checked_test.go:164: false_false_eq lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_comma
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 126fe2368f11 hit 0.01
    predicates_checked_test.go:164: false_comma lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
--- FAIL: TestCheckedPredicateOracle (297.01s)
    --- FAIL: TestCheckedPredicateOracle/true (42.66s)
    --- PASS: TestCheckedPredicateOracle/true_unused (234.33s)
    --- PASS: TestCheckedPredicateOracle/false (0.93s)
    --- PASS: TestCheckedPredicateOracle/false_alias (0.82s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (1.02s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.02s)
    --- PASS: TestCheckedPredicateOracle/valid (1.06s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (0.87s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (1.05s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.13s)
    --- PASS: TestCheckedPredicateOracle/condition_once (1.20s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.13s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (1.03s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.21s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.39s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.99s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.62s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.25s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.28s)
FAIL
gate cache: native hits=0 misses=19
gate cache: node hits=0 misses=62
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	297.096s
FAIL
````

## Main comparison

Created `/tmp/adamic-main-39wtqdj` as a detached worktree at the exact base SHA, initialized submodules recursively, and copied only the requested test file and fixture directory from HEAD. Ran:

````sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestPredicateFalseRead' -count=1 -v -timeout 40m
````

`main-oracle exit=1 wall_seconds=279`

````text
--- FAIL: TestPredicateFalseRead (0.00s)
    --- FAIL: TestPredicateFalseRead/getter (7.49s)
    --- FAIL: TestPredicateFalseRead/no_read (7.49s)
    --- FAIL: TestPredicateFalseRead/spread (7.56s)
    --- FAIL: TestPredicateFalseRead/again (7.56s)
    --- PASS: TestPredicateFalseRead/false_read (30.77s)
    --- PASS: TestPredicateFalseRead/early_exit (30.77s)
    --- PASS: TestPredicateFalseRead/else_read (30.85s)
    --- PASS: TestPredicateFalseRead/renarrowed (38.35s)
    --- PASS: TestPredicateFalseRead/capture (30.87s)
````

````text
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
=== CONT  TestPredicateFalseRead
=== RUN   TestPredicateFalseRead/no_read
=== PAUSE TestPredicateFalseRead/no_read
=== RUN   TestPredicateFalseRead/false_read
=== PAUSE TestPredicateFalseRead/false_read
=== RUN   TestPredicateFalseRead/else_read
=== PAUSE TestPredicateFalseRead/else_read
=== RUN   TestPredicateFalseRead/early_exit
=== PAUSE TestPredicateFalseRead/early_exit
=== RUN   TestPredicateFalseRead/capture
=== PAUSE TestPredicateFalseRead/capture
=== RUN   TestPredicateFalseRead/renarrowed
=== PAUSE TestPredicateFalseRead/renarrowed
=== RUN   TestPredicateFalseRead/getter
=== PAUSE TestPredicateFalseRead/getter
=== RUN   TestPredicateFalseRead/spread
=== PAUSE TestPredicateFalseRead/spread
=== RUN   TestPredicateFalseRead/again
=== PAUSE TestPredicateFalseRead/again
=== CONT  TestPredicateFalseRead/no_read
=== CONT  TestPredicateFalseRead/renarrowed
=== CONT  TestPredicateFalseRead/again
=== CONT  TestPredicateFalseRead/spread
=== CONT  TestPredicateFalseRead/getter
    predicate_false_read_test.go:46: /tmp/adamic-main-39wtqdj/internal/oracle/testdata/predicate_false_read/getter.ts:37:17: stage 0 can't lower a read narrowed to undefined of a property a call may write yet
=== CONT  TestPredicateFalseRead/capture
=== NAME  TestPredicateFalseRead/no_read
    predicate_false_read_test.go:64: some's false direction at 32:7: "checked", want "unobservable", in [{Where:/tmp/adamic-main-39wtqdj/internal/oracle/testdata/predicate_false_read/no_read.ts:32:7 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] -> readonly string[]} {Direction:false Status:checked Reason:readonly string[] -> readonly string[]}]}]
=== CONT  TestPredicateFalseRead/else_read
=== NAME  TestPredicateFalseRead/spread
    predicate_false_read_test.go:46: /tmp/adamic-main-39wtqdj/internal/oracle/testdata/predicate_false_read/spread.ts:24:17: stage 0 can't lower an array of never yet
=== CONT  TestPredicateFalseRead/false_read
=== NAME  TestPredicateFalseRead/again
    predicate_false_read_test.go:46: /tmp/adamic-main-39wtqdj/internal/oracle/testdata/predicate_false_read/again.ts:11:19: stage 0 can't lower an array of unknown with erased element storage (retain its declared element type before reading elements) yet
=== CONT  TestPredicateFalseRead/early_exit
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4b9d09127c3e miss 29.16
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4b9d09127c3e hit 29.17
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4b9d09127c3e hit 29.17
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4b9d09127c3e hit 29.17
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4b9d09127c3e hit 29.17
--- FAIL: TestPredicateFalseRead (0.00s)
    --- FAIL: TestPredicateFalseRead/getter (7.49s)
    --- FAIL: TestPredicateFalseRead/no_read (7.49s)
    --- FAIL: TestPredicateFalseRead/spread (7.56s)
    --- FAIL: TestPredicateFalseRead/again (7.56s)
    --- PASS: TestPredicateFalseRead/false_read (30.77s)
    --- PASS: TestPredicateFalseRead/early_exit (30.77s)
    --- PASS: TestPredicateFalseRead/else_read (30.85s)
    --- PASS: TestPredicateFalseRead/renarrowed (38.35s)
    --- PASS: TestPredicateFalseRead/capture (30.87s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=14
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	38.387s
FAIL
````

The copied `no_read` test fails its metadata assertion before backend execution. A supplemental probe in that disposable main worktree measured all three backends directly without changing the copied test or compiler. It reports observations rather than asserting agreement; its PASS is not a claim that main agrees with Node. Probe source:

````text
package oracle
import (
 "testing"
 "path/filepath"
 "os"
 "fmt"
)
func TestProofMainNoReadRuntime(t *testing.T) {
 path, err := filepath.Abs(filepath.Join(repository,"internal/oracle/testdata/predicate_false_read/no_read.ts"))
 if err != nil { t.Fatal(err) }
 program, err := lowered(t,path)
 if err != nil { t.Fatal(err) }
 node := onNode(t,path)
 t.Logf("Node: exit %d, stdout %q, stderr %q",node.exitCode,node.stdout,node.stderr)
 sanitized,_ := nativelyUncached(t,program)
 for i, got := range []run{sanitized,releasedUncached(t,program),onJavaScriptBackend(t,program)} {
  name := []string{"sanitized native","release native","JavaScript"}[i]
  t.Logf("%s: exit %d, stdout %q, stderr %q",name,got.exitCode,got.stdout,got.stderr)
  if err := os.WriteFile(fmt.Sprintf("/tmp/adamic-proof/main-no-read-%d.stderr",i),got.stderr,0644); err != nil { t.Fatal(err) }
 }
}
````

`main-no-read-runtime exit=0 wall_seconds=31`

````text
=== RUN   TestProofMainNoReadRuntime
    proof_main_no_read_test.go:14: Node: exit 0, stdout "1.2.3\n1.2.3+build.5\n", stderr ""
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 4b9d09127c3e hit 0.84
    proof_main_no_read_test.go:18: sanitized native: exit 70, stdout "", stderr "adamic: panic: predicate some at /tmp/adamic-main-39wtqdj/internal/oracle/testdata/predicate_false_read/no_read.ts:32:7 false branch: source readonly string[], target readonly string[]\n"
    proof_main_no_read_test.go:18: release native: exit 70, stdout "", stderr "adamic: panic: predicate some at /tmp/adamic-main-39wtqdj/internal/oracle/testdata/predicate_false_read/no_read.ts:32:7 false branch: source readonly string[], target readonly string[]\n"
    proof_main_no_read_test.go:18: JavaScript: exit 70, stdout "", stderr "adamic: panic: predicate some at /tmp/adamic-main-39wtqdj/internal/oracle/testdata/predicate_false_read/no_read.ts:32:7 false branch: source readonly string[], target readonly string[]\n"
--- PASS: TestProofMainNoReadRuntime (8.64s)
PASS
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=2
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	8.666s
````

Exact stderr, identical on sanitized native, release native, and JavaScript (one trailing newline):

````text
adamic: panic: predicate some at /tmp/adamic-main-39wtqdj/internal/oracle/testdata/predicate_false_read/no_read.ts:32:7 false branch: source readonly string[], target readonly string[]
````

The temporary main worktree, copied fixtures, and supplemental test were removed before mutation testing.

## Mutants

Each old text match count was asserted exactly one before applying the edit. Both pinned commands above were rerun for each mutant, sequentially. Each mutant was restored before the next; every restoration passed `git diff --exit-code`. A red present at HEAD is retained and is not attributed to the mutant.

### M1

````text
M1: old text occurs exactly once; applied
````

Inserted the fenced Go block under “The condition-only arm” in `internal/lower/predicate_false_read.mutants.md` immediately before the unique final `return directions` followed by `}` and `// predicateFalseRead is`.
````go
	conditionNarrows := func(condition *ast.Node, truth bool) bool {
		if condition == nil || l.predicateConditionDirection(condition, call.AsNode(), truth, 0)&predicateFalse == 0 {
			return false
		}
		flags := ast.FlowFlagsFalseCondition
		if truth {
			flags = ast.FlowFlagsTrueCondition
		}
		branch := &ast.FlowNode{Flags: flags, Node: condition, Antecedent: &ast.FlowNode{Flags: ast.FlowFlagsStart}}
		narrowed := predicateFlowType(l.checker, argument, source, source, nil, branch)
		return !checker.Checker_isTypeIdenticalTo(l.checker, source, narrowed)
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		var condition *ast.Node
		switch node.Kind {
		case ast.KindIfStatement:
			condition = node.AsIfStatement().Expression
		case ast.KindConditionalExpression:
			condition = node.AsConditionalExpression().Condition
		case ast.KindWhileStatement:
			condition = node.AsWhileStatement().Expression
		case ast.KindDoStatement:
			condition = node.AsDoStatement().Expression
		case ast.KindForStatement:
			condition = node.AsForStatement().Condition
		case ast.KindBinaryExpression:
			binary := node.AsBinaryExpression()
			if binary.OperatorToken.Kind == ast.KindBarBarToken && conditionNarrows(binary.Left, false) {
				directions |= predicateFalse
			}
		case ast.KindSwitchStatement:
			statement := node.AsSwitchStatement()
			if ast.SkipParentheses(statement.Expression).Kind == ast.KindTrueKeyword {
				for _, clause := range statement.CaseBlock.AsCaseBlock().Clauses.Nodes {
					if clause.Kind == ast.KindCaseClause && conditionNarrows(clause.Expression(), false) {
						directions |= predicateFalse
					}
				}
			}
		}
		if conditionNarrows(condition, true) || conditionNarrows(condition, false) {
			directions |= predicateFalse
		}
		node.ForEachChild(visit)
		return false
	}
	ast.GetSourceFileOfNode(call.AsNode()).AsNode().ForEachChild(visit)
````


`M1-lower exit=1 wall_seconds=37`

````text
--- PASS: TestPredicateUsesBelongToEachCall (0.23s)
--- FAIL: TestCheckedPredicateFalseReadDirections (0.02s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.17s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/empty_branches (0.24s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.15s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.26s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.23s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.24s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.11s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.09s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.09s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/semver's_toString (0.12s)
--- PASS: TestCheckedPredicatePendingContracts (1.07s)
    --- PASS: TestCheckedPredicatePendingContracts/untagged (0.24s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag (0.22s)
    --- PASS: TestCheckedPredicatePendingContracts/overload_in_.a (0.14s)
    --- PASS: TestCheckedPredicatePendingContracts/loose_false_comparison (0.11s)
    --- PASS: TestCheckedPredicatePendingContracts/open_enum_union_tag (0.05s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_read (0.10s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_pass (0.09s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_domain (0.10s)
--- PASS: TestPredicateUseRegions (0.05s)
    --- PASS: TestPredicateUseRegions/unused (0.09s)
    --- PASS: TestPredicateUseRegions/indexed_argument (0.10s)
    --- PASS: TestPredicateUseRegions/comma_continuation (0.07s)
    --- PASS: TestPredicateUseRegions/switch_true_default (0.06s)
    --- PASS: TestPredicateUseRegions/reversed_false_comparison (0.08s)
    --- PASS: TestPredicateUseRegions/true_inequality (0.11s)
    --- PASS: TestPredicateUseRegions/loose_false_comparison (0.11s)
    --- PASS: TestPredicateUseRegions/false_comparison (0.11s)
    --- PASS: TestPredicateUseRegions/replacement (0.13s)
    --- PASS: TestPredicateUseRegions/later_read (0.08s)
    --- PASS: TestPredicateUseRegions/loop_read (0.09s)
    --- PASS: TestPredicateUseRegions/false_continuation (0.05s)
    --- PASS: TestPredicateUseRegions/computed_index (0.10s)
    --- PASS: TestPredicateUseRegions/indexed_spelling (0.09s)
    --- PASS: TestPredicateUseRegions/closure_read (0.09s)
    --- PASS: TestPredicateUseRegions/compound_result_alias (0.07s)
    --- PASS: TestPredicateUseRegions/short_circuit (0.10s)
    --- PASS: TestPredicateUseRegions/const_result_alias (0.07s)
    --- PASS: TestPredicateUseRegions/both (0.09s)
    --- PASS: TestPredicateUseRegions/throw_continuation (0.09s)
    --- PASS: TestPredicateUseRegions/return_continuation (0.09s)
    --- PASS: TestPredicateUseRegions/false (0.09s)
    --- PASS: TestPredicateUseRegions/true (0.06s)
--- FAIL: TestCheckedPredicateBranches (5.52s)
    --- PASS: TestCheckedPredicateBranches/true (3.26s)
    --- FAIL: TestCheckedPredicateBranches/false (0.46s)
    --- PASS: TestCheckedPredicateBranches/false_read (0.68s)
    --- PASS: TestCheckedPredicateBranches/asserts (0.64s)
    --- PASS: TestCheckedPredicateBranches/valid (0.48s)
````

Complete output, preserving every red subtest’s whole diagnostic and any recorded/measured rows verbatim:

````text
=== RUN   TestCheckedPredicateBranches
=== PAUSE TestCheckedPredicateBranches
=== RUN   TestCheckedPredicatePendingContracts
=== PAUSE TestCheckedPredicatePendingContracts
=== RUN   TestCheckedPredicateFalseReadDirections
=== PAUSE TestCheckedPredicateFalseReadDirections
=== RUN   TestPredicateUseRegions
=== PAUSE TestPredicateUseRegions
=== RUN   TestPredicateUsesBelongToEachCall
=== PAUSE TestPredicateUsesBelongToEachCall
=== CONT  TestCheckedPredicateBranches
=== CONT  TestPredicateUseRegions
=== CONT  TestPredicateUsesBelongToEachCall
=== RUN   TestCheckedPredicateBranches/true
=== CONT  TestCheckedPredicateFalseReadDirections
=== RUN   TestCheckedPredicateFalseReadDirections/empty_branches
=== CONT  TestCheckedPredicatePendingContracts
=== PAUSE TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestPredicateUseRegions/unused
=== PAUSE TestCheckedPredicateFalseReadDirections/semver's_toString
=== PAUSE TestPredicateUseRegions/unused
=== RUN   TestCheckedPredicatePendingContracts/untagged
=== RUN   TestPredicateUseRegions/true
=== PAUSE TestPredicateUseRegions/true
=== RUN   TestPredicateUseRegions/false
=== PAUSE TestPredicateUseRegions/false
=== RUN   TestPredicateUseRegions/both
=== PAUSE TestPredicateUseRegions/both
=== RUN   TestPredicateUseRegions/return_continuation
=== PAUSE TestPredicateUseRegions/return_continuation
=== RUN   TestPredicateUseRegions/throw_continuation
=== PAUSE TestPredicateUseRegions/throw_continuation
=== RUN   TestPredicateUseRegions/false_continuation
=== PAUSE TestPredicateUseRegions/false_continuation
=== RUN   TestPredicateUseRegions/const_result_alias
=== RUN   TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== PAUSE TestPredicateUseRegions/const_result_alias
=== PAUSE TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== RUN   TestPredicateUseRegions/compound_result_alias
=== PAUSE TestPredicateUseRegions/compound_result_alias
=== RUN   TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== PAUSE TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== RUN   TestPredicateUseRegions/short_circuit
=== PAUSE TestPredicateUseRegions/short_circuit
=== RUN   TestCheckedPredicateFalseReadDirections/false_branch
=== PAUSE TestCheckedPredicateFalseReadDirections/false_branch
=== RUN   TestCheckedPredicateFalseReadDirections/else
=== PAUSE TestCheckedPredicateFalseReadDirections/else
=== RUN   TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== PAUSE TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== RUN   TestCheckedPredicateFalseReadDirections/capture
=== PAUSE TestCheckedPredicateFalseReadDirections/capture
=== RUN   TestCheckedPredicateFalseReadDirections/renarrowed
=== PAUSE TestCheckedPredicateFalseReadDirections/renarrowed
=== RUN   TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== PAUSE TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== RUN   TestCheckedPredicateFalseReadDirections/getter
=== PAUSE TestCheckedPredicateFalseReadDirections/getter
=== RUN   TestCheckedPredicateFalseReadDirections/spread
=== PAUSE TestCheckedPredicateFalseReadDirections/spread
=== RUN   TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== PAUSE TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== RUN   TestCheckedPredicateFalseReadDirections/conditional
=== PAUSE TestCheckedPredicateFalseReadDirections/conditional
=== RUN   TestCheckedPredicateFalseReadDirections/||_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/||_operand
=== RUN   TestCheckedPredicateFalseReadDirections/&&_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/&&_operand
=== RUN   TestCheckedPredicateFalseReadDirections/const_alias
=== PAUSE TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestPredicateUseRegions/closure_read
=== PAUSE TestPredicateUseRegions/closure_read
=== RUN   TestPredicateUseRegions/indexed_spelling
=== PAUSE TestPredicateUseRegions/indexed_spelling
=== RUN   TestPredicateUseRegions/indexed_argument
=== PAUSE TestPredicateUseRegions/indexed_argument
=== RUN   TestPredicateUseRegions/computed_index
=== PAUSE TestPredicateUseRegions/computed_index
=== RUN   TestPredicateUseRegions/loop_read
=== PAUSE TestPredicateUseRegions/loop_read
=== RUN   TestPredicateUseRegions/later_read
=== PAUSE TestPredicateUseRegions/later_read
=== RUN   TestPredicateUseRegions/replacement
=== PAUSE TestPredicateUseRegions/replacement
=== RUN   TestPredicateUseRegions/false_comparison
=== PAUSE TestPredicateUseRegions/false_comparison
=== RUN   TestPredicateUseRegions/loose_false_comparison
=== PAUSE TestPredicateUseRegions/loose_false_comparison
=== RUN   TestPredicateUseRegions/true_inequality
=== PAUSE TestPredicateUseRegions/true_inequality
=== RUN   TestPredicateUseRegions/reversed_false_comparison
=== PAUSE TestPredicateUseRegions/reversed_false_comparison
=== RUN   TestPredicateUseRegions/switch_true_default
=== PAUSE TestPredicateUseRegions/switch_true_default
=== RUN   TestPredicateUseRegions/comma_continuation
=== PAUSE TestPredicateUseRegions/comma_continuation
=== CONT  TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== CONT  TestCheckedPredicateFalseReadDirections/const_alias
--- PASS: TestPredicateUsesBelongToEachCall (0.23s)
=== CONT  TestCheckedPredicateFalseReadDirections/&&_operand
=== RUN   TestCheckedPredicatePendingContracts/open_tag
=== NAME  TestCheckedPredicateFalseReadDirections/empty_branches
    predicates_false_read_test.go:89: false side checked true, want false
=== CONT  TestCheckedPredicateFalseReadDirections/||_operand
=== CONT  TestCheckedPredicateFalseReadDirections/conditional
=== RUN   TestCheckedPredicatePendingContracts/overload_in_.a
=== CONT  TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== CONT  TestCheckedPredicateFalseReadDirections/spread
=== CONT  TestCheckedPredicateFalseReadDirections/getter
=== RUN   TestCheckedPredicatePendingContracts/loose_false_comparison
=== CONT  TestCheckedPredicateFalseReadDirections/else
=== CONT  TestCheckedPredicateFalseReadDirections/renarrowed
=== CONT  TestCheckedPredicateFalseReadDirections/capture
=== CONT  TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== RUN   TestCheckedPredicatePendingContracts/open_enum_union_tag
=== CONT  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_read
=== CONT  TestCheckedPredicateFalseReadDirections/false_branch
=== CONT  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== NAME  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
    predicates_false_read_test.go:89: false side checked true, want false
=== CONT  TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_pass
=== CONT  TestPredicateUseRegions/unused
=== NAME  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
    predicates_false_read_test.go:89: false side checked true, want false
=== CONT  TestPredicateUseRegions/indexed_argument
=== NAME  TestCheckedPredicateFalseReadDirections/semver's_toString
    predicates_false_read_test.go:89: false side checked true, want false
--- FAIL: TestCheckedPredicateFalseReadDirections (0.02s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.17s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/empty_branches (0.24s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.15s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.26s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.23s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.24s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.11s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.09s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.09s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/semver's_toString (0.12s)
=== CONT  TestPredicateUseRegions/comma_continuation
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_domain
=== CONT  TestPredicateUseRegions/switch_true_default
=== CONT  TestPredicateUseRegions/reversed_false_comparison
=== CONT  TestPredicateUseRegions/true_inequality
=== CONT  TestPredicateUseRegions/loose_false_comparison
--- PASS: TestCheckedPredicatePendingContracts (1.07s)
    --- PASS: TestCheckedPredicatePendingContracts/untagged (0.24s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag (0.22s)
    --- PASS: TestCheckedPredicatePendingContracts/overload_in_.a (0.14s)
    --- PASS: TestCheckedPredicatePendingContracts/loose_false_comparison (0.11s)
    --- PASS: TestCheckedPredicatePendingContracts/open_enum_union_tag (0.05s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_read (0.10s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_pass (0.09s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_domain (0.10s)
=== CONT  TestPredicateUseRegions/false_comparison
=== CONT  TestPredicateUseRegions/replacement
=== CONT  TestPredicateUseRegions/later_read
=== CONT  TestPredicateUseRegions/loop_read
=== CONT  TestPredicateUseRegions/computed_index
=== CONT  TestPredicateUseRegions/false_continuation
=== CONT  TestPredicateUseRegions/indexed_spelling
=== CONT  TestPredicateUseRegions/closure_read
=== CONT  TestPredicateUseRegions/short_circuit
=== CONT  TestPredicateUseRegions/compound_result_alias
=== CONT  TestPredicateUseRegions/const_result_alias
=== CONT  TestPredicateUseRegions/both
=== CONT  TestPredicateUseRegions/throw_continuation
=== CONT  TestPredicateUseRegions/return_continuation
=== CONT  TestPredicateUseRegions/false
=== CONT  TestPredicateUseRegions/true
--- PASS: TestPredicateUseRegions (0.05s)
    --- PASS: TestPredicateUseRegions/unused (0.09s)
    --- PASS: TestPredicateUseRegions/indexed_argument (0.10s)
    --- PASS: TestPredicateUseRegions/comma_continuation (0.07s)
    --- PASS: TestPredicateUseRegions/switch_true_default (0.06s)
    --- PASS: TestPredicateUseRegions/reversed_false_comparison (0.08s)
    --- PASS: TestPredicateUseRegions/true_inequality (0.11s)
    --- PASS: TestPredicateUseRegions/loose_false_comparison (0.11s)
    --- PASS: TestPredicateUseRegions/false_comparison (0.11s)
    --- PASS: TestPredicateUseRegions/replacement (0.13s)
    --- PASS: TestPredicateUseRegions/later_read (0.08s)
    --- PASS: TestPredicateUseRegions/loop_read (0.09s)
    --- PASS: TestPredicateUseRegions/false_continuation (0.05s)
    --- PASS: TestPredicateUseRegions/computed_index (0.10s)
    --- PASS: TestPredicateUseRegions/indexed_spelling (0.09s)
    --- PASS: TestPredicateUseRegions/closure_read (0.09s)
    --- PASS: TestPredicateUseRegions/compound_result_alias (0.07s)
    --- PASS: TestPredicateUseRegions/short_circuit (0.10s)
    --- PASS: TestPredicateUseRegions/const_result_alias (0.07s)
    --- PASS: TestPredicateUseRegions/both (0.09s)
    --- PASS: TestPredicateUseRegions/throw_continuation (0.09s)
    --- PASS: TestPredicateUseRegions/return_continuation (0.09s)
    --- PASS: TestPredicateUseRegions/false (0.09s)
    --- PASS: TestPredicateUseRegions/true (0.06s)
=== RUN   TestCheckedPredicateBranches/false
    predicates_checked_test.go:90: /tmp/adamic-gate/TestCheckedPredicateBranchesfalse4163730983/002/native: exit 70 stdout "" stderr "adamic: panic: predicate isNumber at /tmp/adamic-gate/TestCheckedPredicateBranchesfalse4163730983/001/main.ts:2:50 false branch: source string | number, target number\n"; want 0 "continued\n" ""
=== RUN   TestCheckedPredicateBranches/false_read
=== RUN   TestCheckedPredicateBranches/asserts
=== RUN   TestCheckedPredicateBranches/valid
--- FAIL: TestCheckedPredicateBranches (5.52s)
    --- PASS: TestCheckedPredicateBranches/true (3.26s)
    --- FAIL: TestCheckedPredicateBranches/false (0.46s)
    --- PASS: TestCheckedPredicateBranches/false_read (0.68s)
    --- PASS: TestCheckedPredicateBranches/asserts (0.64s)
    --- PASS: TestCheckedPredicateBranches/valid (0.48s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	5.705s
FAIL
````

`M1-oracle exit=1 wall_seconds=92`

````text
--- FAIL: TestCheckedPredicateCountsAreRecorded (14.59s)
--- FAIL: TestPredicateFalseRead (0.00s)
    --- FAIL: TestPredicateFalseRead/again (7.68s)
    --- FAIL: TestPredicateFalseRead/spread (7.74s)
    --- FAIL: TestPredicateFalseRead/no_read (7.88s)
    --- FAIL: TestPredicateFalseRead/getter (0.26s)
    --- PASS: TestPredicateFalseRead/early_exit (34.33s)
    --- PASS: TestPredicateFalseRead/else_read (34.28s)
    --- PASS: TestPredicateFalseRead/false_read (34.15s)
    --- PASS: TestPredicateFalseRead/renarrowed (27.24s)
    --- PASS: TestPredicateFalseRead/capture (1.06s)
--- FAIL: TestCheckedPredicateOracle (62.28s)
    --- PASS: TestCheckedPredicateOracle/true (42.68s)
    --- PASS: TestCheckedPredicateOracle/true_unused (1.12s)
    --- FAIL: TestCheckedPredicateOracle/false (0.19s)
    --- FAIL: TestCheckedPredicateOracle/false_alias (0.17s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (1.07s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.05s)
    --- PASS: TestCheckedPredicateOracle/valid (1.18s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (0.89s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (1.15s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.35s)
    --- PASS: TestCheckedPredicateOracle/condition_once (1.11s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.31s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (0.88s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.34s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.17s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.33s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.58s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.35s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.35s)
````

Complete output, preserving every red subtest’s whole diagnostic and any recorded/measured rows verbatim:

````text
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
=== RUN   TestCheckedPredicateOracle
=== PAUSE TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateCountsAreRecorded
=== PAUSE TestCheckedPredicateCountsAreRecorded
=== CONT  TestPredicateFalseRead
=== CONT  TestCheckedPredicateOracle
=== CONT  TestCheckedPredicateCountsAreRecorded
=== RUN   TestCheckedPredicateOracle/true
=== RUN   TestPredicateFalseRead/no_read
=== PAUSE TestPredicateFalseRead/no_read
=== RUN   TestPredicateFalseRead/false_read
=== PAUSE TestPredicateFalseRead/false_read
=== RUN   TestPredicateFalseRead/else_read
=== PAUSE TestPredicateFalseRead/else_read
=== RUN   TestPredicateFalseRead/early_exit
=== PAUSE TestPredicateFalseRead/early_exit
=== RUN   TestPredicateFalseRead/capture
=== PAUSE TestPredicateFalseRead/capture
=== RUN   TestPredicateFalseRead/renarrowed
=== PAUSE TestPredicateFalseRead/renarrowed
=== RUN   TestPredicateFalseRead/getter
=== PAUSE TestPredicateFalseRead/getter
=== RUN   TestPredicateFalseRead/spread
=== PAUSE TestPredicateFalseRead/spread
=== RUN   TestPredicateFalseRead/again
=== PAUSE TestPredicateFalseRead/again
=== CONT  TestPredicateFalseRead/no_read
=== CONT  TestPredicateFalseRead/again
=== CONT  TestPredicateFalseRead/spread
=== NAME  TestPredicateFalseRead/again
    predicate_false_read_test.go:46: /workspace/adamic/internal/oracle/testdata/predicate_false_read/again.ts:11:19: stage 0 can't lower an array of unknown with erased element storage (retain its declared element type before reading elements) yet
=== CONT  TestPredicateFalseRead/early_exit
=== NAME  TestPredicateFalseRead/spread
    predicate_false_read_test.go:46: /workspace/adamic/internal/oracle/testdata/predicate_false_read/spread.ts:24:17: stage 0 can't lower an array of never yet
=== CONT  TestPredicateFalseRead/else_read
=== NAME  TestPredicateFalseRead/no_read
    predicate_false_read_test.go:64: some's false direction at 32:7: "checked", want "unobservable", in [{Where:/workspace/adamic/internal/oracle/testdata/predicate_false_read/no_read.ts:32:7 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] -> readonly string[]} {Direction:false Status:checked Reason:readonly string[] -> readonly string[]}]}]
=== CONT  TestPredicateFalseRead/false_read
=== NAME  TestCheckedPredicateCountsAreRecorded
    predicates_checked_test.go:278: checked predicate counts changed in testdata/counts/runtime/internal/lower/testdata/predicates_checked/false.a (.ts checked mode).counts; update with -update-counts:
        recorded: | internal/lower/testdata/predicates_checked/false.a (.ts checked mode) | 1 | 1 | 0 | 1 | 1 | 0 |
        measured: | internal/lower/testdata/predicates_checked/false.a (.ts checked mode) | 1 | 0 | 0 | 0 | 1 | 0 |
    predicates_checked_test.go:278: checked predicate counts changed in testdata/counts/runtime/internal/lower/testdata/predicates_checked/false_alias.a (.ts checked mode).counts; update with -update-counts:
        recorded: | internal/lower/testdata/predicates_checked/false_alias.a (.ts checked mode) | 1 | 1 | 0 | 1 | 1 | 0 |
        measured: | internal/lower/testdata/predicates_checked/false_alias.a (.ts checked mode) | 1 | 0 | 0 | 0 | 1 | 0 |
=== CONT  TestPredicateFalseRead/getter
--- FAIL: TestCheckedPredicateCountsAreRecorded (14.59s)
=== NAME  TestPredicateFalseRead/getter
    predicate_false_read_test.go:46: /workspace/adamic/internal/oracle/testdata/predicate_false_read/getter.ts:37:17: stage 0 can't lower a read narrowed to undefined of a property a call may write yet
=== CONT  TestPredicateFalseRead/renarrowed
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b miss 32.72
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 32.80
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 32.80
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 32.80
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 26.32
=== CONT  TestPredicateFalseRead/capture
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.02
    predicates_checked_test.go:164: true lying body exits 70; removed-true-check mutant executes Node result "claimed\n" and is caught
=== RUN   TestCheckedPredicateOracle/true_unused
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
--- FAIL: TestPredicateFalseRead (0.00s)
    --- FAIL: TestPredicateFalseRead/again (7.68s)
    --- FAIL: TestPredicateFalseRead/spread (7.74s)
    --- FAIL: TestPredicateFalseRead/no_read (7.88s)
    --- FAIL: TestPredicateFalseRead/getter (0.26s)
    --- PASS: TestPredicateFalseRead/early_exit (34.33s)
    --- PASS: TestPredicateFalseRead/else_read (34.28s)
    --- PASS: TestPredicateFalseRead/false_read (34.15s)
    --- PASS: TestPredicateFalseRead/renarrowed (27.24s)
    --- PASS: TestPredicateFalseRead/capture (1.06s)
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
    predicates_checked_test.go:164: true_unused lying body exits 70; removed-true-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false
    predicates_checked_test.go:99: direction counts {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestCheckedPredicateOraclefalse2802724372/001/false.ts:2:50 Function:isNumber Overload:0 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:checked Reason:string | number -> number}]}]}
=== RUN   TestCheckedPredicateOracle/false_alias
    predicates_checked_test.go:99: direction counts {Proven:0 Checked:2 Unobservable:0 Sites:[{Where:/tmp/adamic-gate/TestCheckedPredicateOraclefalse_alias3913705337/001/false_alias.ts:2:61 Function:isNumber Overload:0 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:checked Reason:string | number -> number}]}]}
=== RUN   TestCheckedPredicateOracle/false_compound_unchanged
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
    predicates_checked_test.go:164: asserts lying body exits 70; removed-asserts-check mutant executes Node result "called\nasserted\n" and is caught
=== RUN   TestCheckedPredicateOracle/valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.02
=== RUN   TestCheckedPredicateOracle/asserts_valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_lie
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.02
    predicates_checked_test.go:164: condition_lie lying body exits 70; removed-asserts-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/condition_once
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_once_lie
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
    predicates_checked_test.go:164: condition_once_lie lying body exits 70; removed-asserts-check mutant executes Node result "argument\ncontinued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_unchanged
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
=== RUN   TestCheckedPredicateOracle/assert_not_node
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.02
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
    predicates_checked_test.go:164: assert_not_node lying body exits 70; removed-asserts-check mutant executes Node result "asserting\nnarrowed\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_switch_true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
    predicates_checked_test.go:164: false_switch_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_eq_false
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
    predicates_checked_test.go:164: false_eq_false lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_ne_true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
    predicates_checked_test.go:164: false_ne_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_false_eq
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
    predicates_checked_test.go:164: false_false_eq lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_comma
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types ea96562e7e0b hit 0.01
    predicates_checked_test.go:164: false_comma lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
--- FAIL: TestCheckedPredicateOracle (62.28s)
    --- PASS: TestCheckedPredicateOracle/true (42.68s)
    --- PASS: TestCheckedPredicateOracle/true_unused (1.12s)
    --- FAIL: TestCheckedPredicateOracle/false (0.19s)
    --- FAIL: TestCheckedPredicateOracle/false_alias (0.17s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (1.07s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.05s)
    --- PASS: TestCheckedPredicateOracle/valid (1.18s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (0.89s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (1.15s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.35s)
    --- PASS: TestCheckedPredicateOracle/condition_once (1.11s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.31s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (0.88s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.34s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.17s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.33s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.58s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.35s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.35s)
FAIL
gate cache: native hits=0 misses=19
gate cache: node hits=0 misses=61
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	62.383s
FAIL
````

````text
M1 restored; git diff --exit-code: 0
````

### M2

````text
M2: old text occurs exactly once; applied
````

````go
if l.predicateUseDirections(call)&predicateFalse != 0 || l.predicateFalseRead(call, argument) {
→
if l.predicateUseDirections(call)&predicateFalse != 0 {
````


`M2-lower exit=1 wall_seconds=35`

````text
--- PASS: TestPredicateUsesBelongToEachCall (0.20s)
--- FAIL: TestCheckedPredicateFalseReadDirections (0.01s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/empty_branches (0.21s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.19s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.17s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.16s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.13s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.13s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/semver's_toString (0.10s)
--- PASS: TestCheckedPredicatePendingContracts (0.90s)
    --- PASS: TestCheckedPredicatePendingContracts/untagged (0.18s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag (0.19s)
    --- PASS: TestCheckedPredicatePendingContracts/overload_in_.a (0.10s)
    --- PASS: TestCheckedPredicatePendingContracts/loose_false_comparison (0.11s)
    --- PASS: TestCheckedPredicatePendingContracts/open_enum_union_tag (0.06s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_read (0.07s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_pass (0.10s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_domain (0.10s)
--- PASS: TestPredicateUseRegions (0.01s)
    --- PASS: TestPredicateUseRegions/unused (0.07s)
    --- PASS: TestPredicateUseRegions/indexed_argument (0.09s)
    --- PASS: TestPredicateUseRegions/indexed_spelling (0.09s)
    --- PASS: TestPredicateUseRegions/closure_read (0.09s)
    --- PASS: TestPredicateUseRegions/short_circuit (0.08s)
    --- PASS: TestPredicateUseRegions/false_continuation (0.08s)
    --- PASS: TestPredicateUseRegions/const_result_alias (0.09s)
    --- PASS: TestPredicateUseRegions/compound_result_alias (0.09s)
    --- PASS: TestPredicateUseRegions/throw_continuation (0.09s)
    --- PASS: TestPredicateUseRegions/false (0.04s)
    --- PASS: TestPredicateUseRegions/both (0.10s)
    --- PASS: TestPredicateUseRegions/return_continuation (0.12s)
    --- PASS: TestPredicateUseRegions/true (0.09s)
    --- PASS: TestPredicateUseRegions/loose_false_comparison (0.12s)
    --- PASS: TestPredicateUseRegions/comma_continuation (0.12s)
    --- PASS: TestPredicateUseRegions/later_read (0.17s)
    --- PASS: TestPredicateUseRegions/switch_true_default (0.19s)
    --- PASS: TestPredicateUseRegions/false_comparison (0.15s)
    --- PASS: TestPredicateUseRegions/replacement (0.11s)
    --- PASS: TestPredicateUseRegions/reversed_false_comparison (0.08s)
    --- PASS: TestPredicateUseRegions/loop_read (0.09s)
    --- PASS: TestPredicateUseRegions/true_inequality (0.06s)
    --- PASS: TestPredicateUseRegions/computed_index (0.09s)
--- PASS: TestCheckedPredicateBranches (5.85s)
    --- PASS: TestCheckedPredicateBranches/true (3.37s)
    --- PASS: TestCheckedPredicateBranches/false (0.78s)
    --- PASS: TestCheckedPredicateBranches/false_read (0.55s)
    --- PASS: TestCheckedPredicateBranches/asserts (0.53s)
    --- PASS: TestCheckedPredicateBranches/valid (0.62s)
````

Complete output, preserving every red subtest’s whole diagnostic and any recorded/measured rows verbatim:

````text
=== RUN   TestCheckedPredicateBranches
=== PAUSE TestCheckedPredicateBranches
=== RUN   TestCheckedPredicatePendingContracts
=== PAUSE TestCheckedPredicatePendingContracts
=== RUN   TestCheckedPredicateFalseReadDirections
=== PAUSE TestCheckedPredicateFalseReadDirections
=== RUN   TestPredicateUseRegions
=== PAUSE TestPredicateUseRegions
=== RUN   TestPredicateUsesBelongToEachCall
=== PAUSE TestPredicateUsesBelongToEachCall
=== CONT  TestCheckedPredicateBranches
=== CONT  TestPredicateUseRegions
=== CONT  TestPredicateUsesBelongToEachCall
=== CONT  TestCheckedPredicatePendingContracts
=== CONT  TestCheckedPredicateFalseReadDirections
=== RUN   TestCheckedPredicateBranches/true
=== RUN   TestPredicateUseRegions/unused
=== PAUSE TestPredicateUseRegions/unused
=== RUN   TestCheckedPredicateFalseReadDirections/empty_branches
=== PAUSE TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestCheckedPredicatePendingContracts/untagged
=== RUN   TestPredicateUseRegions/true
=== PAUSE TestPredicateUseRegions/true
=== RUN   TestPredicateUseRegions/false
=== PAUSE TestPredicateUseRegions/false
=== RUN   TestPredicateUseRegions/both
=== PAUSE TestPredicateUseRegions/both
=== RUN   TestPredicateUseRegions/return_continuation
=== PAUSE TestPredicateUseRegions/return_continuation
=== RUN   TestPredicateUseRegions/throw_continuation
=== PAUSE TestPredicateUseRegions/throw_continuation
=== RUN   TestPredicateUseRegions/false_continuation
=== PAUSE TestPredicateUseRegions/false_continuation
=== RUN   TestPredicateUseRegions/const_result_alias
=== PAUSE TestPredicateUseRegions/const_result_alias
=== RUN   TestPredicateUseRegions/compound_result_alias
=== PAUSE TestPredicateUseRegions/compound_result_alias
=== RUN   TestPredicateUseRegions/short_circuit
=== PAUSE TestPredicateUseRegions/short_circuit
=== RUN   TestPredicateUseRegions/closure_read
=== PAUSE TestPredicateUseRegions/closure_read
=== RUN   TestPredicateUseRegions/indexed_spelling
=== PAUSE TestPredicateUseRegions/indexed_spelling
=== RUN   TestPredicateUseRegions/indexed_argument
=== PAUSE TestPredicateUseRegions/indexed_argument
=== RUN   TestPredicateUseRegions/computed_index
=== PAUSE TestPredicateUseRegions/computed_index
=== RUN   TestPredicateUseRegions/loop_read
=== PAUSE TestPredicateUseRegions/loop_read
=== RUN   TestPredicateUseRegions/later_read
=== PAUSE TestPredicateUseRegions/later_read
=== RUN   TestPredicateUseRegions/replacement
=== PAUSE TestPredicateUseRegions/replacement
=== RUN   TestPredicateUseRegions/false_comparison
=== PAUSE TestPredicateUseRegions/false_comparison
=== RUN   TestCheckedPredicateFalseReadDirections/semver's_toString
=== PAUSE TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestPredicateUseRegions/loose_false_comparison
=== PAUSE TestPredicateUseRegions/loose_false_comparison
=== RUN   TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== PAUSE TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== RUN   TestPredicateUseRegions/true_inequality
=== PAUSE TestPredicateUseRegions/true_inequality
=== RUN   TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== PAUSE TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== RUN   TestCheckedPredicateFalseReadDirections/false_branch
=== PAUSE TestCheckedPredicateFalseReadDirections/false_branch
=== RUN   TestCheckedPredicateFalseReadDirections/else
=== PAUSE TestCheckedPredicateFalseReadDirections/else
=== RUN   TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== PAUSE TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== RUN   TestCheckedPredicateFalseReadDirections/capture
=== RUN   TestPredicateUseRegions/reversed_false_comparison
=== PAUSE TestCheckedPredicateFalseReadDirections/capture
=== RUN   TestCheckedPredicateFalseReadDirections/renarrowed
=== PAUSE TestCheckedPredicateFalseReadDirections/renarrowed
=== RUN   TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== PAUSE TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== RUN   TestCheckedPredicateFalseReadDirections/getter
=== PAUSE TestCheckedPredicateFalseReadDirections/getter
=== RUN   TestCheckedPredicateFalseReadDirections/spread
=== PAUSE TestCheckedPredicateFalseReadDirections/spread
=== RUN   TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== PAUSE TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== RUN   TestCheckedPredicateFalseReadDirections/conditional
=== PAUSE TestCheckedPredicateFalseReadDirections/conditional
=== RUN   TestCheckedPredicateFalseReadDirections/||_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/||_operand
=== RUN   TestCheckedPredicateFalseReadDirections/&&_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/&&_operand
=== RUN   TestCheckedPredicateFalseReadDirections/const_alias
=== PAUSE TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestCheckedPredicateFalseReadDirections/empty_branches
=== PAUSE TestPredicateUseRegions/reversed_false_comparison
=== RUN   TestPredicateUseRegions/switch_true_default
=== PAUSE TestPredicateUseRegions/switch_true_default
=== RUN   TestPredicateUseRegions/comma_continuation
=== PAUSE TestPredicateUseRegions/comma_continuation
=== CONT  TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
    predicates_false_read_test.go:89: false side checked false, want true
=== CONT  TestCheckedPredicateFalseReadDirections/const_alias
=== RUN   TestCheckedPredicatePendingContracts/open_tag
--- PASS: TestPredicateUsesBelongToEachCall (0.20s)
=== CONT  TestCheckedPredicateFalseReadDirections/&&_operand
=== CONT  TestCheckedPredicateFalseReadDirections/||_operand
=== CONT  TestCheckedPredicateFalseReadDirections/conditional
=== CONT  TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== RUN   TestCheckedPredicatePendingContracts/overload_in_.a
=== CONT  TestCheckedPredicateFalseReadDirections/spread
=== CONT  TestCheckedPredicateFalseReadDirections/getter
=== RUN   TestCheckedPredicatePendingContracts/loose_false_comparison
=== CONT  TestCheckedPredicateFalseReadDirections/else
=== CONT  TestCheckedPredicateFalseReadDirections/renarrowed
=== CONT  TestCheckedPredicateFalseReadDirections/capture
=== RUN   TestCheckedPredicatePendingContracts/open_enum_union_tag
=== CONT  TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== CONT  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_read
=== CONT  TestCheckedPredicateFalseReadDirections/false_branch
=== CONT  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== CONT  TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_pass
=== CONT  TestPredicateUseRegions/unused
=== CONT  TestPredicateUseRegions/indexed_argument
=== CONT  TestPredicateUseRegions/indexed_spelling
=== CONT  TestPredicateUseRegions/closure_read
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_domain
--- FAIL: TestCheckedPredicateFalseReadDirections (0.01s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/empty_branches (0.21s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.19s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.17s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.16s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.13s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.13s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.07s)
    --- PASS: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/semver's_toString (0.10s)
=== CONT  TestPredicateUseRegions/short_circuit
=== CONT  TestPredicateUseRegions/compound_result_alias
=== CONT  TestPredicateUseRegions/const_result_alias
--- PASS: TestCheckedPredicatePendingContracts (0.90s)
    --- PASS: TestCheckedPredicatePendingContracts/untagged (0.18s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag (0.19s)
    --- PASS: TestCheckedPredicatePendingContracts/overload_in_.a (0.10s)
    --- PASS: TestCheckedPredicatePendingContracts/loose_false_comparison (0.11s)
    --- PASS: TestCheckedPredicatePendingContracts/open_enum_union_tag (0.06s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_read (0.07s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_pass (0.10s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_domain (0.10s)
=== CONT  TestPredicateUseRegions/false_continuation
=== CONT  TestPredicateUseRegions/throw_continuation
=== CONT  TestPredicateUseRegions/return_continuation
=== CONT  TestPredicateUseRegions/both
=== CONT  TestPredicateUseRegions/false
=== CONT  TestPredicateUseRegions/true
=== CONT  TestPredicateUseRegions/loose_false_comparison
=== CONT  TestPredicateUseRegions/switch_true_default
=== CONT  TestPredicateUseRegions/comma_continuation
=== CONT  TestPredicateUseRegions/later_read
=== CONT  TestPredicateUseRegions/false_comparison
=== CONT  TestPredicateUseRegions/replacement
=== CONT  TestPredicateUseRegions/loop_read
=== CONT  TestPredicateUseRegions/reversed_false_comparison
=== CONT  TestPredicateUseRegions/computed_index
=== CONT  TestPredicateUseRegions/true_inequality
--- PASS: TestPredicateUseRegions (0.01s)
    --- PASS: TestPredicateUseRegions/unused (0.07s)
    --- PASS: TestPredicateUseRegions/indexed_argument (0.09s)
    --- PASS: TestPredicateUseRegions/indexed_spelling (0.09s)
    --- PASS: TestPredicateUseRegions/closure_read (0.09s)
    --- PASS: TestPredicateUseRegions/short_circuit (0.08s)
    --- PASS: TestPredicateUseRegions/false_continuation (0.08s)
    --- PASS: TestPredicateUseRegions/const_result_alias (0.09s)
    --- PASS: TestPredicateUseRegions/compound_result_alias (0.09s)
    --- PASS: TestPredicateUseRegions/throw_continuation (0.09s)
    --- PASS: TestPredicateUseRegions/false (0.04s)
    --- PASS: TestPredicateUseRegions/both (0.10s)
    --- PASS: TestPredicateUseRegions/return_continuation (0.12s)
    --- PASS: TestPredicateUseRegions/true (0.09s)
    --- PASS: TestPredicateUseRegions/loose_false_comparison (0.12s)
    --- PASS: TestPredicateUseRegions/comma_continuation (0.12s)
    --- PASS: TestPredicateUseRegions/later_read (0.17s)
    --- PASS: TestPredicateUseRegions/switch_true_default (0.19s)
    --- PASS: TestPredicateUseRegions/false_comparison (0.15s)
    --- PASS: TestPredicateUseRegions/replacement (0.11s)
    --- PASS: TestPredicateUseRegions/reversed_false_comparison (0.08s)
    --- PASS: TestPredicateUseRegions/loop_read (0.09s)
    --- PASS: TestPredicateUseRegions/true_inequality (0.06s)
    --- PASS: TestPredicateUseRegions/computed_index (0.09s)
=== RUN   TestCheckedPredicateBranches/false
=== RUN   TestCheckedPredicateBranches/false_read
=== RUN   TestCheckedPredicateBranches/asserts
=== RUN   TestCheckedPredicateBranches/valid
--- PASS: TestCheckedPredicateBranches (5.85s)
    --- PASS: TestCheckedPredicateBranches/true (3.37s)
    --- PASS: TestCheckedPredicateBranches/false (0.78s)
    --- PASS: TestCheckedPredicateBranches/false_read (0.55s)
    --- PASS: TestCheckedPredicateBranches/asserts (0.53s)
    --- PASS: TestCheckedPredicateBranches/valid (0.62s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	5.930s
FAIL
````

`M2-oracle exit=1 wall_seconds=86`

````text
--- PASS: TestCheckedPredicateCountsAreRecorded (10.95s)
--- FAIL: TestPredicateFalseRead (0.00s)
    --- FAIL: TestPredicateFalseRead/spread (0.22s)
    --- FAIL: TestPredicateFalseRead/again (0.22s)
    --- FAIL: TestPredicateFalseRead/getter (0.32s)
    --- PASS: TestPredicateFalseRead/capture (35.02s)
    --- PASS: TestPredicateFalseRead/else_read (23.33s)
    --- PASS: TestPredicateFalseRead/renarrowed (35.05s)
    --- PASS: TestPredicateFalseRead/no_read (35.08s)
    --- PASS: TestPredicateFalseRead/early_exit (1.02s)
    --- PASS: TestPredicateFalseRead/false_read (1.07s)
--- FAIL: TestCheckedPredicateOracle (55.29s)
    --- FAIL: TestCheckedPredicateOracle/true (6.08s)
    --- PASS: TestCheckedPredicateOracle/true_unused (29.40s)
    --- PASS: TestCheckedPredicateOracle/false (1.09s)
    --- PASS: TestCheckedPredicateOracle/false_alias (1.12s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (1.04s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.14s)
    --- PASS: TestCheckedPredicateOracle/valid (1.07s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (0.98s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (1.08s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.08s)
    --- PASS: TestCheckedPredicateOracle/condition_once (1.09s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.28s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (1.04s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.31s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.20s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.45s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.14s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.32s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.36s)
````

Complete output, preserving every red subtest’s whole diagnostic and any recorded/measured rows verbatim:

````text
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
=== RUN   TestCheckedPredicateOracle
=== PAUSE TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateCountsAreRecorded
=== PAUSE TestCheckedPredicateCountsAreRecorded
=== CONT  TestPredicateFalseRead
=== RUN   TestPredicateFalseRead/no_read
=== PAUSE TestPredicateFalseRead/no_read
=== RUN   TestPredicateFalseRead/false_read
=== PAUSE TestPredicateFalseRead/false_read
=== RUN   TestPredicateFalseRead/else_read
=== CONT  TestCheckedPredicateCountsAreRecorded
=== CONT  TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateOracle/true
=== PAUSE TestPredicateFalseRead/else_read
=== RUN   TestPredicateFalseRead/early_exit
=== PAUSE TestPredicateFalseRead/early_exit
=== RUN   TestPredicateFalseRead/capture
=== PAUSE TestPredicateFalseRead/capture
=== RUN   TestPredicateFalseRead/renarrowed
=== PAUSE TestPredicateFalseRead/renarrowed
=== RUN   TestPredicateFalseRead/getter
=== PAUSE TestPredicateFalseRead/getter
=== RUN   TestPredicateFalseRead/spread
=== PAUSE TestPredicateFalseRead/spread
=== RUN   TestPredicateFalseRead/again
=== PAUSE TestPredicateFalseRead/again
=== CONT  TestPredicateFalseRead/no_read
=== CONT  TestPredicateFalseRead/renarrowed
=== CONT  TestPredicateFalseRead/capture
=== NAME  TestCheckedPredicateOracle/true
    predicates_checked_test.go:99: direction counts {Proven:0 Checked:1 Unobservable:1 Sites:[{Where:/tmp/adamic-gate/TestCheckedPredicateOracletrue1423480830/001/true.ts:2:50 Function:isNumber Overload:0 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:unobservable Reason:string | number -> number}]}]}
=== RUN   TestCheckedPredicateOracle/true_unused
--- PASS: TestCheckedPredicateCountsAreRecorded (10.95s)
=== CONT  TestPredicateFalseRead/spread
    predicate_false_read_test.go:46: /workspace/adamic/internal/oracle/testdata/predicate_false_read/spread.ts:24:17: stage 0 can't lower an array of never yet
=== CONT  TestPredicateFalseRead/again
    predicate_false_read_test.go:46: /workspace/adamic/internal/oracle/testdata/predicate_false_read/again.ts:11:19: stage 0 can't lower an array of unknown with erased element storage (retain its declared element type before reading elements) yet
=== CONT  TestPredicateFalseRead/getter
    predicate_false_read_test.go:46: /workspace/adamic/internal/oracle/testdata/predicate_false_read/getter.ts:37:17: stage 0 can't lower a read narrowed to undefined of a property a call may write yet
=== CONT  TestPredicateFalseRead/else_read
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 miss 27.68
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 27.68
=== NAME  TestPredicateFalseRead/no_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 27.68
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 22.51
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 27.68
=== CONT  TestPredicateFalseRead/early_exit
=== CONT  TestPredicateFalseRead/false_read
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
    predicates_checked_test.go:164: true_unused lying body exits 70; removed-true-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.02
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
--- FAIL: TestPredicateFalseRead (0.00s)
    --- FAIL: TestPredicateFalseRead/spread (0.22s)
    --- FAIL: TestPredicateFalseRead/again (0.22s)
    --- FAIL: TestPredicateFalseRead/getter (0.32s)
    --- PASS: TestPredicateFalseRead/capture (35.02s)
    --- PASS: TestPredicateFalseRead/else_read (23.33s)
    --- PASS: TestPredicateFalseRead/renarrowed (35.05s)
    --- PASS: TestPredicateFalseRead/no_read (35.08s)
    --- PASS: TestPredicateFalseRead/early_exit (1.02s)
    --- PASS: TestPredicateFalseRead/false_read (1.07s)
=== NAME  TestCheckedPredicateOracle/false
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
=== RUN   TestCheckedPredicateOracle/false_alias
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
=== RUN   TestCheckedPredicateOracle/false_compound_unchanged
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
    predicates_checked_test.go:164: asserts lying body exits 70; removed-asserts-check mutant executes Node result "called\nasserted\n" and is caught
=== RUN   TestCheckedPredicateOracle/valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts_valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.02
=== RUN   TestCheckedPredicateOracle/condition_valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_lie
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
    predicates_checked_test.go:164: condition_lie lying body exits 70; removed-asserts-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/condition_once
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_once_lie
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
    predicates_checked_test.go:164: condition_once_lie lying body exits 70; removed-asserts-check mutant executes Node result "argument\ncontinued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_unchanged
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.02
=== RUN   TestCheckedPredicateOracle/assert_not_node
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
    predicates_checked_test.go:164: assert_not_node lying body exits 70; removed-asserts-check mutant executes Node result "asserting\nnarrowed\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_switch_true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
    predicates_checked_test.go:164: false_switch_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_eq_false
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
    predicates_checked_test.go:164: false_eq_false lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_ne_true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
    predicates_checked_test.go:164: false_ne_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_false_eq
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
    predicates_checked_test.go:164: false_false_eq lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_comma
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types adc4f20d54c0 hit 0.02
    predicates_checked_test.go:164: false_comma lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
--- FAIL: TestCheckedPredicateOracle (55.29s)
    --- FAIL: TestCheckedPredicateOracle/true (6.08s)
    --- PASS: TestCheckedPredicateOracle/true_unused (29.40s)
    --- PASS: TestCheckedPredicateOracle/false (1.09s)
    --- PASS: TestCheckedPredicateOracle/false_alias (1.12s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (1.04s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.14s)
    --- PASS: TestCheckedPredicateOracle/valid (1.07s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (0.98s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (1.08s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.08s)
    --- PASS: TestCheckedPredicateOracle/condition_once (1.09s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.28s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (1.04s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.31s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.20s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.45s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.14s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.32s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.36s)
FAIL
gate cache: native hits=0 misses=19
gate cache: node hits=0 misses=62
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	55.386s
FAIL
````

````text
M2 restored; git diff --exit-code: 0
````

### M3

````text
M3: old text occurs exactly once; applied
````

````go
return falseMet && (!trueMet || reduced)
→
return falseMet || trueMet && reduced && false
````


`M3-lower exit=1 wall_seconds=31`

````text
--- PASS: TestPredicateUsesBelongToEachCall (0.21s)
--- FAIL: TestCheckedPredicateFalseReadDirections (0.00s)
    --- PASS: TestCheckedPredicateFalseReadDirections/empty_branches (0.18s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.22s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.17s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.16s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.15s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.15s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.23s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.24s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.20s)
    --- PASS: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.22s)
    --- PASS: TestCheckedPredicateFalseReadDirections/semver's_toString (0.17s)
--- PASS: TestCheckedPredicatePendingContracts (1.18s)
    --- PASS: TestCheckedPredicatePendingContracts/untagged (0.18s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag (0.18s)
    --- PASS: TestCheckedPredicatePendingContracts/overload_in_.a (0.10s)
    --- PASS: TestCheckedPredicatePendingContracts/loose_false_comparison (0.07s)
    --- PASS: TestCheckedPredicatePendingContracts/open_enum_union_tag (0.11s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_read (0.34s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_pass (0.13s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_domain (0.07s)
--- PASS: TestPredicateUseRegions (0.01s)
    --- PASS: TestPredicateUseRegions/unused (0.09s)
    --- PASS: TestPredicateUseRegions/switch_true_default (0.08s)
    --- PASS: TestPredicateUseRegions/comma_continuation (0.13s)
    --- PASS: TestPredicateUseRegions/reversed_false_comparison (0.12s)
    --- PASS: TestPredicateUseRegions/loose_false_comparison (0.06s)
    --- PASS: TestPredicateUseRegions/true_inequality (0.12s)
    --- PASS: TestPredicateUseRegions/replacement (0.09s)
    --- PASS: TestPredicateUseRegions/later_read (0.08s)
    --- PASS: TestPredicateUseRegions/false_comparison (0.12s)
    --- PASS: TestPredicateUseRegions/loop_read (0.07s)
    --- PASS: TestPredicateUseRegions/indexed_argument (0.07s)
    --- PASS: TestPredicateUseRegions/computed_index (0.08s)
    --- PASS: TestPredicateUseRegions/indexed_spelling (0.07s)
    --- PASS: TestPredicateUseRegions/closure_read (0.09s)
    --- PASS: TestPredicateUseRegions/short_circuit (0.08s)
    --- PASS: TestPredicateUseRegions/const_result_alias (0.10s)
    --- PASS: TestPredicateUseRegions/compound_result_alias (0.12s)
    --- PASS: TestPredicateUseRegions/false_continuation (0.09s)
    --- PASS: TestPredicateUseRegions/throw_continuation (0.10s)
    --- PASS: TestPredicateUseRegions/return_continuation (0.07s)
    --- PASS: TestPredicateUseRegions/both (0.09s)
    --- PASS: TestPredicateUseRegions/false (0.08s)
    --- PASS: TestPredicateUseRegions/true (0.06s)
--- PASS: TestCheckedPredicateBranches (5.83s)
    --- PASS: TestCheckedPredicateBranches/true (3.33s)
    --- PASS: TestCheckedPredicateBranches/false (0.55s)
    --- PASS: TestCheckedPredicateBranches/false_read (0.57s)
    --- PASS: TestCheckedPredicateBranches/asserts (0.74s)
    --- PASS: TestCheckedPredicateBranches/valid (0.64s)
````

Complete output, preserving every red subtest’s whole diagnostic and any recorded/measured rows verbatim:

````text
=== RUN   TestCheckedPredicateBranches
=== PAUSE TestCheckedPredicateBranches
=== RUN   TestCheckedPredicatePendingContracts
=== PAUSE TestCheckedPredicatePendingContracts
=== RUN   TestCheckedPredicateFalseReadDirections
=== PAUSE TestCheckedPredicateFalseReadDirections
=== RUN   TestPredicateUseRegions
=== PAUSE TestPredicateUseRegions
=== RUN   TestPredicateUsesBelongToEachCall
=== PAUSE TestPredicateUsesBelongToEachCall
=== CONT  TestCheckedPredicateBranches
=== CONT  TestPredicateUseRegions
=== CONT  TestCheckedPredicateFalseReadDirections
=== CONT  TestPredicateUsesBelongToEachCall
=== CONT  TestCheckedPredicatePendingContracts
=== RUN   TestCheckedPredicatePendingContracts/untagged
=== RUN   TestPredicateUseRegions/unused
=== RUN   TestCheckedPredicateFalseReadDirections/empty_branches
=== PAUSE TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestCheckedPredicateFalseReadDirections/semver's_toString
=== PAUSE TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== PAUSE TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== RUN   TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== PAUSE TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== RUN   TestCheckedPredicateFalseReadDirections/false_branch
=== PAUSE TestCheckedPredicateFalseReadDirections/false_branch
=== RUN   TestCheckedPredicateFalseReadDirections/else
=== PAUSE TestCheckedPredicateFalseReadDirections/else
=== RUN   TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== PAUSE TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== RUN   TestCheckedPredicateFalseReadDirections/capture
=== PAUSE TestCheckedPredicateFalseReadDirections/capture
=== RUN   TestCheckedPredicateFalseReadDirections/renarrowed
=== PAUSE TestCheckedPredicateFalseReadDirections/renarrowed
=== RUN   TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== PAUSE TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== RUN   TestCheckedPredicateFalseReadDirections/getter
=== PAUSE TestCheckedPredicateFalseReadDirections/getter
=== RUN   TestCheckedPredicateFalseReadDirections/spread
=== PAUSE TestCheckedPredicateFalseReadDirections/spread
=== RUN   TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== PAUSE TestPredicateUseRegions/unused
=== RUN   TestPredicateUseRegions/true
=== PAUSE TestPredicateUseRegions/true
=== RUN   TestCheckedPredicateBranches/true
=== RUN   TestPredicateUseRegions/false
=== PAUSE TestPredicateUseRegions/false
=== RUN   TestPredicateUseRegions/both
=== PAUSE TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== PAUSE TestPredicateUseRegions/both
=== RUN   TestCheckedPredicateFalseReadDirections/conditional
=== PAUSE TestCheckedPredicateFalseReadDirections/conditional
=== RUN   TestCheckedPredicateFalseReadDirections/||_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/||_operand
=== RUN   TestCheckedPredicateFalseReadDirections/&&_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/&&_operand
=== RUN   TestCheckedPredicateFalseReadDirections/const_alias
=== PAUSE TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestPredicateUseRegions/return_continuation
=== PAUSE TestPredicateUseRegions/return_continuation
=== RUN   TestPredicateUseRegions/throw_continuation
=== PAUSE TestPredicateUseRegions/throw_continuation
=== RUN   TestPredicateUseRegions/false_continuation
=== PAUSE TestPredicateUseRegions/false_continuation
=== RUN   TestPredicateUseRegions/const_result_alias
=== PAUSE TestPredicateUseRegions/const_result_alias
=== RUN   TestPredicateUseRegions/compound_result_alias
=== PAUSE TestPredicateUseRegions/compound_result_alias
=== RUN   TestPredicateUseRegions/short_circuit
=== PAUSE TestPredicateUseRegions/short_circuit
=== RUN   TestPredicateUseRegions/closure_read
=== PAUSE TestPredicateUseRegions/closure_read
=== RUN   TestPredicateUseRegions/indexed_spelling
=== PAUSE TestPredicateUseRegions/indexed_spelling
=== RUN   TestPredicateUseRegions/indexed_argument
=== PAUSE TestPredicateUseRegions/indexed_argument
=== RUN   TestPredicateUseRegions/computed_index
=== PAUSE TestPredicateUseRegions/computed_index
=== RUN   TestPredicateUseRegions/loop_read
=== PAUSE TestPredicateUseRegions/loop_read
=== RUN   TestPredicateUseRegions/later_read
=== PAUSE TestPredicateUseRegions/later_read
=== RUN   TestPredicateUseRegions/replacement
=== PAUSE TestPredicateUseRegions/replacement
=== RUN   TestPredicateUseRegions/false_comparison
=== PAUSE TestPredicateUseRegions/false_comparison
=== RUN   TestPredicateUseRegions/loose_false_comparison
=== PAUSE TestPredicateUseRegions/loose_false_comparison
=== RUN   TestPredicateUseRegions/true_inequality
=== PAUSE TestPredicateUseRegions/true_inequality
=== RUN   TestPredicateUseRegions/reversed_false_comparison
=== PAUSE TestPredicateUseRegions/reversed_false_comparison
=== RUN   TestPredicateUseRegions/switch_true_default
=== PAUSE TestPredicateUseRegions/switch_true_default
=== RUN   TestPredicateUseRegions/comma_continuation
=== PAUSE TestPredicateUseRegions/comma_continuation
=== CONT  TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== CONT  TestCheckedPredicateFalseReadDirections/const_alias
=== RUN   TestCheckedPredicatePendingContracts/open_tag
--- PASS: TestPredicateUsesBelongToEachCall (0.21s)
=== CONT  TestCheckedPredicateFalseReadDirections/&&_operand
=== CONT  TestCheckedPredicateFalseReadDirections/||_operand
=== CONT  TestCheckedPredicateFalseReadDirections/conditional
=== CONT  TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== RUN   TestCheckedPredicatePendingContracts/overload_in_.a
=== CONT  TestCheckedPredicateFalseReadDirections/spread
=== RUN   TestCheckedPredicatePendingContracts/loose_false_comparison
=== CONT  TestCheckedPredicateFalseReadDirections/getter
=== CONT  TestCheckedPredicateFalseReadDirections/false_branch
=== CONT  TestCheckedPredicateFalseReadDirections/renarrowed
=== RUN   TestCheckedPredicatePendingContracts/open_enum_union_tag
=== CONT  TestCheckedPredicateFalseReadDirections/capture
=== CONT  TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== CONT  TestCheckedPredicateFalseReadDirections/else
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_read
=== CONT  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== CONT  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== CONT  TestCheckedPredicateFalseReadDirections/semver's_toString
=== NAME  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
    predicates_false_read_test.go:89: false side checked true, want false
=== CONT  TestPredicateUseRegions/unused
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_pass
=== CONT  TestPredicateUseRegions/comma_continuation
--- FAIL: TestCheckedPredicateFalseReadDirections (0.00s)
    --- PASS: TestCheckedPredicateFalseReadDirections/empty_branches (0.18s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.22s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.17s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.16s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.15s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.15s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.23s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.24s)
    --- FAIL: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.20s)
    --- PASS: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.22s)
    --- PASS: TestCheckedPredicateFalseReadDirections/semver's_toString (0.17s)
=== CONT  TestPredicateUseRegions/switch_true_default
=== CONT  TestPredicateUseRegions/reversed_false_comparison
=== RUN   TestCheckedPredicatePendingContracts/open_tag_false_domain
=== CONT  TestPredicateUseRegions/true_inequality
=== CONT  TestPredicateUseRegions/loose_false_comparison
=== CONT  TestPredicateUseRegions/false_comparison
--- PASS: TestCheckedPredicatePendingContracts (1.18s)
    --- PASS: TestCheckedPredicatePendingContracts/untagged (0.18s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag (0.18s)
    --- PASS: TestCheckedPredicatePendingContracts/overload_in_.a (0.10s)
    --- PASS: TestCheckedPredicatePendingContracts/loose_false_comparison (0.07s)
    --- PASS: TestCheckedPredicatePendingContracts/open_enum_union_tag (0.11s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_read (0.34s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_pass (0.13s)
    --- PASS: TestCheckedPredicatePendingContracts/open_tag_false_domain (0.07s)
=== CONT  TestPredicateUseRegions/replacement
=== CONT  TestPredicateUseRegions/later_read
=== CONT  TestPredicateUseRegions/loop_read
=== CONT  TestPredicateUseRegions/computed_index
=== CONT  TestPredicateUseRegions/indexed_argument
=== CONT  TestPredicateUseRegions/indexed_spelling
=== CONT  TestPredicateUseRegions/closure_read
=== CONT  TestPredicateUseRegions/short_circuit
=== CONT  TestPredicateUseRegions/compound_result_alias
=== CONT  TestPredicateUseRegions/const_result_alias
=== CONT  TestPredicateUseRegions/false_continuation
=== CONT  TestPredicateUseRegions/throw_continuation
=== CONT  TestPredicateUseRegions/return_continuation
=== CONT  TestPredicateUseRegions/both
=== CONT  TestPredicateUseRegions/false
=== CONT  TestPredicateUseRegions/true
--- PASS: TestPredicateUseRegions (0.01s)
    --- PASS: TestPredicateUseRegions/unused (0.09s)
    --- PASS: TestPredicateUseRegions/switch_true_default (0.08s)
    --- PASS: TestPredicateUseRegions/comma_continuation (0.13s)
    --- PASS: TestPredicateUseRegions/reversed_false_comparison (0.12s)
    --- PASS: TestPredicateUseRegions/loose_false_comparison (0.06s)
    --- PASS: TestPredicateUseRegions/true_inequality (0.12s)
    --- PASS: TestPredicateUseRegions/replacement (0.09s)
    --- PASS: TestPredicateUseRegions/later_read (0.08s)
    --- PASS: TestPredicateUseRegions/false_comparison (0.12s)
    --- PASS: TestPredicateUseRegions/loop_read (0.07s)
    --- PASS: TestPredicateUseRegions/indexed_argument (0.07s)
    --- PASS: TestPredicateUseRegions/computed_index (0.08s)
    --- PASS: TestPredicateUseRegions/indexed_spelling (0.07s)
    --- PASS: TestPredicateUseRegions/closure_read (0.09s)
    --- PASS: TestPredicateUseRegions/short_circuit (0.08s)
    --- PASS: TestPredicateUseRegions/const_result_alias (0.10s)
    --- PASS: TestPredicateUseRegions/compound_result_alias (0.12s)
    --- PASS: TestPredicateUseRegions/false_continuation (0.09s)
    --- PASS: TestPredicateUseRegions/throw_continuation (0.10s)
    --- PASS: TestPredicateUseRegions/return_continuation (0.07s)
    --- PASS: TestPredicateUseRegions/both (0.09s)
    --- PASS: TestPredicateUseRegions/false (0.08s)
    --- PASS: TestPredicateUseRegions/true (0.06s)
=== RUN   TestCheckedPredicateBranches/false
=== RUN   TestCheckedPredicateBranches/false_read
=== RUN   TestCheckedPredicateBranches/asserts
=== RUN   TestCheckedPredicateBranches/valid
--- PASS: TestCheckedPredicateBranches (5.83s)
    --- PASS: TestCheckedPredicateBranches/true (3.33s)
    --- PASS: TestCheckedPredicateBranches/false (0.55s)
    --- PASS: TestCheckedPredicateBranches/false_read (0.57s)
    --- PASS: TestCheckedPredicateBranches/asserts (0.74s)
    --- PASS: TestCheckedPredicateBranches/valid (0.64s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	5.902s
FAIL
````

`M3-oracle exit=1 wall_seconds=94`

````text
--- PASS: TestCheckedPredicateCountsAreRecorded (11.87s)
--- FAIL: TestPredicateFalseRead (0.00s)
    --- FAIL: TestPredicateFalseRead/again (6.97s)
    --- PASS: TestPredicateFalseRead/capture (30.58s)
    --- PASS: TestPredicateFalseRead/early_exit (25.74s)
    --- PASS: TestPredicateFalseRead/renarrowed (37.62s)
    --- PASS: TestPredicateFalseRead/no_read (37.65s)
    --- FAIL: TestPredicateFalseRead/spread (0.33s)
    --- FAIL: TestPredicateFalseRead/getter (0.34s)
    --- PASS: TestPredicateFalseRead/false_read (0.99s)
    --- PASS: TestPredicateFalseRead/else_read (1.08s)
--- FAIL: TestCheckedPredicateOracle (58.80s)
    --- FAIL: TestCheckedPredicateOracle/true (6.92s)
    --- PASS: TestCheckedPredicateOracle/true_unused (31.26s)
    --- PASS: TestCheckedPredicateOracle/false (1.06s)
    --- PASS: TestCheckedPredicateOracle/false_alias (0.86s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (0.93s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.16s)
    --- PASS: TestCheckedPredicateOracle/valid (1.00s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (0.93s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (0.91s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.43s)
    --- PASS: TestCheckedPredicateOracle/condition_once (1.09s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.19s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (0.99s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.26s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.41s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.35s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.31s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (2.14s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.58s)
````

Complete output, preserving every red subtest’s whole diagnostic and any recorded/measured rows verbatim:

````text
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
=== RUN   TestCheckedPredicateOracle
=== PAUSE TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateCountsAreRecorded
=== PAUSE TestCheckedPredicateCountsAreRecorded
=== CONT  TestPredicateFalseRead
=== CONT  TestCheckedPredicateCountsAreRecorded
=== CONT  TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateOracle/true
=== RUN   TestPredicateFalseRead/no_read
=== PAUSE TestPredicateFalseRead/no_read
=== RUN   TestPredicateFalseRead/false_read
=== PAUSE TestPredicateFalseRead/false_read
=== RUN   TestPredicateFalseRead/else_read
=== PAUSE TestPredicateFalseRead/else_read
=== RUN   TestPredicateFalseRead/early_exit
=== PAUSE TestPredicateFalseRead/early_exit
=== RUN   TestPredicateFalseRead/capture
=== PAUSE TestPredicateFalseRead/capture
=== RUN   TestPredicateFalseRead/renarrowed
=== PAUSE TestPredicateFalseRead/renarrowed
=== RUN   TestPredicateFalseRead/getter
=== PAUSE TestPredicateFalseRead/getter
=== RUN   TestPredicateFalseRead/spread
=== PAUSE TestPredicateFalseRead/spread
=== RUN   TestPredicateFalseRead/again
=== PAUSE TestPredicateFalseRead/again
=== CONT  TestPredicateFalseRead/no_read
=== CONT  TestPredicateFalseRead/renarrowed
=== CONT  TestPredicateFalseRead/again
=== NAME  TestCheckedPredicateOracle/true
    predicates_checked_test.go:99: direction counts {Proven:0 Checked:1 Unobservable:1 Sites:[{Where:/tmp/adamic-gate/TestCheckedPredicateOracletrue3932030645/001/true.ts:2:50 Function:isNumber Overload:0 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:unobservable Reason:string | number -> number}]}]}
=== RUN   TestCheckedPredicateOracle/true_unused
=== NAME  TestPredicateFalseRead/again
    predicate_false_read_test.go:46: /workspace/adamic/internal/oracle/testdata/predicate_false_read/again.ts:11:19: stage 0 can't lower an array of unknown with erased element storage (retain its declared element type before reading elements) yet
=== CONT  TestPredicateFalseRead/capture
--- PASS: TestCheckedPredicateCountsAreRecorded (11.87s)
=== CONT  TestPredicateFalseRead/early_exit
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 miss 29.23
=== NAME  TestPredicateFalseRead/no_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 29.23
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 29.23
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 29.21
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 24.55
=== CONT  TestPredicateFalseRead/else_read
=== CONT  TestPredicateFalseRead/false_read
=== CONT  TestPredicateFalseRead/spread
=== CONT  TestPredicateFalseRead/getter
=== NAME  TestPredicateFalseRead/spread
    predicate_false_read_test.go:46: /workspace/adamic/internal/oracle/testdata/predicate_false_read/spread.ts:24:17: stage 0 can't lower an array of never yet
=== NAME  TestPredicateFalseRead/getter
    predicate_false_read_test.go:46: /workspace/adamic/internal/oracle/testdata/predicate_false_read/getter.ts:37:17: stage 0 can't lower a read narrowed to undefined of a property a call may write yet
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
    predicates_checked_test.go:164: true_unused lying body exits 70; removed-true-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
--- FAIL: TestPredicateFalseRead (0.00s)
    --- FAIL: TestPredicateFalseRead/again (6.97s)
    --- PASS: TestPredicateFalseRead/capture (30.58s)
    --- PASS: TestPredicateFalseRead/early_exit (25.74s)
    --- PASS: TestPredicateFalseRead/renarrowed (37.62s)
    --- PASS: TestPredicateFalseRead/no_read (37.65s)
    --- FAIL: TestPredicateFalseRead/spread (0.33s)
    --- FAIL: TestPredicateFalseRead/getter (0.34s)
    --- PASS: TestPredicateFalseRead/false_read (0.99s)
    --- PASS: TestPredicateFalseRead/else_read (1.08s)
=== NAME  TestCheckedPredicateOracle/false
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.02
=== RUN   TestCheckedPredicateOracle/false_alias
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
=== RUN   TestCheckedPredicateOracle/false_compound_unchanged
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
    predicates_checked_test.go:164: asserts lying body exits 70; removed-asserts-check mutant executes Node result "called\nasserted\n" and is caught
=== RUN   TestCheckedPredicateOracle/valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts_valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_valid
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_lie
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
    predicates_checked_test.go:164: condition_lie lying body exits 70; removed-asserts-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/condition_once
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_once_lie
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
    predicates_checked_test.go:164: condition_once_lie lying body exits 70; removed-asserts-check mutant executes Node result "argument\ncontinued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_unchanged
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.03
=== RUN   TestCheckedPredicateOracle/assert_not_node
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
    predicates_checked_test.go:164: assert_not_node lying body exits 70; removed-asserts-check mutant executes Node result "asserting\nnarrowed\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_switch_true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.03
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
    predicates_checked_test.go:164: false_switch_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_eq_false
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.05
    predicates_checked_test.go:164: false_eq_false lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_ne_true
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
    predicates_checked_test.go:164: false_ne_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_false_eq
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
    predicates_checked_test.go:164: false_false_eq lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_comma
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 7a08a0fb9140 hit 0.01
    predicates_checked_test.go:164: false_comma lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
--- FAIL: TestCheckedPredicateOracle (58.80s)
    --- FAIL: TestCheckedPredicateOracle/true (6.92s)
    --- PASS: TestCheckedPredicateOracle/true_unused (31.26s)
    --- PASS: TestCheckedPredicateOracle/false (1.06s)
    --- PASS: TestCheckedPredicateOracle/false_alias (0.86s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (0.93s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.16s)
    --- PASS: TestCheckedPredicateOracle/valid (1.00s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (0.93s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (0.91s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.43s)
    --- PASS: TestCheckedPredicateOracle/condition_once (1.09s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.19s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (0.99s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.26s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.41s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.35s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.31s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (2.14s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.58s)
FAIL
gate cache: native hits=0 misses=19
gate cache: node hits=0 misses=62
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	58.893s
FAIL
````

````text
M3 restored; git diff --exit-code: 0
````

## Full lower suite

`full-lower exit=1 wall_seconds=1515`

````sh
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
````

Six failing leaf tests, with no environment failure indicated by their diagnostics.

Every failing test/subtest and its first direct diagnostic follows. Parents without a direct diagnostic are identified separately. Environment attribution is based on the emitted diagnostic, not on whether the test was expected to pass.

* `TestPredicateOverloadRuntime` — aggregate failure

````text
(no direct diagnostic; failing child subtests below)
````

* `TestPredicateOverloadRuntime/overload_some_true_only` — expectation/compiler failure

````text
predicates_overload_test.go:87: predicate counts: {Proven:0 Checked:1 Unobservable:1 Sites:[{Where:/tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_some_true_only1074505791/001/main.ts:7:9 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly number[] | undefined -> readonly number[]} {Direction:false Status:unobservable Reason:readonly number[] | undefined -> readonly number[]}]}]}, want [0 2 0] (proven, checked, unobservable)
````

* `TestPredicateOverloadRuntime/overload_checked` — expectation/compiler failure

````text
predicates_overload_test.go:87: predicate counts: {Proven:0 Checked:1 Unobservable:1 Sites:[{Where:/tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_checked2612932811/001/main.ts:3:50 Function:lie Overload:1 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:unobservable Reason:string | number -> number}]}]}, want [0 2 0] (proven, checked, unobservable)
````

* `TestBrandedSupertypeIdentityNarrowingLowers` — expectation/compiler failure

````text
predicates_branded_test.go:401: want T at TypeElement NotYet on the false side, got <nil>
````

* `TestKindPredicateMatchesTheWrittenComparison` — aggregate failure

````text
(no direct diagnostic; failing child subtests below)
````

* `TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag` — expectation/compiler failure

````text
enum_tag_kind_predicate_test.go:105: want "membership over an open tag domain", got <nil>
````

* `TestBrandedSupertypeIsJudgedWhereItIsLowered` — aggregate failure

````text
(no direct diagnostic; failing child subtests below)
````

* `TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims` — expectation/compiler failure

````text
predicates_branded_test.go:827: want NotYet with "checked predicate membership over an open tag domain", got <nil>
````

* `TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance` — expectation/compiler failure

````text
predicates_branded_test.go:827: want NotYet with "checked predicate membership over an open tag domain", got <nil>
````

Complete command output:

````text
--- FAIL: TestPredicateOverloadRuntime (23.33s)
    --- FAIL: TestPredicateOverloadRuntime/overload_some_true_only (0.07s)
        predicates_overload_test.go:87: predicate counts: {Proven:0 Checked:1 Unobservable:1 Sites:[{Where:/tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_some_true_only1074505791/001/main.ts:7:9 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly number[] | undefined -> readonly number[]} {Direction:false Status:unobservable Reason:readonly number[] | undefined -> readonly number[]}]}]}, want [0 2 0] (proven, checked, unobservable)
    --- FAIL: TestPredicateOverloadRuntime/overload_checked (0.09s)
        predicates_overload_test.go:87: predicate counts: {Proven:0 Checked:1 Unobservable:1 Sites:[{Where:/tmp/adamic-gate/TestPredicateOverloadRuntimeoverload_checked2612932811/001/main.ts:3:50 Function:lie Overload:1 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:unobservable Reason:string | number -> number}]}]}, want [0 2 0] (proven, checked, unobservable)
--- FAIL: TestBrandedSupertypeIdentityNarrowingLowers (0.61s)
    predicates_branded_test.go:401: want T at TypeElement NotYet on the false side, got <nil>
--- FAIL: TestKindPredicateMatchesTheWrittenComparison (0.00s)
    --- FAIL: TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag (0.12s)
        enum_tag_kind_predicate_test.go:105: want "membership over an open tag domain", got <nil>
--- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered (1.05s)
    --- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.09s)
        predicates_branded_test.go:827: want NotYet with "checked predicate membership over an open tag domain", got <nil>
    --- FAIL: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.14s)
        predicates_branded_test.go:827: want NotYet with "checked predicate membership over an open tag domain", got <nil>
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	1510.684s
FAIL
````

## Full oracle suite

`full-oracle exit=1 wall_seconds=3738`

````sh
go test ./internal/oracle -count=1 -timeout 120m > /tmp/oracle.log 2>&1
````

Twenty failing leaf tests: six compiler/count expectation failures; twelve assignment-proof environment failures because their helper hardcodes NODE_PATH to the absent repository stage3/api/node_modules rather than the externally provisioned package tree; one missing-WASI_SYSROOT environment failure; and one 90-second probe timeout (timing/infrastructure observation, root cause unproven). The twelve module-resolution failures occurred despite tsc 6.0.3 being installed outside the repository and on PATH. The installed package tree remained outside the repository as requested.

Every failing test/subtest and its first direct diagnostic follows. Parents without a direct diagnostic are identified separately. Environment attribution is based on the emitted diagnostic, not on whether the test was expected to pass.

* `TestAssignmentProofWiderTargetMutant` — ENVIRONMENT

````text
assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
````

* `TestAssignmentProofWiderTarget` — ENVIRONMENT

````text
assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
````

* `TestAssignmentProofCompoundMutant` — ENVIRONMENT

````text
assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
````

* `TestAssignmentProofCompound` — ENVIRONMENT

````text
assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
````

* `TestAssignmentProofChainedMutant` — ENVIRONMENT

````text
assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
````

* `TestAssignmentProofChained` — ENVIRONMENT

````text
assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
````

* `TestAssignmentProofReturnDefinedMutant` — ENVIRONMENT

````text
assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
````

* `TestAssignmentProofIfNarrowingMutant` — ENVIRONMENT

````text
assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
````

* `TestAssignmentProofReturnDefined` — ENVIRONMENT

````text
assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
````

* `TestAssignmentProofIfNarrowing` — ENVIRONMENT

````text
assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
````

* `TestAssignmentProofScannerKeywordMutant` — ENVIRONMENT

````text
assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
````

* `TestAssignmentProofScannerKeyword` — ENVIRONMENT

````text
assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
````

* `TestPredicateDirectionCountsAreRecorded` — expectation/compiler failure

````text
predicate_counts_test.go:100: predicate counts changed in testdata/counts/predicates/internal/lower/testdata/predicates/overload_some_objects.a.counts; update with -update-counts:
````

* `TestCheckedPredicateOracle` — aggregate failure

````text
(no direct diagnostic; failing child subtests below)
````

* `TestCheckedPredicateOracle/true` — expectation/compiler failure

````text
predicates_checked_test.go:99: direction counts {Proven:0 Checked:1 Unobservable:1 Sites:[{Where:/tmp/adamic-gate/TestCheckedPredicateOracletrue3897202539/001/true.ts:2:50 Function:isNumber Overload:0 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:unobservable Reason:string | number -> number}]}]}
````

* `TestStdoutColumnsRefusesWASI` — ENVIRONMENT

````text
library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT
````

* `TestFamiliesCatchAPlantedCase_CountsAreRecorded` — ENVIRONMENT/TIMING (root cause unproven)

````text
families_generated_test.go:5190: the probe of TestCountsAreRecorded was killed at 90 s:
````

* `TestPredicateCountsRegenerateDeterministically` — expectation/compiler failure

````text
count_files_test.go:197: regenerated counts differ in testdata/counts/predicates/internal/lower/testdata/predicates/overload_some_objects.a.counts:
````

* `TestPredicateFalseRead` — aggregate failure

````text
(no direct diagnostic; failing child subtests below)
````

* `TestPredicateFalseRead/again` — expectation/compiler failure

````text
predicate_false_read_test.go:46: /workspace/adamic/internal/oracle/testdata/predicate_false_read/again.ts:11:19: stage 0 can't lower an array of unknown with erased element storage (retain its declared element type before reading elements) yet
````

* `TestPredicateFalseRead/spread` — expectation/compiler failure

````text
predicate_false_read_test.go:46: /workspace/adamic/internal/oracle/testdata/predicate_false_read/spread.ts:24:17: stage 0 can't lower an array of never yet
````

* `TestPredicateFalseRead/getter` — expectation/compiler failure

````text
predicate_false_read_test.go:46: /workspace/adamic/internal/oracle/testdata/predicate_false_read/getter.ts:37:17: stage 0 can't lower a read narrowed to undefined of a property a call may write yet
````

Complete command output:

````text
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
--- FAIL: TestAssignmentProofWiderTargetMutant (0.05s)
    assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofWiderTarget (0.04s)
    assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompoundMutant (0.05s)
    assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompound (0.06s)
    assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofChainedMutant (0.06s)
    assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofChained (0.06s)
    assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefinedMutant (0.04s)
    assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowingMutant (0.05s)
    assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefined (0.07s)
    assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowing (0.06s)
    assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeywordMutant (0.05s)
    assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeyword (0.06s)
    assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestPredicateDirectionCountsAreRecorded (1.32s)
    predicate_counts_test.go:100: predicate counts changed in testdata/counts/predicates/internal/lower/testdata/predicates/overload_some_objects.a.counts; update with -update-counts:
        recorded: | internal/lower/testdata/predicates/overload_some_objects.a | 1 | 0 | 2 | 0 |
        measured: | internal/lower/testdata/predicates/overload_some_objects.a | 1 | 0 | 1 | 1 |
    predicate_counts_test.go:100: predicate counts changed in testdata/counts/predicates/internal/lower/testdata/predicates/overload_some_true_only.a.counts; update with -update-counts:
        recorded: | internal/lower/testdata/predicates/overload_some_true_only.a | 1 | 0 | 2 | 0 |
        measured: | internal/lower/testdata/predicates/overload_some_true_only.a | 1 | 0 | 1 | 1 |
    predicate_counts_test.go:100: predicate counts changed in testdata/counts/predicates/internal/lower/testdata/predicates/overload_checked.a.counts; update with -update-counts:
        recorded: | internal/lower/testdata/predicates/overload_checked.a | 1 | 0 | 2 | 0 |
        measured: | internal/lower/testdata/predicates/overload_checked.a | 1 | 0 | 1 | 1 |
--- FAIL: TestCheckedPredicateOracle (32.14s)
    --- FAIL: TestCheckedPredicateOracle/true (0.22s)
        predicates_checked_test.go:82: gate cache node miss
        predicates_checked_test.go:99: direction counts {Proven:0 Checked:1 Unobservable:1 Sites:[{Where:/tmp/adamic-gate/TestCheckedPredicateOracletrue3897202539/001/true.ts:2:50 Function:isNumber Overload:0 Directions:[{Direction:true Status:checked Reason:string | number -> number} {Direction:false Status:unobservable Reason:string | number -> number}]}]}
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
--- FAIL: TestStdoutColumnsRefusesWASI (0.95s)
    library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestFamiliesCatchAPlantedCase_CountsAreRecorded (90.02s)
    families_generated_test.go:5190: the probe of TestCountsAreRecorded was killed at 90 s:
        === RUN   TestCountsAreRecorded_000
        === PAUSE TestCountsAreRecorded_000
        === RUN   TestCountsAreRecorded_001
        === PAUSE TestCountsAreRecorded_001
        === RUN   TestCountsAreRecorded_002
        === PAUSE TestCountsAreRecorded_002
        === RUN   TestCountsAreRecorded_003
        === PAUSE TestCountsAreRecorded_003
        === RUN   TestCountsAreRecorded_004
        === PAUSE TestCountsAreRecorded_004
        === RUN   TestCountsAreRecorded_005
        === PAUSE TestCountsAreRecorded_005
        === RUN   TestCountsAreRecorded_006
        === PAUSE TestCountsAreRecorded_006
        === RUN   TestCountsAreRecorded_007
        === PAUSE TestCountsAreRecorded_007
        === RUN   TestCountsAreRecorded_008
        === PAUSE TestCountsAreRecorded_008
        === RUN   TestCountsAreRecorded_009
        === PAUSE TestCountsAreRecorded_009
        === RUN   TestCountsAreRecorded_010
        === PAUSE TestCountsAreRecorded_010
        === RUN   TestCountsAreRecorded_011
        === PAUSE TestCountsAreRecorded_011
        === RUN   TestCountsAreRecorded_012
        === PAUSE TestCountsAreRecorded_012
        === RUN   TestCountsAreRecorded_013
        === PAUSE TestCountsAreRecorded_013
        === RUN   TestCountsAreRecorded_014
        === PAUSE TestCountsAreRecorded_014
        === RUN   TestCountsAreRecorded_015
        === PAUSE TestCountsAreRecorded_015
        === RUN   TestCountsAreRecorded_016
        === PAUSE TestCountsAreRecorded_016
        === RUN   TestCountsAreRecorded_017
        === PAUSE TestCountsAreRecorded_017
        === RUN   TestCountsAreRecorded_018
        === PAUSE TestCountsAreRecorded_018
        === RUN   TestCountsAreRecorded_019
        === PAUSE TestCountsAreRecorded_019
        === RUN   TestCountsAreRecorded_020
        === PAUSE TestCountsAreRecorded_020
        === RUN   TestCountsAreRecorded_021
        === PAUSE TestCountsAreRecorded_021
        === RUN   TestCountsAreRecorded_022
        === PAUSE TestCountsAreRecorded_022
        === RUN   TestCountsAreRecorded_023
        === PAUSE TestCountsAreRecorded_023
        === RUN   TestCountsAreRecorded_024
        === PAUSE TestCountsAreRecorded_024
        === RUN   TestCountsAreRecorded_025
        === PAUSE TestCountsAreRecorded_025
        === RUN   TestCountsAreRecorded_026
        === PAUSE TestCountsAreRecorded_026
        === RUN   TestCountsAreRecorded_027
        === PAUSE TestCountsAreRecorded_027
        === RUN   TestCountsAreRecorded_028
        === PAUSE TestCountsAreRecorded_028
        === RUN   TestCountsAreRecorded_029
        === PAUSE TestCountsAreRecorded_029
        === RUN   TestCountsAreRecorded_030
        === PAUSE TestCountsAreRecorded_030
        === RUN   TestCountsAreRecorded_031
        === PAUSE TestCountsAreRecorded_031
        === RUN   TestCountsAreRecorded_little_streams
        === PAUSE TestCountsAreRecorded_little_streams
        === RUN   TestCountsAreRecorded_lots_streams
        === PAUSE TestCountsAreRecorded_lots_streams
        === RUN   TestCountsAreRecorded_stderr_only
        === PAUSE TestCountsAreRecorded_stderr_only
        === RUN   TestCountsAreRecorded_panic_edges
        === PAUSE TestCountsAreRecorded_panic_edges
        === RUN   TestCountsAreRecorded_read_after_print
        === PAUSE TestCountsAreRecorded_read_after_print
        === RUN   TestCountsAreRecorded_async_counts_payload
        === PAUSE TestCountsAreRecorded_async_counts_payload
        === RUN   TestCountsAreRecorded_async_counts_stable
        === PAUSE TestCountsAreRecorded_async_counts_stable
        === RUN   TestCountsAreRecorded_async_counts_reassigned
        === PAUSE TestCountsAreRecorded_async_counts_reassigned
        === RUN   TestCountsAreRecorded_async_counts_captured
        === PAUSE TestCountsAreRecorded_async_counts_captured
        === RUN   TestCountsAreRecorded_async_counts_expression_write
        === PAUSE TestCountsAreRecorded_async_counts_expression_write
        === RUN   TestCountsAreRecorded_async_counts_parameter
        === PAUSE TestCountsAreRecorded_async_counts_parameter
        === RUN   TestCountsAreRecorded_async_counts_thrown
        === PAUSE TestCountsAreRecorded_async_counts_thrown
        === RUN   TestCountsAreRecorded_async_counts_root_error
        === PAUSE TestCountsAreRecorded_async_counts_root_error
        === RUN   TestCountsAreRecorded_async_catch_return
        === PAUSE TestCountsAreRecorded_async_catch_return
        === RUN   TestCountsAreRecorded_async_catches
        === PAUSE TestCountsAreRecorded_async_catches
        === RUN   TestCountsAreRecorded_async_closures
        === PAUSE TestCountsAreRecorded_async_closures
        === RUN   TestCountsAreRecorded_async_conditions
        === PAUSE TestCountsAreRecorded_async_conditions
        === RUN   TestCountsAreRecorded_async_control
        === PAUSE TestCountsAreRecorded_async_control
        === RUN   TestCountsAreRecorded_async_escaped_capture
        === PAUSE TestCountsAreRecorded_async_escaped_capture
        === RUN   TestCountsAreRecorded_async_frame_capture_safe
        === PAUSE TestCountsAreRecorded_async_frame_capture_safe
        === RUN   TestCountsAreRecorded_async_fresh_holder
        === PAUSE TestCountsAreRecorded_async_fresh_holder
        === RUN   TestCountsAreRecorded_async_live_values
        === PAUSE TestCountsAreRecorded_async_live_values
        === RUN   TestCountsAreRecorded_async_methods
        === PAUSE TestCountsAreRecorded_async_methods
        === RUN   TestCountsAreRecorded_async_mutation
        === PAUSE TestCountsAreRecorded_async_mutation
        === RUN   TestCountsAreRecorded_async_operand_order
        === PAUSE TestCountsAreRecorded_async_operand_order
        === RUN   TestCountsAreRecorded_async_promise_expression
        === PAUSE TestCountsAreRecorded_async_promise_expression
        === RUN   TestCountsAreRecorded_async_receiver
        === PAUSE TestCountsAreRecorded_async_receiver
        === RUN   TestCountsAreRecorded_async_snapshots
        === PAUSE TestCountsAreRecorded_async_snapshots
        === RUN   TestCountsAreRecorded_async_void_value
        === PAUSE TestCountsAreRecorded_async_void_value
        === RUN   TestCountsAreRecorded_typed_arrays_integer
        === PAUSE TestCountsAreRecorded_typed_arrays_integer
        === RUN   TestCountsAreRecorded_typed_arrays_integer_stop
        === PAUSE TestCountsAreRecorded_typed_arrays_integer_stop
        === RUN   TestCountsAreRecorded_typed_arrays_sort_uint8array
        === PAUSE TestCountsAreRecorded_typed_arrays_sort_uint8array
        === RUN   TestCountsAreRecorded_typed_arrays_sort_uint16array
        === PAUSE TestCountsAreRecorded_typed_arrays_sort_uint16array
        === RUN   TestCountsAreRecorded_typed_arrays_sort_int32array
        === PAUSE TestCountsAreRecorded_typed_arrays_sort_int32array
        === RUN   TestCountsAreRecorded_typed_arrays_sort_float64array
        === PAUSE TestCountsAreRecorded_typed_arrays_sort_float64array
        === RUN   TestCountsAreRecorded_counters_shapes
        === PAUSE TestCountsAreRecorded_counters_shapes
        === RUN   TestCountsAreRecorded_counters_square_jump
        === PAUSE TestCountsAreRecorded_counters_square_jump
        === RUN   TestCountsAreRecorded_release_fast_graph
        === PAUSE TestCountsAreRecorded_release_fast_graph
        === RUN   TestCountsAreRecorded_release_fast_shared
        === PAUSE TestCountsAreRecorded_release_fast_shared
        === RUN   TestCountsAreRecorded_flat
        === PAUSE TestCountsAreRecorded_flat
        === RUN   TestCountsAreRecorded_array
        === PAUSE TestCountsAreRecorded_array
        === RUN   TestCountsAreRecorded_async_loop_body_capture
        === PAUSE TestCountsAreRecorded_async_loop_body_capture
        === RUN   TestCountsAreRecorded_async_finally_completion
        === PAUSE TestCountsAreRecorded_async_finally_completion
        === RUN   TestCountsAreRecorded_async_labels
        === PAUSE TestCountsAreRecorded_async_labels
        === RUN   TestCountsAreRecorded_async_labels_finally
        === PAUSE TestCountsAreRecorded_async_labels_finally
        === RUN   TestCountsAreRecorded_async_labels_block_finally
        === PAUSE TestCountsAreRecorded_async_labels_block_finally
        === RUN   TestCountsAreRecorded_label_block_edges
        === PAUSE TestCountsAreRecorded_label_block_edges
        === RUN   TestCountsAreRecorded_async_iterables_cells_labels
        === PAUSE TestCountsAreRecorded_async_iterables_cells_labels
        === RUN   TestCountsAreRecorded_label_block_break_sync
        === PAUSE TestCountsAreRecorded_label_block_break_sync
        === RUN   TestCountsAreRecorded_label_block_continue_sync
        === PAUSE TestCountsAreRecorded_label_block_continue_sync
        === RUN   TestCountsAreRecorded_label_block_switch_break_sync
        === PAUSE TestCountsAreRecorded_label_block_switch_break_sync
        === RUN   TestCountsAreRecorded_label_block_outer_continue_sync
        === PAUSE TestCountsAreRecorded_label_block_outer_continue_sync
        === RUN   TestCountsAreRecorded_label_block_break_async
        === PAUSE TestCountsAreRecorded_label_block_break_async
        === RUN   TestCountsAreRecorded_label_block_continue_async
        === PAUSE TestCountsAreRecorded_label_block_continue_async
        === RUN   TestCountsAreRecorded_label_block_switch_break_async
        === PAUSE TestCountsAreRecorded_label_block_switch_break_async
        === RUN   TestCountsAreRecorded_label_block_outer_continue_async
        === PAUSE TestCountsAreRecorded_label_block_outer_continue_async
        === RUN   TestCountsAreRecorded_async_finally_return
        === PAUSE TestCountsAreRecorded_async_finally_return
        === RUN   TestCountsAreRecorded_async_finally_throw
        === PAUSE TestCountsAreRecorded_async_finally_throw
        === RUN   TestCountsAreRecorded_async_finally_break
        === PAUSE TestCountsAreRecorded_async_finally_break
        === RUN   TestCountsAreRecorded_async_finally_continue
        === PAUSE TestCountsAreRecorded_async_finally_continue
        === RUN   TestCountsAreRecorded_async_finally_void_return
        === PAUSE TestCountsAreRecorded_async_finally_void_return
        === RUN   TestCountsAreRecorded_async_finally_snapshot
        === PAUSE TestCountsAreRecorded_async_finally_snapshot
        === RUN   TestCountsAreRecorded_async_gap_label_admission
        === PAUSE TestCountsAreRecorded_async_gap_label_admission
        === RUN   TestCountsAreRecorded_async_gap_catch_admission
        === PAUSE TestCountsAreRecorded_async_gap_catch_admission
        === RUN   TestCountsAreRecorded_async_gap_finally_admission
        === PAUSE TestCountsAreRecorded_async_gap_finally_admission
        === RUN   TestCountsAreRecorded_async_gap_finally_return_admission
        === PAUSE TestCountsAreRecorded_async_gap_finally_return_admission
        === RUN   TestCountsAreRecorded_async_iterable_string_ascii_function
        === PAUSE TestCountsAreRecorded_async_iterable_string_ascii_function
        === RUN   TestCountsAreRecorded_async_for_of_string
        === PAUSE TestCountsAreRecorded_async_for_of_string
        === RUN   TestCountsAreRecorded_async_for_of_map
        === PAUSE TestCountsAreRecorded_async_for_of_map
        === RUN   TestCountsAreRecorded_async_for_of_collections_mutation
        === PAUSE TestCountsAreRecorded_async_for_of_collections_mutation
        === RUN   TestCountsAreRecorded_async_iterable_string_ascii_module
        === PAUSE TestCountsAreRecorded_async_iterable_string_ascii_module
        === RUN   TestCountsAreRecorded_async_iterable_string_emoji_function
        === PAUSE TestCountsAreRecorded_async_iterable_string_emoji_function
        === RUN   TestCountsAreRecorded_async_iterable_string_emoji_module
        === PAUSE TestCountsAreRecorded_async_iterable_string_emoji_module
        === RUN   TestCountsAreRecorded_async_iterable_string_surrogates_function
        === PAUSE TestCountsAreRecorded_async_iterable_string_surrogates_function
        === RUN   TestCountsAreRecorded_async_iterable_string_surrogates_module
        === PAUSE TestCountsAreRecorded_async_iterable_string_surrogates_module
        === RUN   TestCountsAreRecorded_async_iterable_map_entries_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_entries_function
        === RUN   TestCountsAreRecorded_async_iterable_map_entries_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_entries_module
        === RUN   TestCountsAreRecorded_async_iterable_map_keys_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_keys_function
        === RUN   TestCountsAreRecorded_async_iterable_map_keys_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_keys_module
        === RUN   TestCountsAreRecorded_async_iterable_map_values_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_values_function
        === RUN   TestCountsAreRecorded_async_iterable_map_values_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_values_module
        === RUN   TestCountsAreRecorded_async_iterable_set_values_function
        === PAUSE TestCountsAreRecorded_async_iterable_set_values_function
        === RUN   TestCountsAreRecorded_async_iterable_set_values_module
        === PAUSE TestCountsAreRecorded_async_iterable_set_values_module
        === RUN   TestCountsAreRecorded_async_iterable_set_entries_function
        === PAUSE TestCountsAreRecorded_async_iterable_set_entries_function
        === RUN   TestCountsAreRecorded_async_iterable_set_entries_module
        === PAUSE TestCountsAreRecorded_async_iterable_set_entries_module
        === RUN   TestCountsAreRecorded_async_iterable_map_clear_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_clear_function
        === RUN   TestCountsAreRecorded_async_iterable_map_clear_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_clear_module
        === RUN   TestCountsAreRecorded_async_iterable_map_readd_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_readd_function
        === RUN   TestCountsAreRecorded_async_iterable_map_readd_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_readd_module
        === RUN   TestCountsAreRecorded_async_iterable_map_rehash_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_rehash_function
        === RUN   TestCountsAreRecorded_async_iterable_map_rehash_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_rehash_module
        === RUN   TestCountsAreRecorded_async_iterable_nested_maps_function
        === PAUSE TestCountsAreRecorded_async_iterable_nested_maps_function
        === RUN   TestCountsAreRecorded_async_iterable_nested_maps_module
        === PAUSE TestCountsAreRecorded_async_iterable_nested_maps_module
        === RUN   TestCountsAreRecorded_async_iterable_map_break_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_break_function
        === RUN   TestCountsAreRecorded_async_iterable_map_break_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_break_module
        === RUN   TestCountsAreRecorded_async_iterable_map_return_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_return_function
        === RUN   TestCountsAreRecorded_async_iterable_map_return_module
        === PAUSE TestCountsAreRecorded_async_iterable_map_return_module
        === RUN   TestCountsAreRecorded_async_iterable_string_break_function
        === PAUSE TestCountsAreRecorded_async_iterable_string_break_function
        === RUN   TestCountsAreRecorded_async_iterable_string_break_module
        === PAUSE TestCountsAreRecorded_async_iterable_string_break_module
        === RUN   TestCountsAreRecorded_async_iterable_set_return_function
        === PAUSE TestCountsAreRecorded_async_iterable_set_return_function
        === RUN   TestCountsAreRecorded_async_iterable_set_return_module
        === PAUSE TestCountsAreRecorded_async_iterable_set_return_module
        === RUN   TestCountsAreRecorded_async_for_of_string_admission
        === PAUSE TestCountsAreRecorded_async_for_of_string_admission
        === RUN   TestCountsAreRecorded_async_top_catch_binding
        === PAUSE TestCountsAreRecorded_async_top_catch_binding
        === RUN   TestCountsAreRecorded_async_top_catch_after_await
        === PAUSE TestCountsAreRecorded_async_top_catch_after_await
        === RUN   TestCountsAreRecorded_async_top_loop
        === PAUSE TestCountsAreRecorded_async_top_loop
        === RUN   TestCountsAreRecorded_async_for_of_body
        === PAUSE TestCountsAreRecorded_async_for_of_body
        === RUN   TestCountsAreRecorded_async_for_of_mutation
        === PAUSE TestCountsAreRecorded_async_for_of_mutation
        === RUN   TestCountsAreRecorded_async_for_of_return
        === PAUSE TestCountsAreRecorded_async_for_of_return
        === RUN   TestCountsAreRecorded_async_switch_body
        === PAUSE TestCountsAreRecorded_async_switch_body
        === RUN   TestCountsAreRecorded_async_loop_cells
        === PAUSE TestCountsAreRecorded_async_loop_cells
        === RUN   TestCountsAreRecorded_async_for_of_capture
        === PAUSE TestCountsAreRecorded_async_for_of_capture
        === RUN   TestCountsAreRecorded_async_for_of_cells
        === PAUSE TestCountsAreRecorded_async_for_of_cells
        === RUN   TestCountsAreRecorded_async_for_of_body_cells
        === PAUSE TestCountsAreRecorded_async_for_of_body_cells
        === RUN   TestCountsAreRecorded_async_for_of_break_cells
        === PAUSE TestCountsAreRecorded_async_for_of_break_cells
        === RUN   TestCountsAreRecorded_async_coverage_discard
        === PAUSE TestCountsAreRecorded_async_coverage_discard
        === RUN   TestCountsAreRecorded_async_coverage_parameters
        === PAUSE TestCountsAreRecorded_async_coverage_parameters
        === RUN   TestCountsAreRecorded_async_coverage_reject_eager
        === PAUSE TestCountsAreRecorded_async_coverage_reject_eager
        === RUN   TestCountsAreRecorded_async_coverage_reject_empty
        === PAUSE TestCountsAreRecorded_async_coverage_reject_empty
        === RUN   TestCountsAreRecorded_async_coverage_reject_nested
        === PAUSE TestCountsAreRecorded_async_coverage_reject_nested
        === RUN   TestCountsAreRecorded_async_coverage_typeof
        === PAUSE TestCountsAreRecorded_async_coverage_typeof
        === RUN   TestCountsAreRecorded_async_coverage_unions
        === PAUSE TestCountsAreRecorded_async_coverage_unions
        === RUN   TestCountsAreRecorded_async_coverage_values
        === PAUSE TestCountsAreRecorded_async_coverage_values
        === RUN   TestCountsAreRecorded_async_nested
        === PAUSE TestCountsAreRecorded_async_nested
        === RUN   TestCountsAreRecorded_async_plain
        === PAUSE TestCountsAreRecorded_async_plain
        === RUN   TestCountsAreRecorded_async_three
        === PAUSE TestCountsAreRecorded_async_three
        === RUN   TestCountsAreRecorded_async_throw
        === PAUSE TestCountsAreRecorded_async_throw
        === RUN   TestCountsAreRecorded_async_typeof
        === PAUSE TestCountsAreRecorded_async_typeof
        === RUN   TestCountsAreRecorded_slab_sizes
        === PAUSE TestCountsAreRecorded_slab_sizes
        === RUN   TestCountsAreRecorded_reuse_handover_fresh
        === PAUSE TestCountsAreRecorded_reuse_handover_fresh
        === RUN   TestCountsAreRecorded_async_iterable_map_boolean_function
        === PAUSE TestCountsAreRecorded_async_iterable_map_boolean_function
        === RUN   TestCountsAreRecorded_async_iterable_set_boolean_function
        === PAUSE TestCountsAreRecorded_async_iterable_set_boolean_function
        === CONT  TestCountsAreRecorded_000
        === CONT  TestCountsAreRecorded_label_block_switch_break_sync
        === CONT  TestCountsAreRecorded_label_block_continue_sync
        === CONT  TestCountsAreRecorded_async_counts_expression_write
        --- PASS: TestCountsAreRecorded_000 (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_set_boolean_function
        --- PASS: TestCountsAreRecorded_label_block_continue_sync (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_map_boolean_function
        --- PASS: TestCountsAreRecorded_label_block_switch_break_sync (0.01s)
        === CONT  TestCountsAreRecorded_reuse_handover_fresh
        --- PASS: TestCountsAreRecorded_async_iterable_set_boolean_function (0.00s)
        === CONT  TestCountsAreRecorded_slab_sizes
        --- PASS: TestCountsAreRecorded_async_counts_expression_write (0.01s)
        === CONT  TestCountsAreRecorded_async_typeof
        --- PASS: TestCountsAreRecorded_async_iterable_map_boolean_function (0.02s)
        === CONT  TestCountsAreRecorded_async_throw
        --- PASS: TestCountsAreRecorded_reuse_handover_fresh (0.02s)
        === CONT  TestCountsAreRecorded_async_three
        --- PASS: TestCountsAreRecorded_slab_sizes (0.02s)
        === CONT  TestCountsAreRecorded_async_plain
        --- PASS: TestCountsAreRecorded_async_throw (0.00s)
        === CONT  TestCountsAreRecorded_async_nested
        --- PASS: TestCountsAreRecorded_async_typeof (0.03s)
        === CONT  TestCountsAreRecorded_async_coverage_values
        --- PASS: TestCountsAreRecorded_async_nested (0.02s)
        === CONT  TestCountsAreRecorded_async_coverage_unions
        --- PASS: TestCountsAreRecorded_async_plain (0.03s)
        === CONT  TestCountsAreRecorded_async_coverage_typeof
        --- PASS: TestCountsAreRecorded_async_coverage_typeof (0.00s)
        === CONT  TestCountsAreRecorded_async_coverage_reject_nested
        --- PASS: TestCountsAreRecorded_async_coverage_unions (0.01s)
        === CONT  TestCountsAreRecorded_async_coverage_reject_empty
        --- PASS: TestCountsAreRecorded_async_coverage_reject_empty (0.00s)
        === CONT  TestCountsAreRecorded_async_coverage_reject_eager
        --- PASS: TestCountsAreRecorded_async_coverage_values (0.02s)
        === CONT  TestCountsAreRecorded_async_coverage_parameters
        --- PASS: TestCountsAreRecorded_async_coverage_reject_nested (0.01s)
        === CONT  TestCountsAreRecorded_async_coverage_discard
        --- PASS: TestCountsAreRecorded_async_coverage_discard (0.00s)
        === CONT  TestCountsAreRecorded_async_for_of_break_cells
        --- PASS: TestCountsAreRecorded_async_for_of_break_cells (0.00s)
        === CONT  TestCountsAreRecorded_async_for_of_body_cells
        --- PASS: TestCountsAreRecorded_async_three (0.05s)
        === CONT  TestCountsAreRecorded_async_for_of_cells
        --- PASS: TestCountsAreRecorded_async_for_of_cells (0.00s)
        === CONT  TestCountsAreRecorded_async_for_of_capture
        --- PASS: TestCountsAreRecorded_async_coverage_reject_eager (0.02s)
        === CONT  TestCountsAreRecorded_async_loop_cells
        --- PASS: TestCountsAreRecorded_async_loop_cells (0.00s)
        === CONT  TestCountsAreRecorded_async_switch_body
        --- PASS: TestCountsAreRecorded_async_switch_body (0.00s)
        === CONT  TestCountsAreRecorded_async_for_of_return
        --- PASS: TestCountsAreRecorded_async_coverage_parameters (0.02s)
        === CONT  TestCountsAreRecorded_async_for_of_mutation
        --- PASS: TestCountsAreRecorded_async_for_of_return (0.01s)
        === CONT  TestCountsAreRecorded_async_for_of_body
        --- PASS: TestCountsAreRecorded_async_for_of_body_cells (0.03s)
        === CONT  TestCountsAreRecorded_async_top_loop
        --- PASS: TestCountsAreRecorded_async_for_of_body (0.01s)
        === CONT  TestCountsAreRecorded_async_top_catch_after_await
        --- PASS: TestCountsAreRecorded_async_top_catch_after_await (0.00s)
        === CONT  TestCountsAreRecorded_async_for_of_string_admission
        --- PASS: TestCountsAreRecorded_async_for_of_string_admission (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_set_return_module
        --- PASS: TestCountsAreRecorded_async_top_loop (0.00s)
        === CONT  TestCountsAreRecorded_async_top_catch_binding
        --- PASS: TestCountsAreRecorded_async_top_catch_binding (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_set_return_function
        === CONT  TestCountsAreRecorded_async_iterable_string_break_module
        --- PASS: TestCountsAreRecorded_async_for_of_capture (0.03s)
        === CONT  TestCountsAreRecorded_async_iterable_string_break_function
        --- PASS: TestCountsAreRecorded_async_for_of_mutation (0.02s)
        --- PASS: TestCountsAreRecorded_async_iterable_set_return_module (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_map_return_module
        --- PASS: TestCountsAreRecorded_async_iterable_set_return_function (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_map_return_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_return_module (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_map_break_module
        --- PASS: TestCountsAreRecorded_async_iterable_map_return_function (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_map_break_function
        --- PASS: TestCountsAreRecorded_async_iterable_string_break_function (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_nested_maps_module
        --- PASS: TestCountsAreRecorded_async_iterable_map_break_function (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_nested_maps_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_break_module (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_map_rehash_module
        --- PASS: TestCountsAreRecorded_async_iterable_map_rehash_module (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_map_rehash_function
        --- PASS: TestCountsAreRecorded_async_iterable_nested_maps_function (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_map_readd_module
        --- PASS: TestCountsAreRecorded_async_iterable_string_break_module (0.02s)
        === CONT  TestCountsAreRecorded_async_iterable_map_readd_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_readd_module (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_map_clear_module
        --- PASS: TestCountsAreRecorded_async_iterable_map_clear_module (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_map_clear_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_readd_function (0.02s)
        === CONT  TestCountsAreRecorded_async_iterable_set_entries_module
        --- PASS: TestCountsAreRecorded_async_iterable_nested_maps_module (0.03s)
        --- PASS: TestCountsAreRecorded_async_iterable_map_clear_function (0.03s)
        --- PASS: TestCountsAreRecorded_async_iterable_set_entries_module (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_set_values_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_rehash_function (0.03s)
        --- PASS: TestCountsAreRecorded_async_iterable_set_values_function (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_map_values_module
        === CONT  TestCountsAreRecorded_async_iterable_set_values_module
        === CONT  TestCountsAreRecorded_async_iterable_map_values_function
        --- PASS: TestCountsAreRecorded_async_iterable_set_values_module (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_map_keys_module
        --- PASS: TestCountsAreRecorded_async_iterable_map_values_function (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_map_keys_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_values_module (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_set_entries_function
        === CONT  TestCountsAreRecorded_async_iterable_map_entries_module
        --- PASS: TestCountsAreRecorded_async_iterable_set_entries_function (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_string_surrogates_module
        --- PASS: TestCountsAreRecorded_async_iterable_map_entries_module (0.01s)
        === CONT  TestCountsAreRecorded_async_iterable_map_entries_function
        --- PASS: TestCountsAreRecorded_async_iterable_map_entries_function (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_string_surrogates_function
        --- PASS: TestCountsAreRecorded_async_iterable_string_surrogates_module (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_string_emoji_module
        --- PASS: TestCountsAreRecorded_async_iterable_string_surrogates_function (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_string_emoji_function
        --- PASS: TestCountsAreRecorded_async_iterable_string_emoji_function (0.00s)
        === CONT  TestCountsAreRecorded_async_iterable_string_ascii_module
        --- PASS: TestCountsAreRecorded_async_iterable_string_ascii_module (0.00s)
        === CONT  TestCountsAreRecorded_async_for_of_collections_mutation
        --- PASS: TestCountsAreRecorded_async_iterable_string_emoji_module (0.00s)
        === CONT  TestCountsAreRecorded_async_for_of_map
        --- PASS: TestCountsAreRecorded_async_iterable_map_keys_function (0.03s)
        === CONT  TestCountsAreRecorded_async_iterable_string_ascii_function
        --- PASS: TestCountsAreRecorded_async_for_of_map (0.01s)
        === CONT  TestCountsAreRecorded_async_gap_finally_return_admission
        --- PASS: TestCountsAreRecorded_async_iterable_string_ascii_function (0.00s)
        === CONT  TestCountsAreRecorded_async_gap_finally_admission
        --- PASS: TestCountsAreRecorded_async_for_of_collections_mutation (0.00s)
        === CONT  TestCountsAreRecorded_async_for_of_string
        --- PASS: TestCountsAreRecorded_async_gap_finally_admission (0.00s)
        === CONT  TestCountsAreRecorded_async_gap_catch_admission
        --- PASS: TestCountsAreRecorded_async_for_of_string (0.01s)
        === CONT  TestCountsAreRecorded_async_gap_label_admission
        --- PASS: TestCountsAreRecorded_async_gap_finally_return_admission (0.01s)
        === CONT  TestCountsAreRecorded_async_finally_snapshot
        --- PASS: TestCountsAreRecorded_async_gap_label_admission (0.00s)
        === CONT  TestCountsAreRecorded_async_finally_void_return
        --- PASS: TestCountsAreRecorded_async_gap_catch_admission (0.01s)
        === CONT  TestCountsAreRecorded_async_finally_continue
        --- PASS: TestCountsAreRecorded_async_iterable_map_keys_module (0.04s)
        --- PASS: TestCountsAreRecorded_async_finally_snapshot (0.01s)
        === CONT  TestCountsAreRecorded_async_finally_throw
        --- PASS: TestCountsAreRecorded_async_finally_void_return (0.01s)
        === CONT  TestCountsAreRecorded_async_finally_return
        --- PASS: TestCountsAreRecorded_async_finally_return (0.00s)
        === CONT  TestCountsAreRecorded_label_block_outer_continue_async
        --- PASS: TestCountsAreRecorded_async_finally_throw (0.00s)
        === CONT  TestCountsAreRecorded_label_block_switch_break_async
        --- PASS: TestCountsAreRecorded_label_block_outer_continue_async (0.00s)
        === CONT  TestCountsAreRecorded_label_block_continue_async
        === CONT  TestCountsAreRecorded_async_finally_break
        --- PASS: TestCountsAreRecorded_label_block_switch_break_async (0.02s)
        === CONT  TestCountsAreRecorded_label_block_break_async
        --- PASS: TestCountsAreRecorded_async_finally_break (0.00s)
        === CONT  TestCountsAreRecorded_label_block_outer_continue_sync
        --- PASS: TestCountsAreRecorded_label_block_continue_async (0.02s)
        === CONT  TestCountsAreRecorded_021
        --- PASS: TestCountsAreRecorded_async_finally_continue (0.03s)
        === CONT  TestCountsAreRecorded_async_counts_captured
        --- PASS: TestCountsAreRecorded_async_counts_captured (0.00s)
        === CONT  TestCountsAreRecorded_async_counts_reassigned
        --- PASS: TestCountsAreRecorded_async_counts_reassigned (0.00s)
        === CONT  TestCountsAreRecorded_async_counts_stable
        --- PASS: TestCountsAreRecorded_021 (0.02s)
        === CONT  TestCountsAreRecorded_async_counts_payload
        --- PASS: TestCountsAreRecorded_async_counts_payload (0.00s)
        === CONT  TestCountsAreRecorded_read_after_print
        --- PASS: TestCountsAreRecorded_read_after_print (0.01s)
        === CONT  TestCountsAreRecorded_panic_edges
        --- PASS: TestCountsAreRecorded_async_counts_stable (0.03s)
        === CONT  TestCountsAreRecorded_stderr_only
        --- PASS: TestCountsAreRecorded_label_block_outer_continue_sync (0.03s)
        --- PASS: TestCountsAreRecorded_label_block_break_async (0.04s)
        === CONT  TestCountsAreRecorded_little_streams
        --- PASS: TestCountsAreRecorded_stderr_only (0.01s)
        === CONT  TestCountsAreRecorded_031
        --- PASS: TestCountsAreRecorded_little_streams (0.00s)
        === CONT  TestCountsAreRecorded_030
        === CONT  TestCountsAreRecorded_lots_streams
        --- PASS: TestCountsAreRecorded_lots_streams (0.00s)
        === CONT  TestCountsAreRecorded_029
        --- PASS: TestCountsAreRecorded_030 (0.01s)
        === CONT  TestCountsAreRecorded_028
        --- PASS: TestCountsAreRecorded_031 (0.01s)
        === CONT  TestCountsAreRecorded_027
        --- PASS: TestCountsAreRecorded_029 (0.01s)
        === CONT  TestCountsAreRecorded_026
        --- PASS: TestCountsAreRecorded_panic_edges (0.02s)
        === CONT  TestCountsAreRecorded_025
        --- PASS: TestCountsAreRecorded_025 (0.00s)
        --- PASS: TestCountsAreRecorded_026 (0.01s)
        === CONT  TestCountsAreRecorded_024
        --- PASS: TestCountsAreRecorded_027 (0.01s)
        === CONT  TestCountsAreRecorded_022
        === CONT  TestCountsAreRecorded_023
        --- PASS: TestCountsAreRecorded_024 (0.00s)
        === CONT  TestCountsAreRecorded_typed_arrays_integer
        --- PASS: TestCountsAreRecorded_023 (0.00s)
        === CONT  TestCountsAreRecorded_label_block_break_sync
        --- PASS: TestCountsAreRecorded_022 (0.00s)
        === CONT  TestCountsAreRecorded_async_iterables_cells_labels
        --- PASS: TestCountsAreRecorded_028 (0.01s)
        === CONT  TestCountsAreRecorded_label_block_edges
        --- PASS: TestCountsAreRecorded_label_block_break_sync (0.01s)
        === CONT  TestCountsAreRecorded_async_labels_block_finally
        --- PASS: TestCountsAreRecorded_async_iterables_cells_labels (0.01s)
        === CONT  TestCountsAreRecorded_async_labels_finally
        --- PASS: TestCountsAreRecorded_async_labels_block_finally (0.00s)
        === CONT  TestCountsAreRecorded_async_labels
        --- PASS: TestCountsAreRecorded_async_labels (0.00s)
        === CONT  TestCountsAreRecorded_async_finally_completion
        --- PASS: TestCountsAreRecorded_async_labels_finally (0.00s)
        === CONT  TestCountsAreRecorded_async_loop_body_capture
        --- PASS: TestCountsAreRecorded_async_finally_completion (0.00s)
        === CONT  TestCountsAreRecorded_array
        --- PASS: TestCountsAreRecorded_async_loop_body_capture (0.01s)
        === CONT  TestCountsAreRecorded_flat
        --- PASS: TestCountsAreRecorded_flat (0.00s)
        === CONT  TestCountsAreRecorded_release_fast_shared
        --- PASS: TestCountsAreRecorded_label_block_edges (0.02s)
        === CONT  TestCountsAreRecorded_release_fast_graph
        --- PASS: TestCountsAreRecorded_release_fast_shared (0.01s)
        === CONT  TestCountsAreRecorded_counters_square_jump
        --- PASS: TestCountsAreRecorded_release_fast_graph (0.00s)
        === CONT  TestCountsAreRecorded_counters_shapes
        --- PASS: TestCountsAreRecorded_counters_shapes (0.00s)
        === CONT  TestCountsAreRecorded_typed_arrays_sort_float64array
        --- PASS: TestCountsAreRecorded_counters_square_jump (0.00s)
        === CONT  TestCountsAreRecorded_typed_arrays_sort_int32array
        --- PASS: TestCountsAreRecorded_typed_arrays_sort_int32array (0.00s)
        === CONT  TestCountsAreRecorded_typed_arrays_sort_uint16array
        --- PASS: TestCountsAreRecorded_array (0.02s)
        === CONT  TestCountsAreRecorded_typed_arrays_sort_uint8array
        --- PASS: TestCountsAreRecorded_typed_arrays_sort_uint16array (0.00s)
        --- PASS: TestCountsAreRecorded_typed_arrays_sort_float64array (0.01s)
        === CONT  TestCountsAreRecorded_async_frame_capture_safe
        === CONT  TestCountsAreRecorded_typed_arrays_integer_stop
        --- PASS: TestCountsAreRecorded_typed_arrays_sort_uint8array (0.03s)
        === CONT  TestCountsAreRecorded_async_void_value
        --- PASS: TestCountsAreRecorded_typed_arrays_integer_stop (0.00s)
        === CONT  TestCountsAreRecorded_async_snapshots
        --- PASS: TestCountsAreRecorded_async_void_value (0.00s)
        === CONT  TestCountsAreRecorded_async_receiver
        --- PASS: TestCountsAreRecorded_typed_arrays_integer (0.06s)
        === CONT  TestCountsAreRecorded_async_promise_expression
        --- PASS: TestCountsAreRecorded_async_frame_capture_safe (0.03s)
        === CONT  TestCountsAreRecorded_async_operand_order
        --- PASS: TestCountsAreRecorded_async_receiver (0.00s)
        === CONT  TestCountsAreRecorded_async_mutation
        --- PASS: TestCountsAreRecorded_async_mutation (0.00s)
        --- PASS: TestCountsAreRecorded_async_promise_expression (0.01s)
        === CONT  TestCountsAreRecorded_async_methods
        --- PASS: TestCountsAreRecorded_async_operand_order (0.01s)
        === CONT  TestCountsAreRecorded_async_fresh_holder
        --- PASS: TestCountsAreRecorded_async_methods (0.00s)
        === CONT  TestCountsAreRecorded_async_catches
        --- PASS: TestCountsAreRecorded_async_fresh_holder (0.00s)
        === CONT  TestCountsAreRecorded_async_escaped_capture
        --- PASS: TestCountsAreRecorded_async_catches (0.00s)
        === CONT  TestCountsAreRecorded_async_control
        === CONT  TestCountsAreRecorded_async_live_values
        --- PASS: TestCountsAreRecorded_async_escaped_capture (0.01s)
        === CONT  TestCountsAreRecorded_async_conditions
        --- PASS: TestCountsAreRecorded_async_snapshots (0.02s)
        === CONT  TestCountsAreRecorded_async_closures
        --- PASS: TestCountsAreRecorded_async_control (0.01s)
        === CONT  TestCountsAreRecorded_011
        --- PASS: TestCountsAreRecorded_async_conditions (0.01s)
        --- PASS: TestCountsAreRecorded_async_live_values (0.01s)
        --- PASS: TestCountsAreRecorded_async_closures (0.01s)
        === CONT  TestCountsAreRecorded_020
        === CONT  TestCountsAreRecorded_018
        === CONT  TestCountsAreRecorded_019
        --- PASS: TestCountsAreRecorded_011 (0.00s)
        === CONT  TestCountsAreRecorded_017
        --- PASS: TestCountsAreRecorded_020 (0.00s)
        === CONT  TestCountsAreRecorded_016
        --- PASS: TestCountsAreRecorded_018 (0.01s)
        === CONT  TestCountsAreRecorded_015
        --- PASS: TestCountsAreRecorded_016 (0.01s)
        === CONT  TestCountsAreRecorded_014
        --- PASS: TestCountsAreRecorded_019 (0.01s)
        === CONT  TestCountsAreRecorded_013
        --- PASS: TestCountsAreRecorded_017 (0.01s)
        === CONT  TestCountsAreRecorded_012
        --- PASS: TestCountsAreRecorded_015 (0.00s)
        === CONT  TestCountsAreRecorded_async_counts_root_error
        --- PASS: TestCountsAreRecorded_013 (0.00s)
        === CONT  TestCountsAreRecorded_async_catch_return
        --- PASS: TestCountsAreRecorded_async_counts_root_error (0.00s)
        === CONT  TestCountsAreRecorded_006
        --- PASS: TestCountsAreRecorded_012 (0.00s)
        === CONT  TestCountsAreRecorded_010
        --- PASS: TestCountsAreRecorded_async_catch_return (0.01s)
        === CONT  TestCountsAreRecorded_009
        --- PASS: TestCountsAreRecorded_010 (0.01s)
        === CONT  TestCountsAreRecorded_008
        --- PASS: TestCountsAreRecorded_006 (0.01s)
        === CONT  TestCountsAreRecorded_007
        --- PASS: TestCountsAreRecorded_007 (0.00s)
        === CONT  TestCountsAreRecorded_003
        --- PASS: TestCountsAreRecorded_009 (0.01s)
        --- PASS: TestCountsAreRecorded_008 (0.01s)
        === CONT  TestCountsAreRecorded_005
        === CONT  TestCountsAreRecorded_004
        === CONT  TestCountsAreRecorded_async_counts_thrown
        === CONT  TestCountsAreRecorded_002
        --- PASS: TestCountsAreRecorded_014 (0.02s)
        --- PASS: TestCountsAreRecorded_004 (0.00s)
        --- PASS: TestCountsAreRecorded_async_counts_thrown (0.01s)
        === CONT  TestCountsAreRecorded_async_counts_parameter
        --- PASS: TestCountsAreRecorded_002 (0.00s)
        === CONT  TestCountsAreRecorded_001
        --- PASS: TestCountsAreRecorded_async_counts_parameter (0.00s)
        --- PASS: TestCountsAreRecorded_001 (0.00s)
        --- PASS: TestCountsAreRecorded_005 (0.01s)
        === NAME  TestCountsAreRecorded_003
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
            counts_test.go:191: gate cache native miss
--- FAIL: TestPredicateCountsRegenerateDeterministically (1.60s)
    count_files_test.go:197: regenerated counts differ in testdata/counts/predicates/internal/lower/testdata/predicates/overload_some_objects.a.counts:
        recorded: | internal/lower/testdata/predicates/overload_some_objects.a | 1 | 0 | 2 | 0 |
        regenerated: | internal/lower/testdata/predicates/overload_some_objects.a | 1 | 0 | 1 | 1 |
    count_files_test.go:197: regenerated counts differ in testdata/counts/predicates/internal/lower/testdata/predicates/overload_some_true_only.a.counts:
        recorded: | internal/lower/testdata/predicates/overload_some_true_only.a | 1 | 0 | 2 | 0 |
        regenerated: | internal/lower/testdata/predicates/overload_some_true_only.a | 1 | 0 | 1 | 1 |
    count_files_test.go:197: regenerated counts differ in testdata/counts/predicates/internal/lower/testdata/predicates/overload_checked.a.counts:
        recorded: | internal/lower/testdata/predicates/overload_checked.a | 1 | 0 | 2 | 0 |
        regenerated: | internal/lower/testdata/predicates/overload_checked.a | 1 | 0 | 1 | 1 |
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
native: unit main.c compiles uncached: main.c holds a directive past its first line
--- FAIL: TestPredicateFalseRead (0.01s)
    --- FAIL: TestPredicateFalseRead/again (0.37s)
        predicate_false_read_test.go:40: gate cache node miss
        predicate_false_read_test.go:46: /workspace/adamic/internal/oracle/testdata/predicate_false_read/again.ts:11:19: stage 0 can't lower an array of unknown with erased element storage (retain its declared element type before reading elements) yet
    --- FAIL: TestPredicateFalseRead/spread (0.28s)
        predicate_false_read_test.go:40: gate cache node miss
        predicate_false_read_test.go:46: /workspace/adamic/internal/oracle/testdata/predicate_false_read/spread.ts:24:17: stage 0 can't lower an array of never yet
    --- FAIL: TestPredicateFalseRead/getter (0.41s)
        predicate_false_read_test.go:40: gate cache node miss
        predicate_false_read_test.go:46: /workspace/adamic/internal/oracle/testdata/predicate_false_read/getter.ts:37:17: stage 0 can't lower a read narrowed to undefined of a property a call may write yet
FAIL
gate cache: native hits=3072 misses=6584
gate cache: node hits=1745 misses=6522
gate cache: probe hits=0 misses=40
FAIL	github.com/system-inc/adamic/internal/oracle	3729.942s
FAIL
````

## Admission delta

````sh
source /tmp/adamic-proof/gate-env.sh
python3 cloud/admission-corpus/manifest.py --sha 81525c8b690b9af1662d8275b059eca547faf86c > /tmp/adamic-proof/manifest.json
go run ./cmd/adamic-admission-delta --base 16429a30452d027809f0ea85bc395479423e493e --head 81525c8b690b9af1662d8275b059eca547faf86c --manifest /tmp/adamic-proof/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --workers 2 --json > /tmp/adamic-proof/admission.json 2> /tmp/adamic-proof/admission.stderr
````

`admission exit=0 wall_seconds=1397`

````text
witnesses=2 fixtures=4276 gaps=105 review=20 fuzz=0 stage3=821
````

````text
negative witnesses: 50
ruled run-time stops: 52
ruled backend divergences: 0
runtime timeouts: 0 runs timed out, 0 retried alone, 0 timed out again
changed programs: 10
````

````json
{
  "negative_witness_count": 50,
  "accepted_witnesses": 50,
  "node_agreements": 13,
  "runtime_notyet_count": 52,
  "ruled_runtime_stops": 52,
  "ruled_runtime_stop_paths": [
    "internal/oracle/testdata/intl_collator/in_compare.a",
    "internal/oracle/testdata/intl_collator/locale_en.a",
    "internal/oracle/testdata/intl_collator/locale_malformed.a",
    "internal/oracle/testdata/intl_collator/locale_variable.a",
    "internal/oracle/testdata/intl_collator/own_keys.a",
    "internal/oracle/testdata/intl_collator/tsc_shape.a",
    "internal/oracle/testdata/intl_collator/user_and_host.a",
    "internal/oracle/testdata/intl_collator_compare_value/p01_tsc_shape.a",
    "internal/oracle/testdata/intl_collator_compare_value/p01_tsc_shape.ts",
    "internal/oracle/testdata/reflect_get/stop_error_stack.a",
    "internal/oracle/testdata/reflect_get/stop_function_name.a",
    "internal/oracle/testdata/reflect_get/stop_hasownproperty_host.a",
    "internal/oracle/testdata/reflect_get/stop_host_collator.a",
    "internal/oracle/testdata/reflect_get/stop_host_dirent.a",
    "internal/oracle/testdata/reflect_get/stop_host_error_prototype.a",
    "internal/oracle/testdata/reflect_get/stop_host_fs_error_prototype.a",
    "internal/oracle/testdata/reflect_get/stop_host_groups.a",
    "internal/oracle/testdata/reflect_get/stop_host_handle.a",
    "internal/oracle/testdata/reflect_get/stop_host_iterator.a",
    "internal/oracle/testdata/reflect_get/stop_host_iterator_next.a",
    "internal/oracle/testdata/reflect_get/stop_host_memory_field.a",
    "internal/oracle/testdata/reflect_get/stop_host_memory_heap.a",
    "internal/oracle/testdata/reflect_get/stop_host_performance_mark.a",
    "internal/oracle/testdata/reflect_get/stop_host_stats_ctime.a",
    "internal/oracle/testdata/reflect_get/stop_host_stats_date.a",
    "internal/oracle/testdata/reflect_get/stop_host_stats_field.a",
    "internal/oracle/testdata/reflect_get/stop_host_timeout.a",
    "internal/oracle/testdata/reflect_get/stop_in_host_error_code.a",
    "internal/oracle/testdata/reflect_get/stop_in_host_internal.a",
    "internal/oracle/testdata/reflect_get/stop_in_host_iterator.a",
    "internal/oracle/testdata/reflect_get/stop_in_host_literal.a",
    "internal/oracle/testdata/reflect_get/stop_in_host_method.a",
    "internal/oracle/testdata/reflect_get/stop_keys_host.a",
    "internal/oracle/testdata/reflect_get/stop_keys_iterator.a",
    "internal/oracle/testdata/reflect_get/stop_keys_memory.a",
    "internal/oracle/testdata/reflect_get/stop_literal_getter.a",
    "internal/oracle/testdata/reflect_get/stop_method.a",
    "internal/oracle/testdata/reflect_get/stop_primitive.a",
    "internal/oracle/testdata/reflect_get/stop_private.a",
    "internal/oracle/testdata/reflect_get/stop_proto.a",
    "internal/oracle/testdata/reflect_get/stop_prototype_class.a",
    "internal/oracle/testdata/reflect_get/stop_string_host.a",
    "internal/oracle/testdata/timers_source/stop_foreign_generic.a",
    "internal/oracle/testdata/timers_source/stop_foreign_handle.a",
    "internal/oracle/testdata/timers_source/stop_foreign_intersection.a",
    "internal/oracle/testdata/timers_source/stop_foreign_record.a",
    "internal/oracle/testdata/to_primitive/stop_boolean_array.a",
    "internal/oracle/testdata/to_primitive/stop_record_function.a",
    "internal/oracle/testdata/to_primitive/stop_scalar_array.a",
    "internal/oracle/testdata/to_primitive/stop_static_class.a",
    "internal/oracle/testdata/to_primitive/stop_undefined_array.a",
    "internal/oracle/testdata/to_primitive/stop_union_numbers.a"
  ],
  "stale_runtime_stops": [],
  "pending_runtime_stops": [],
  "backend_divergence_count": 12,
  "ruled_backend_divergences": 0,
  "ruled_backend_divergence_paths": [],
  "corpus_filter": "",
  "shard": "",
  "complete_corpus": true,
  "phase_seconds": {
    "base_build": 51.918775897,
    "checkout": 17.131211982,
    "checkout_and_provision": 19.552344709,
    "classification": 782.387733596,
    "head_build": 38.27491183,
    "manifest_and_verification": 35.807798386,
    "provision": 2.421089722,
    "runtime": 458.206662483,
    "total": 1386.166250173
  },
  "workers": 2,
  "classification": "lowering-only",
  "diff_count": 10,
  "diff_ts_unmatched": [],
  "compile_timeout_seconds": 45,
  "runtime_timeout_seconds": 10,
  "base": "16429a30452d027809f0ea85bc395479423e493e",
  "head": "81525c8b690b9af1662d8275b059eca547faf86c",
  "generator_revision": "81525c8b690b9af1662d8275b059eca547faf86c",
  "generator_blob": "09d061a03b1e016387e975d25ee0cd5b75ac6c1f",
  "manifest_blob": "9fcce57fd7e717b686bf75537c21500906646d83",
  "admitted": 0,
  "verdict": "pass",
  "void": false,
  "sampling_seed": "4d8df2ef5a9c1434364ee2946a53b8ae16df50ef81842b73370922024b590e05",
  "seed_given": false,
  "sampling_size": 115,
  "omitted": 0,
  "budget_seconds": 0,
  "budget_used_seconds": 458.206662483,
  "timed_out_runs": 0,
  "retried_runs": 0,
  "still_timed_out_runs": 0,
  "corpus_sizes": {
    "diff": 10,
    "witnesses": 2,
    "fixtures": 4276,
    "gaps": 105,
    "review": 20,
    "fuzz": 0,
    "stage3": 821
  },
  "total_classified": 5224,
  "newly_refused": 0,
  "classification_counts": {
    "refused-by-both": 1948,
    "accepted-by-both": 3276
  }
}
````

### Void reasons

````json
{
  "gate": null,
  "programs": []
}
````

### Every classification stderr difference

0 programs have differing classification stderr. There are no differing texts to report.

### Timeout reruns

No admission timeouts occurred: every base_compile/compile observation has no timeout; runtime timed_out_runs=0, retried_runs=0, still_timed_out_runs=0, and no program has a timed_out record. There were therefore no admission timeouts to rerun at either SHA. The separate full-oracle 90-second planted-case probe failure is retained and marked in the full-suite inventory above.

## Report-only delivery

All temporary compiler mutations were restored, with clean diff checks recorded above. The report branch is `proof/lower-39w-1-report`, created from the requested detached HEAD. Only `proof/lower-39w-1/report.md` is staged and committed. The only authorized push target is `origin proof/lower-39w-1-report`. The final chat message records the resulting commit and push result.

