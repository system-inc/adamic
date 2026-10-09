u074 completed at origin/main 09fe4b54913753188a9357982bfd47cdf36ef97c; all four names remain in their listed files.
Clean baseline: 1.881 s; final restored run: 1.654 s; nproc: 5; no skips.
Verdicts: one sacred, three subsumed; no whole-row vacuity, but 13 vacuous positive parser cases.
Matrix: complete package, 17 top-level names grouped into 12 rows; 11 menu mutants plus supplemental M6; three probes.
Evidence: test-audit/internal-regexp-parser, review/test-audit/internal-regexp-parser/.

```json
[
  {
    "test": "TestParse",
    "package": "internal/regexp",
    "file": "internal/regexp/parser_test.go",
    "seconds": 0.004,
    "oracle": "Handwritten valid/invalid cases. Only error presence is checked, not AST or diagnostic reason; 13 positive cases pass the empty Parse answer.",
    "oracle_kind": "self",
    "kills": [
      "M3",
      "M5",
      "M9"
    ],
    "unique_kills": [],
    "last_proven_fail": "M9: parser_test.go:27: Parse(\"[^[\\\\q{ab|a}]&&[a]]\",\"v\") error=regexp: cannot negate a class containing strings at byte 18, want valid=true",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestMatcherAgreement family"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 1.504,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestMatcherAgreement family",
      "TestMatcherStepLimit",
      "TestMatcherPropertyProviderStrings",
      "TestMatcherProviderSnapshot",
      "TestMatcherStepLimitBoundary",
      "TestMatcherOct6Mutants",
      "TestMatcherOct6LoopsNode",
      "TestNodeAgreement",
      "TestParse",
      "TestFlags",
      "TestQuantifierBounds",
      "TestSharedCanonicalize"
    ],
    "evidence": "ADAMIC_MUTANT=M9 ADAMIC_BUILD_CACHE_DIR=/tmp/u074/cache/M9 timeout 120 go test -json -count=1 -timeout 90s ./internal/regexp/ -run . > review/test-audit/internal-regexp-parser/M9.log 2>&1; parser_test.go:27: Parse(\"[^[\\\\q{ab|a}]&&[a]]\",\"v\") error=regexp: cannot negate a class containing strings at byte 18, want valid=true",
    "eligible_mutants": 11,
    "subsumption_kills": 3,
    "vacuous_subcases": [
      {
        "pattern": "[^[\\q{ab|a}]&&[a]]",
        "flags": "v"
      },
      {
        "pattern": "[🌍-🌎]",
        "flags": "u"
      },
      {
        "pattern": "[a-🌍]",
        "flags": ""
      },
      {
        "pattern": "[[]",
        "flags": ""
      },
      {
        "pattern": "(?:a|b)+?",
        "flags": "gi"
      },
      {
        "pattern": "(?<word>\\p{Letter}+)\\k<word>",
        "flags": "u"
      },
      {
        "pattern": "(?<=a)b(?<!c)",
        "flags": ""
      },
      {
        "pattern": "[[a-z]&&[^aeiou]]",
        "flags": "v"
      },
      {
        "pattern": "[\\q{ab|cd}]",
        "flags": "v"
      },
      {
        "pattern": "(a)\\1",
        "flags": "u"
      },
      {
        "pattern": "a{2,2}",
        "flags": ""
      },
      {
        "pattern": "\\0\\cA\\p{sc=Latn}",
        "flags": "u"
      },
      {
        "pattern": "\\8",
        "flags": ""
      }
    ],
    "supplemental_kills": [
      "M6"
    ]
  },
  {
    "test": "TestFlags",
    "package": "internal/regexp",
    "file": "internal/regexp/parser_test.go",
    "seconds": 0.004,
    "oracle": "Handwritten invalid flags uu, uv and z. Only error presence is checked, not diagnostic reason.",
    "oracle_kind": "self",
    "kills": [
      "M1",
      "M2"
    ],
    "unique_kills": [],
    "last_proven_fail": "M2: parser_test.go:36: flags \"uv\" accepted",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNodeAgreement"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 0.41,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestMatcherAgreement family",
      "TestMatcherStepLimit",
      "TestMatcherPropertyProviderStrings",
      "TestMatcherProviderSnapshot",
      "TestMatcherStepLimitBoundary",
      "TestMatcherOct6Mutants",
      "TestMatcherOct6LoopsNode",
      "TestNodeAgreement",
      "TestParse",
      "TestFlags",
      "TestQuantifierBounds",
      "TestSharedCanonicalize"
    ],
    "evidence": "ADAMIC_MUTANT=M2 ADAMIC_BUILD_CACHE_DIR=/tmp/u074/cache/M2 timeout 120 go test -json -count=1 -timeout 90s ./internal/regexp/ -run . > review/test-audit/internal-regexp-parser/M2.log 2>&1; parser_test.go:36: flags \"uv\" accepted",
    "eligible_mutants": 11,
    "subsumption_kills": 2
  },
  {
    "test": "TestQuantifierBounds",
    "package": "internal/regexp",
    "file": "internal/regexp/parser_test.go",
    "seconds": 0.003,
    "oracle": "Handwritten AST fields: Min=2, Max=4, Greedy=false for a{2,4}?.",
    "oracle_kind": "self",
    "kills": [
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: parser_test.go:49: quantifier = &regexp.Quantifier{Atom:(*regexp.Character)(0x24facfb56540), Min:2, Max:4, Greedy:true}",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestMatcherOct6LoopsNode"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 0.13,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestMatcherAgreement family",
      "TestMatcherStepLimit",
      "TestMatcherPropertyProviderStrings",
      "TestMatcherProviderSnapshot",
      "TestMatcherStepLimitBoundary",
      "TestMatcherOct6Mutants",
      "TestMatcherOct6LoopsNode",
      "TestNodeAgreement",
      "TestParse",
      "TestFlags",
      "TestQuantifierBounds",
      "TestSharedCanonicalize"
    ],
    "evidence": "ADAMIC_MUTANT=M4 ADAMIC_BUILD_CACHE_DIR=/tmp/u074/cache/M4 timeout 120 go test -json -count=1 -timeout 90s ./internal/regexp/ -run . > review/test-audit/internal-regexp-parser/M4.log 2>&1; parser_test.go:49: quantifier = &regexp.Quantifier{Atom:(*regexp.Character)(0x24facfb56540), Min:2, Max:4, Greedy:true}",
    "eligible_mutants": 11,
    "subsumption_kills": 1
  },
  {
    "test": "TestSharedCanonicalize",
    "package": "internal/regexp",
    "file": "internal/regexp/properties_test.go",
    "seconds": 0.036,
    "oracle": "Internal generated regexp simpleCaseFold and legacyUppercase tables, derived from Unicode 17, decide expected values. Exact numeric agreement across all code points and BMP units; no independent authority value checked. Shared data provenance can conceal common errors.",
    "oracle_kind": "self",
    "kills": [
      "M10",
      "M11",
      "M12"
    ],
    "unique_kills": [
      "M10",
      "M11",
      "M12"
    ],
    "last_proven_fail": "M12: properties_test.go:19: Unicode U+0100: shared=100 local=101",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P2",
      "P3"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestMatcherAgreement family",
      "TestMatcherStepLimit",
      "TestMatcherPropertyProviderStrings",
      "TestMatcherProviderSnapshot",
      "TestMatcherStepLimitBoundary",
      "TestMatcherOct6Mutants",
      "TestMatcherOct6LoopsNode",
      "TestNodeAgreement",
      "TestParse",
      "TestFlags",
      "TestQuantifierBounds",
      "TestSharedCanonicalize"
    ],
    "evidence": "ADAMIC_MUTANT=M12 ADAMIC_BUILD_CACHE_DIR=/tmp/u074/cache/M12 timeout 120 go test -json -count=1 -timeout 90s ./internal/regexp/ -run . > review/test-audit/internal-regexp-parser/M12.log 2>&1; properties_test.go:19: Unicode U+0100: shared=100 local=101",
    "eligible_mutants": 11,
    "subsumption_kills": 3
  }
]
```

