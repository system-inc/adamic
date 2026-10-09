Agreement family defended at the released-inspect diagnostic boundary.
Corpus family not defended after three aimed import attempts; two are shared and one survives.
Production sources and tests are restored; bounded evidence and standalone diffs are retained.

[
  {
    "test": "TestVolumeProfileCorpora family",
    "package": "stage1/cohere/typeaware",
    "prior_verdict": "subsumed",
    "prior_subsumed_by": [
      "TestVolumeProfileControls family"
    ],
    "subsumed_by": [
      "TestVolumeAgreementRepository family"
    ],
    "defense": "not defended",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D1",
        "file_line": "stage1/cohere/typeaware/shadow.ts:70",
        "change": "node.kind === 'ImportSpecifier' && node.semantic === '1' -> node.kind === 'ImportSpecifier' && node.semantic === '0'",
        "rows_failed": [
          "TestVolumeAgreementRepository family",
          "TestVolumeProfileCorpora family"
        ]
      },
      {
        "mutant": "D2",
        "file_line": "stage1/cohere/typeaware/shadow.ts:74",
        "change": "return node.semantic !== 'TypeKeyword'; -> return node.semantic !== 'TypeKeyword_';",
        "rows_failed": []
      },
      {
        "mutant": "D3",
        "file_line": "stage1/cohere/typeaware/shadow.ts:66",
        "change": "(binding.flags & 2097152) !== 0 -> (binding.flags & 1048576) !== 0",
        "rows_failed": [
          "TestVolumeAgreementRepository family",
          "TestVolumeProfileCorpora family"
        ]
      }
    ],
    "evidence": "See matrix.json exact per-family timeout 120 go test -json -count=1 -timeout 90s commands. D1 and D3: volume_profile_Corpora_test.go:560 repository mismatch; the other repository family catches both. D2 passes all observed rows; manual-survivor-witness.json shows clean/Go findings 1 versus mutant findings 0.",
    "bounded": true,
    "name_assertion_finding": "The name fits its exact corpus finding-byte comparisons. The logged 60s budget is enforced with a 90s deadline, not a 60s performance assertion."
  },
  {
    "test": "TestVolumeAgreementAndMutants family",
    "package": "stage1/cohere/typeaware",
    "prior_verdict": "subsumed",
    "prior_subsumed_by": [
      "TestVolumeProfileControls family"
    ],
    "subsumed_by": [],
    "defense": "defended",
    "unique_mutant": "D4 bridge/tsgo/archive/main.go:118",
    "attempts": [
      {
        "mutant": "D4",
        "file_line": "bridge/tsgo/archive/main.go:118",
        "change": "return failed(error, C.TSGO_HANDLE, \"invalid or released checker handle\") -> return failed(error, C.TSGO_HANDLE, \"invalid or released checker handle.\")",
        "rows_failed": [
          "TestVolumeAgreementAndMutants family"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/defend-typeaware-volume/cache/D4 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/typeaware/ -run ^TestVolumeAgreementAndMutants_022$; volume_agreement_shards_test.go:377: released program escaped: exit status 70 adamic: panic: invalid or released checker handle. Other reached rows pass: TestInspectRequestRefusals and TestSixRuleAgreementAndMutants_045.",
    "bounded": true
  }
]

Code under test and oracles

See code-and-oracle.md. Semantic corpus checks compare unchanged Go cohere findings byte for byte. The D4 lifecycle oracle is the self-written exact panic diagnostic and code 70. The production bridge guard is code under test; no oracle, harness, test or built-in mutant was changed.

Passing rows and scope

D4 fails only VolumeAgreementAndMutants_022 among the known released-inspect checks. TestInspectRequestRefusals and TestSixRuleAgreementAndMutants_045 pass; full per-test lists are in pass-lists.json. D1-D3 run all 64 repository corpus leaves, all 32 other repository agreement leaves, both controls leaves, agreement leaves 014/022/030 and unions. Family members and exact commands are in scope.json and matrix.json. Results outside these bounded sets are unknown. The source grep found the two other native released-inspect assertions use bytes.Contains, while the agreement leaf compares the entire diagnostic. The legacy exact released.ts check uses tsgoTypeParts, whose separate guard is unchanged.

Standalone mutants

D1 stage1/cohere/typeaware/shadow.ts:70 node.kind === 'ImportSpecifier' && node.semantic === '1' -> node.kind === 'ImportSpecifier' && node.semantic === '0'; failed TestVolumeAgreementRepository family, TestVolumeProfileCorpora family
D2 stage1/cohere/typeaware/shadow.ts:74 return node.semantic !== 'TypeKeyword'; -> return node.semantic !== 'TypeKeyword_';; failed 
D3 stage1/cohere/typeaware/shadow.ts:66 (binding.flags & 2097152) !== 0 -> (binding.flags & 1048576) !== 0; failed TestVolumeAgreementRepository family, TestVolumeProfileCorpora family
D4 bridge/tsgo/archive/main.go:118 return failed(error, C.TSGO_HANDLE, "invalid or released checker handle") -> return failed(error, C.TSGO_HANDLE, "invalid or released checker handle."); failed TestVolumeAgreementAndMutants family

Survivor

D2: manual type-only Point import shadowed by a function type parameter. Clean port and unchanged Go oracle report one finding and match byte for byte; the mutant reports zero. Native products are identified by exact content-addressed paths in manual-survivor-witness.json. This witness is not a matrix kill.

Brief ambiguities, costs and owner findings

1. Starting main is bfe0553300773c0b37db2c10df97adeb909705f8, rather than the old audit base. All 578 top-level names are unchanged from its list.
2. The whole package cooked at 90.114s in unrelated nonparallel archive setup. The first bounded corpus/control baseline also cooked at 90.108s in archive preparation. Individual and separate family baselines passed before mutants.
3. The combined lifecycle catcher baseline passed Inspect but cooked during Six preparation. Six alone then passed at 9.645s. The D4 reached rows were run separately so each stayed below 90s.
4. Go coverage cannot instrument the TypeScript port, C runtime, or bridge archive subprocess. Linked compiler profiles measure cache misses and preparation, not exclusive port branches. Controls and corpora each cover 13 projected Go lines; agreement covers 7637. These are not semantic-exclusive coverage claims.
5. The old audit measured only agreement semantic control leaves 014 and 030. Its whole-family subsumption label omitted the lifecycle and planted checker-witness members. This defense uses a production bridge lifecycle guard, never edits the oracle or witness comparison.
6. No Node/native twin pair is present in these matrix families. Plain and ASan wrappers are grouped inside each family.
7. The prior evidence uses REPORT.txt, not REPORT.md. A first coverage invocation missed sourcing env.sh and called the system Go command; it exited immediately and was corrected. The erroneous commands are retained separately, with no verdict based on them.
8. Ambient global and conditional infer candidates were rejected because their apparent corpus occurrences were absent or merely scanner token strings. An inactive setup was interrupted; no verdict rests on it. Only final D1-D3 import changes count as corpus attempts.
9. Timed-out Go tests left compiler children. Those observed session-owned compiler processes were stopped before narrowing; completed products and normal Go compilation caches allowed the later green baselines.
10. ADAMIC_BUILD_CACHE_DIR is distinct per mutant. Native port products and bridge archives were rebuilt using their normal builders; build phases are recorded in build-times.json. Separate phase durations can overlap and are not summed as elapsed wall time.
11. The first D1 broad matrix cooked with observed corpus failures. Its per-family replays all completed. Unknown cells from that cooked command are not inferred to pass.
12. Compiler corpus leaves 064-217 were not enabled, and bridge planted-witness leaves outside the semantic/lifecycle selection were not replayed. Wider package and repository behavior outside the listed reached matrices remains unknown.
13. Three aimed corpus attempts give no deletion recommendation: D1 and D3 are caught by another family; D2 is a demonstrated survivor in this matrix. The manual fixture is a behavior witness, not a test kill.
14. The corpus name agrees with its assertions, but its 60s budget wording is enforced with 90s. Its exact comparisons cannot protect a type-only-import shadow boundary absent from its chosen inputs.

Timing

Warm setup skipped; nproc=5. npm ci completed. Clean plain controls baseline: 58.233s; released agreement leaf: 21.552s; other repository family: 10.493s. Coverage wall commands: 49.095s, 42.71s, 53.589s. Per-mutant wall and binary logs, cache IDs, apply checks, bridge go vet and native rebuild phases are retained. D4 go vet passes; D1-D3 build and execute native plain/ASan products under the port builders. No main push, PR, test change or production fix.
