RawInputGap defended by D1 in the bounded matrix, with native fileStatus's sole package caller confirmed.
InterfaceDefaultGap not defended after D2-D4; TypeMethodGap not defended after D5-D7.
Seven standalone production diffs apply to base; tests and oracles are untouched.

Base: 955e3eb92b9cd04aca420974d1006048b4615d51. See code-and-oracle.md for independent oracles, per-row differences, limitations and owner findings.

```json
[
  {
    "test": "TestRawInputGap",
    "package": "stage1/cohere/estree",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestJSXAgreement family"
    ],
    "defense": "defended",
    "unique_mutant": "D1 internal/native/runtime/directory.c:170",
    "attempts": [
      {
        "mutant": "D1",
        "file_line": "internal/native/runtime/directory.c:170",
        "change": "stat size + 1",
        "rows_failed": [
          "TestRawInputGap"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/defend-estree/cache/D1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestRawInputGap|TestInterfaceDefaultGap|TestInterfaceTypeMethodGap|TestPostfixValueGap|TestMethodReplacementGap|TestJSXAgreement(_[0-9]{3})?)$' > D1.log 2>&1 ; gaps_test.go:86: reader changed: \"9:8://\ufffd\\nx;\\n\" != \"8:8://\ufffd\\nx;\\n\"",
    "bounded": true,
    "matrix_rows": [
      "TestInterfaceDefaultGap",
      "TestInterfaceTypeMethodGap",
      "TestJSXAgreement",
      "TestJSXAgreement_000",
      "TestJSXAgreement_001",
      "TestJSXAgreement_002",
      "TestJSXAgreement_003",
      "TestJSXAgreement_004",
      "TestJSXAgreement_005",
      "TestJSXAgreement_006",
      "TestJSXAgreement_007",
      "TestJSXAgreement_008",
      "TestJSXAgreement_009",
      "TestJSXAgreement_010",
      "TestJSXAgreement_011",
      "TestJSXAgreement_012",
      "TestJSXAgreement_013",
      "TestJSXAgreement_014",
      "TestJSXAgreement_015",
      "TestJSXAgreement_016",
      "TestJSXAgreement_017",
      "TestJSXAgreement_018",
      "TestJSXAgreement_019",
      "TestJSXAgreement_020",
      "TestJSXAgreement_021",
      "TestJSXAgreement_022",
      "TestJSXAgreement_023",
      "TestJSXAgreement_024",
      "TestJSXAgreement_025",
      "TestJSXAgreement_026",
      "TestJSXAgreement_027",
      "TestJSXAgreement_028",
      "TestJSXAgreement_029",
      "TestJSXAgreement_030",
      "TestJSXAgreement_031",
      "TestMethodReplacementGap",
      "TestPostfixValueGap",
      "TestRawInputGap"
    ]
  },
  {
    "test": "TestInterfaceDefaultGap",
    "package": "stage1/cohere/estree",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestInterfaceTypeMethodGap"
    ],
    "defense": "not defended",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D2",
        "file_line": "internal/lower/class_inheritance.go:761",
        "change": "missing optional callback argument: break -> return to",
        "rows_failed": [
          "TestInterfaceDefaultGap",
          "TestJSXAgreement family"
        ]
      },
      {
        "mutant": "D3",
        "file_line": "internal/lower/iteration_origin.go:95",
        "change": "drop prototype-origin diagnostic assignment",
        "rows_failed": [
          "TestInterfaceDefaultGap",
          "TestInterfaceTypeMethodGap"
        ]
      },
      {
        "mutant": "D4",
        "file_line": "internal/lower/iteration_origin.go:88",
        "change": "class parent == -> !=",
        "rows_failed": [
          "TestInterfaceDefaultGap",
          "TestInterfaceTypeMethodGap"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/defend-estree/cache/D2 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestRawInputGap|TestInterfaceDefaultGap|TestInterfaceTypeMethodGap|TestPostfixValueGap|TestMethodReplacementGap|TestJSXAgreement(_[0-9]{3})?)$' > D2.log 2>&1 ; gaps_test.go:135: recorded lowering gap changed: /workspace/adamic/stage1/cohere/estree/gaps/interfaceDefault.ts:12:20: Adamic 0.1 refuses a value without nominal ancestry seen as (value: number, step?: number) => number; construct that class or a subclass; use an interface for structural values (adamic/nominal-class) ; ADAMIC_BUILD_CACHE_DIR=/tmp/defend-estree/cache/D3 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestRawInputGap|TestInterfaceDefaultGap|TestInterfaceTypeMethodGap|TestPostfixValueGap|TestMethodReplacementGap|TestJSXAgreement(_[0-9]{3})?)$' > D3.log 2>&1 ; gaps_test.go:135: recorded lowering gap changed: <nil> ; ADAMIC_BUILD_CACHE_DIR=/tmp/defend-estree/cache/D4 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestRawInputGap|TestInterfaceDefaultGap|TestInterfaceTypeMethodGap|TestPostfixValueGap|TestMethodReplacementGap|TestJSXAgreement(_[0-9]{3})?)$' > D4.log 2>&1 ; gaps_test.go:135: recorded lowering gap changed: <nil>",
    "bounded": true,
    "matrix_rows": [
      "TestInterfaceDefaultGap",
      "TestInterfaceTypeMethodGap",
      "TestJSXAgreement",
      "TestJSXAgreement_000",
      "TestJSXAgreement_001",
      "TestJSXAgreement_002",
      "TestJSXAgreement_003",
      "TestJSXAgreement_004",
      "TestJSXAgreement_005",
      "TestJSXAgreement_006",
      "TestJSXAgreement_007",
      "TestJSXAgreement_008",
      "TestJSXAgreement_009",
      "TestJSXAgreement_010",
      "TestJSXAgreement_011",
      "TestJSXAgreement_012",
      "TestJSXAgreement_013",
      "TestJSXAgreement_014",
      "TestJSXAgreement_015",
      "TestJSXAgreement_016",
      "TestJSXAgreement_017",
      "TestJSXAgreement_018",
      "TestJSXAgreement_019",
      "TestJSXAgreement_020",
      "TestJSXAgreement_021",
      "TestJSXAgreement_022",
      "TestJSXAgreement_023",
      "TestJSXAgreement_024",
      "TestJSXAgreement_025",
      "TestJSXAgreement_026",
      "TestJSXAgreement_027",
      "TestJSXAgreement_028",
      "TestJSXAgreement_029",
      "TestJSXAgreement_030",
      "TestJSXAgreement_031",
      "TestMethodReplacementGap",
      "TestPostfixValueGap",
      "TestRawInputGap"
    ]
  },
  {
    "test": "TestInterfaceTypeMethodGap",
    "package": "stage1/cohere/estree",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestInterfaceDefaultGap"
    ],
    "defense": "not defended",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D5",
        "file_line": "internal/lower/expression.go:1024",
        "change": "swap conditional branch operands",
        "rows_failed": [
          "TestInterfaceTypeMethodGap",
          "TestJSXAgreement family"
        ]
      },
      {
        "mutant": "D6",
        "file_line": "internal/lower/expression.go:508",
        "change": "Boolean literal TrueKeyword == -> FalseKeyword ==",
        "rows_failed": [
          "TestInterfaceTypeMethodGap",
          "TestJSXAgreement family"
        ]
      },
      {
        "mutant": "D7",
        "file_line": "internal/lower/expression.go:783",
        "change": "PlusToken ir.Add -> ir.Subtract",
        "rows_failed": [
          "TestInterfaceTypeMethodGap",
          "TestJSXAgreement family"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/defend-estree/cache/D5 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestRawInputGap|TestInterfaceDefaultGap|TestInterfaceTypeMethodGap|TestPostfixValueGap|TestMethodReplacementGap|TestJSXAgreement(_[0-9]{3})?)$' > D5.log 2>&1 ; interface_type_test.go:54: emitted default-free control: 0 ; ADAMIC_BUILD_CACHE_DIR=/tmp/defend-estree/cache/D6 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestRawInputGap|TestInterfaceDefaultGap|TestInterfaceTypeMethodGap|TestPostfixValueGap|TestMethodReplacementGap|TestJSXAgreement(_[0-9]{3})?)$' > D6.log 2>&1 ; interface_type_test.go:54: native default-free control: 0 ; ADAMIC_BUILD_CACHE_DIR=/tmp/defend-estree/cache/D7 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestRawInputGap|TestInterfaceDefaultGap|TestInterfaceTypeMethodGap|TestPostfixValueGap|TestMethodReplacementGap|TestJSXAgreement(_[0-9]{3})?)$' > D7.log 2>&1 ; interface_type_test.go:54: emitted default-free control: -1",
    "bounded": true,
    "matrix_rows": [
      "TestInterfaceDefaultGap",
      "TestInterfaceTypeMethodGap",
      "TestJSXAgreement",
      "TestJSXAgreement_000",
      "TestJSXAgreement_001",
      "TestJSXAgreement_002",
      "TestJSXAgreement_003",
      "TestJSXAgreement_004",
      "TestJSXAgreement_005",
      "TestJSXAgreement_006",
      "TestJSXAgreement_007",
      "TestJSXAgreement_008",
      "TestJSXAgreement_009",
      "TestJSXAgreement_010",
      "TestJSXAgreement_011",
      "TestJSXAgreement_012",
      "TestJSXAgreement_013",
      "TestJSXAgreement_014",
      "TestJSXAgreement_015",
      "TestJSXAgreement_016",
      "TestJSXAgreement_017",
      "TestJSXAgreement_018",
      "TestJSXAgreement_019",
      "TestJSXAgreement_020",
      "TestJSXAgreement_021",
      "TestJSXAgreement_022",
      "TestJSXAgreement_023",
      "TestJSXAgreement_024",
      "TestJSXAgreement_025",
      "TestJSXAgreement_026",
      "TestJSXAgreement_027",
      "TestJSXAgreement_028",
      "TestJSXAgreement_029",
      "TestJSXAgreement_030",
      "TestJSXAgreement_031",
      "TestMethodReplacementGap",
      "TestPostfixValueGap",
      "TestRawInputGap"
    ]
  }
]
```