Mutants, all locations at starting origin/main. Abbreviations: F = TestMatcherAgreement family, N = TestNodeAgreement, P = TestParse, G = TestFlags, Q = TestQuantifierBounds, S = TestSharedCanonicalize, L = TestMatcherOct6LoopsNode. Full names and raw member results are in matrix.json.

| id | file:line | change | failed production rows |
|---|---|---|---|
| M1 | internal/regexp/parser.go:88 | change constant: stop tracking duplicate flags | N, G |
| M2 | internal/regexp/parser.go:110 | drop the whole u/v exclusion if statement | N, G |
| M3 | internal/regexp/parser.go:325 | off-by-one quantifier range comparison | F, N, P |
| M4 | internal/regexp/parser.go:329 | flip greediness | F, L, Q |
| M5 | internal/regexp/parser.go:561 | flip class range comparison | F, N, P |
| M6 supplemental | internal/regexp/parser.go:254 | supplemental: remove two condition operands, outside the strict menu; excluded from verdicts | F, N, P |
| M7 | internal/regexp/parser.go:927 | change capture count increment constant |  |
| M8 | internal/regexp/parser.go:429 | change control escape modulus constant | F |
| M9 | internal/regexp/parser.go:617 | flip intersection condition | F, P |
| M10 | internal/unicodeproperties/canonicalize.go:15 | off-by-one Unicode search bound | S |
| M11 | internal/unicodeproperties/canonicalize.go:54 | change legacy fold column constant | S |
| M12 | internal/unicodeproperties/canonicalize.go:11 | change Unicode range constant | S |

