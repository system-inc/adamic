Rebased twelve completed default-option ports onto lint-area d3a37422c, containing main b6b1538b0.
Tested code 18808598fb96bf7ab5768261f9bb41fdf7c8c741; publication stays on codex/typeaware-wave-27.
Four rebuilt Go oracle gates, sanitizers, handles, regressions, vet and focused typeof oracle PASS.
Thirteen lint rule mutants, five bridge fact mutants and four typeof mutation groups are caught by external comparisons.
No unclaimed rules remain; dynamic RegExp, parked React analysis, options/shared certification and full external gate remain unresolved.

# Base and commands

Origin/area/stage1-lint is d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898,
containing main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06. Incoming changes preserve
null and lookup presence in typeof. All owned patches rebased without conflict.
Compiler and all owned native/oracle artifacts were rebuilt for this compiler
change. No shared or protected source was edited this turn, and incoming changes
are retained. Only codex/typeaware-wave-27 is pushed.

Shells source /workspace/adamic-tools/env.sh, TMPDIR=/tmp/adamic-gate. Artifact
and frozen corpus environments remain as in ../landing_c019/REPORT.md. All test
output goes directly to retained wave-27-d3-*.log files, never piped. Only named
obsolete scratch archives were deleted for space; current artifacts and evidence
remain. Original setup passed ready 0s/warm cache 116s/total 116s; nproc=5.

```
go build -o /workspace/wave-27-scratch/f801-third/adamic ./cmd/adamic
go run ./cmd/lint-registry
go test -v -count=1 -timeout 15m ./stage1/cohere/typeaware -run '^TestWave27AgreementAndMutants$'
python3 stage1/cohere/typeaware/wave_27_next/validate.py
python3 stage1/cohere/typeaware/wave_27_third/validate.py
python3 stage1/cohere/typeaware/wave_27_fifth/validate.py
go test -v -count=1 -timeout 15m ./bridge/tsgo/checker ./stage1/cohere/typeaware -run '^(TestCoverageCheckerQuestions|TestFactEncoding|TestShapeAndNameFacts|TestExactIndexMatchesCompilerNodes|TestAdamicRootKeepsConfigDeclarations|TestFactsDecoderGuards|TestInspectRequestRefusals|TestPinnedTypeFlags)$'
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware
go test -v -count=1 -timeout 15m ./internal/oracle -run '^(TestTypeOfConstructorMutant|TestTypeOfStringLiteralMutant|TestTypeOfNullMutant|TestTypeOfNullSlotPresenceMutant)$'
```

First gate PASS in 195.519s. Three Python validators print PASS. Focused checker
and type-aware regressions PASS in 0.230s/98.733s; vet exits 0 with empty output.
Focused typeof oracle PASS in 29.020s with normal worker cache (native hits=5,
misses=5; Node hits=1, misses=7). Controls and compiler/repository inputs are
supplied in the owned gates. No correctness check was skipped, relaxed or deleted
in these selected gates. The full repository and full 17-check external suite
were not run and are not claimed passing.

# Exact agreement, mutants and handles

Controls match full finding ranges, ids, messages, fixes and ordered suggestion
edits against production Go cohere: first 105 files/39 findings/17177 bytes;
next 93/101/62201; third 605/437/225707; fifth 186/108/64739. Third retains 34
Go-parser exclusions from 639 candidates; fifth 24 from 210, not passing cases.
Every batch matches 287 frozen repository files (18485 bytes) and 77 TypeScript
compiler files (5934 bytes), zero findings. Controls and both corpora agree
normally and with Go-ASan/native--sanitize; sanitizer stderr is empty. Fifth
checks 351 internal kind constants and three named listener declarations.

All thirteen lint mutants compile and exit 0 with empty stderr; only complete
Go byte comparison catches ORM nullable 57, Serializable optional 5466, Verify
array 8264, pending output 80, race ownership 52234, blocking early return 581,
Effect regex 28395, rest declaration 54, Hook message 4962, regex arity 112016,
symbol ambient 39039, typeof legal-string 46485, await bypass 33546. Numbers are
first differing bytes. Five normally exiting bridge fact mutants are caught at
56 (declaration file), 1318 (CFG reachability), 12452 (never flags), 42069
(type-only import), 61584 (declaration kind). Released checker queries reject
stale handles with exact panic 70. Retained-registry mutants exit 0 and fail the
required refusal check. Focused regressions cover nineteen malformed-wire guards,
wrong-kind/unknown-question mutants and pinned flags.

Incoming typeof mutation groups are independently caught by Node: constructor
classified as object instead of function; string literal classified undefined;
null collapsed to undefined across five fixtures (3/9/3/11/1 restored observations);
missing slot treated as present. Exact Node/mutant output is in wave-27-d3-typeof.log.
These proofs supplement the Go cohere rule oracle; they do not certify emitted
JavaScript for the owned lint rules.

# Timing and limits

Three alternating whole-process samples retain identical streams. Gates overlapped,
so these are contended measurements, not speedup claims:
fifth repository native 588.946ms / Go 254.736ms (2.312x);
fifth compiler native 3994.100ms / Go 427.036ms (9.353x).

The rebuilt compiler still rejects additional_hooks_pattern.a:4:23 with exit 1:
stage 0 can't lower RegExp with a nonconstant pattern yet. The required isolated
new RegExp(pattern, 'u') helper is present, but runtime additionalHooks integration
remains blocked on shared lowering. Three React HIR/SSA/capture/post-dominance
claims remain parked. Older nine ports await handed-node migration. Nondefault
options, shared registration/emitted-JavaScript certification, inherited gates,
full repository gate and full external correctness suite remain untested.
The new shared nil-options adapter guard is preserved; owned differential drivers
cover default options only and do not claim testing that shared guard.

Origin audit: 659 refs, 197 ranked rules, 33 Markdown claim records, zero unclaimed
rules. No further claim is taken. selection.json retains complete evidence.