Mutant matrix

| Id | Base file:line | Change | Failed rows (families grouped) | Binary seconds |
| --- | --- | --- | --- | --- |
| D1 | internal/native/runtime/directory.c:170 | stat size + 1 | TestRawInputGap | 46.612 |
| D2 | internal/lower/class_inheritance.go:761 | missing optional callback argument: break -> return to | TestInterfaceDefaultGap, TestJSXAgreement family | 5.438 |
| D3 | internal/lower/iteration_origin.go:95 | drop prototype-origin diagnostic assignment | TestInterfaceDefaultGap, TestInterfaceTypeMethodGap | 33.314 |
| D4 | internal/lower/iteration_origin.go:88 | class parent == -> != | TestInterfaceDefaultGap, TestInterfaceTypeMethodGap | 33.078 |
| D5 | internal/lower/expression.go:1024 | swap conditional branch operands | TestInterfaceTypeMethodGap, TestJSXAgreement family | 53.26 |
| D6 | internal/lower/expression.go:508 | Boolean literal TrueKeyword == -> FalseKeyword == | TestInterfaceTypeMethodGap, TestJSXAgreement family | 90.085 |
| D7 | internal/lower/expression.go:783 | PlusToken ir.Add -> ir.Subtract | TestInterfaceTypeMethodGap, TestJSXAgreement family | 44.447 |