M4 also breaks a control precondition in TestMatcherOct6Mutants. This witness failure is excluded from kills and subsumption. Scoped rows do not form a family: their assertions differ. The six matcher corpus builders listed in families.json form one comparison family, including its stored-expectation test262 member. Family timing runs all six together, three times. Subsumption is a hint based on P: three kills, G: two kills and Q: one kill, not a deletion recommendation.

Survivor M7: Parse("(a)\\2","u"): accepted=false error=regexp: invalid decimal escape at byte 3; mutant: Parse("(a)\\2","u"): accepted=true error=<nil>. Every package row passes. This is changed, unguarded admission behavior in this matrix, not an equivalent candidate. Witness command: ADAMIC_MUTANT=M7 go run ./review/test-audit/internal-regexp-parser/survivor-witness.go (copy survivor-witness.go.txt first).

Probes: P1 replaces Parse with return nil,nil; P2 and P3 replace the two canonicalization entries with return 0. All intended rows fail. P1 TestQuantifierBounds panics at parser_test.go:47; each intended row was already run separately, so no other result is inferred from the aborted binary. TestParse has 13 positive cases that pass P1 and 15 negative cases that fail; its vacuous_subcases lists every positive case. Probe failures are excluded from mutant kills.

Brief ambiguities, errors and costs:

- The historical commit 8de93800f4 is not current origin/main. Fetch selected 09fe4b5491; no scoped name moved or vanished. Every reported production line and diff uses that starting commit.
- npm ci is required in stage3/api even though these tests execute plain Node and load no node_modules. It took 1.089 s. The first attempt failed because /usr/bin/time was missing, so I used the shell timer and repeated the baseline after successful installation.
- The first planting script stopped before mutation because a guard substring occurs in two functions. I restricted the edit to the first function. A subsequent gofmt lookup failed because env.sh was sourced too late; sourcing it before the script corrected this. Neither attempt supplied a mutant result.
- TestSharedCanonicalize compares two internal implementations. I explicitly chose the called unicodeproperties functions as code under test and left the generated regexp tables unchanged as the self oracle. This avoids mutating the expected-value source. Unicode provenance does not independently prove this comparison correct.
- The menu says flip a condition or drop a statement, but M6 removes two condition operands. I conservatively labeled it supplemental after checking menu compliance, excluded it from kills, unique kills and verdicts, and retained its log and diff. Its initial choice was made before outcomes.
- M2 uses a false-condition selector in scratch; its standalone diff drops the whole exclusion if statement. They are behaviorally equivalent. The standalone diff passes go vet.
- Family boundaries are broader than filenames or top-level names. Six tests build different inputs for compareExecutionCases, so uniqueness and subsumption use their family row. Tests with additional custom checks, stateful loops and the witness remain separate.
- The full-package run is only about two seconds, so narrowing was unnecessary. bounded=false means all package rows were observed; it does not claim repo-wide uniqueness.
- A production failure of a built-in mutant witness can be a broken control precondition. M4 does that; its witness row is excluded rather than treated as another production kill. The witness itself is outside the requested four-row audit, so I did not weaken its checker.
- Nil is the empty Parse answer, which crashes the AST-inspecting row. Probes were run per intended row to retain complete observations. Standalone probe diffs replace full bodies so go vet sees no unreachable statements.
- The budget language about four compiler mutants concerns native rebuilds. These rows run Go parser/folding code; no native product is built. Twelve selector choices need one Go test-binary build, with separate ADAMIC_BUILD_CACHE_DIR values on all matrix runs.
- Three isolated measurements were taken for each scoped row and all candidate subsumers. The family needed an additional three combined runs; reporting a fast member as the family cost would understate the cost.

Timing, seconds: {"setup_seconds": 0, "nproc": 5, "npm_ci_seconds": 1.089, "switch_build_seconds": 1.12, "matrix_command_wall_seconds": 26.63472190700213, "matrix_binary_seconds": 23.517, "standalone_validation_seconds": 2.678388520000226, "isolated_binary_seconds": 15.677, "clean_baseline_binary_seconds": 1.881, "final_clean_binary_seconds": 1.654}. See phase and per-mutant timing JSON for separation of command wall time and binary time. Full-session wall time was not separately instrumented. No step exceeded 90 s. No runtime C/native build, opt-in dependency or skipped row occurred.

Limits: four rows receive verdicts; all package rows supply matrix context. No other package was tested, no repo-wide uniqueness was assessed, no independent Unicode/spec value was checked, and no new mutant was added after outcomes. Production sources were restored and the package passed. Standalone Go mutant/probe diffs all apply and pass go vet. No PR or main push.