D1: all 37 other selected top-level tests pass. Full passed list is matrix.json/D1/rows_passed. This is 38 current tests, or six rows after grouping JSX. Static fileStatus caller evidence is runtime-callers.txt. Other package rows remain unexecuted by this matrix.

D2: its missing optional callback bound is exclusive relative to the paired TypeMethodGap coverage, but JSX's sourceValidation callback also needs it. D3 and D4: removing the prototype refusal or reversing its class-origin classification makes both negative gap assertions fail with nil. Neither behavior is unique to DefaultGap.

D5-D7: TypeMethodGap's positive control observes swapped conditional output 0, inverted Boolean output 0 and subtraction output -1 respectively. DefaultGap passes those mutants. JSX also catches each, so the positive-control difference is real but no attempted break is uniquely caught. This supersedes the audit's reciprocal one-diagnostic-mutant hint without claiming these tests are interchangeable.

D6 cooked at 90.085 binary seconds. The three target rows were rerun separately; RawInputGap and DefaultGap pass, TypeMethodGap fails. JSX_027 also fails in its individual rerun. matrix.json preserves all unfinished other results as unknown. Five orphaned native D6 subprocesses were stopped after the test binary timed out. No whole-package uniqueness is inferred from that timeout.

Validation: D1's changed C compiled through sanitized native.Build and executed. D2-D7 each passed go vet ./internal/lower/; empty stdout logs are preserved. All seven standalone diffs pass git apply --check after source restoration (apply-check.json). The clean coverage runs pass; whole package exceeds 90 seconds rather than yielding a red assertion baseline. clean-bounded-baseline.log passes all 38 top-level tests without skips.

Costs and unclear parts: Go -coverpkg cannot instrument the native C runtime or generated TypeScript; exclusive C entry evidence is static and semantic. The package baseline required narrowing, and compiler kills outside the six grouped rows are unknown. Audit discovery was a slice of 47 members, whereas current discovery lists 245 tests; the difference is not proof that 198 were added. No requested name vanished. Current JSX members were discovered afresh and all included. The 15 GB /tmp threshold is impossible on its 8.8 GB filesystem; authorized scratch cleanup gave 6.4 GB free. A mistyped IR name caused one unmutated additional baseline run, saved rather than counted as a mutant.

Owner findings: RawInputGap's historic Gap name now covers implemented behavior. DefaultGap promises a recorded gap and asserts its precise refusal, not successful native execution of defaults. TypeMethodGap also exercises a successful default-free callback. No speed or budget promise is left unchecked in these names. Neither not-defended verdict authorizes deletion or weakening; only three aimed attempts per interface row were tried.

Timing: warm setup skipped (0 s); nproc 5; npm ci 0.578 s. Whole baseline 90.037 binary seconds; extra clean bounded baseline 35.554 s. Seven matrix binaries total 306.234 s, including builds performed within the test binary and the cooked D6 budget. Individual D6 reruns and four coverage commands have their own logged binary timings. Total session approximately 26 minutes. Build and assertion time are not separable where native products are compiled inside tests.
Not covered: full-package completion, compiler kills outside the bounded matrix, repo-wide uniqueness, exhaustive input boundaries or native C coverage instrumentation. Production sources are restored. No test edits, main push or pull request.
