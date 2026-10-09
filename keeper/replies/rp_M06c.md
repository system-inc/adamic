The diff applied at the exact base. An environment restart interrupted the initial run after 33m1s; four whole-package retries with fresh caches took up to 21m22s. Clean-base reruns took 98.9s. Three killed/deadline failures passed isolated base runs; this does not establish mutant causality. All 36 lint failures remained red at base. ESTree’s 30-minute timeout and JSON/YAML’s internal 90-second deadlines left coverage incomplete; default opt-in skips also remained. The tree is restored and clean. Nothing was pushed.

```json
{
  "mutant": "u045 M06 (element_borrow.go:276, captured roots may borrow), covered packages only",
  "hash": "e742af58576a",
  "base": "bc9edc5560379b63d4688a21006efb6032b60434",
  "applied": true,
  "packages_run": [
    "./internal/native",
    "./internal/oracle",
    "./stage1/cohere/css",
    "./stage1/cohere/estree",
    "./stage1/cohere/formatfiles",
    "./stage1/cohere/graphql",
    "./stage1/cohere/graphql/printer",
    "./stage1/cohere/json",
    "./stage1/cohere/lint",
    "./stage1/cohere/lint/helpers/comments",
    "./stage1/cohere/lint/rules/no-unsafe-negation",
    "./stage1/cohere/lint/rules/no-unsafe-optional-chaining",
    "./stage1/cohere/lint/rules/typescript-no-this-alias",
    "./stage1/cohere/markdownblocks",
    "./stage1/cohere/mediaquery",
    "./stage1/cohere/values",
    "./stage1/cohere/yaml",
    "./stage1/typescript/parser"
  ],
  "packages_failed_to_build": [],
  "caught_by": [
    {
      "package": "internal/native",
      "test": "TestDecodeASCIIUnit26",
      "subtest": null,
      "line": "decode_ascii_test.go:267: /tmp/replay-e742af58576a/cache/0350515d3f3cafa375ca973cf1b2373228a1c37fbf9a04d3fd66069bf2f7d650/decode-native [/tmp/replay-e742af58576a/cache/4e441548e24c77da2487d636be1bda64816b06e9a7cbb65187f44847ab1156f1/cases_053248_055296.bin]: signal: killed"
    },
    {
      "package": "stage1/cohere/json",
      "test": "TestPortMatchesGoCohere_123",
      "subtest": null,
      "line": "TestPortMatchesGoCohere_123 exceeded its 90s deadline"
    },
    {
      "package": "stage1/cohere/yaml",
      "test": "TestFileDriver_Setup",
      "subtest": null,
      "line": "panic: test timed out after 1m30s"
    }
  ],
  "red_at_base": [
    {"package": "stage1/cohere/lint", "test": "TestProduct_ShardsAgreePrepared"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_000"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_001"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_002"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_003"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_004"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_005"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_006"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_007"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_008"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_009"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_010"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_011"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_012"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_013"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_014"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_015"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_016"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_017"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_018"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_019"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_020"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_021"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_022"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_023"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_024"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_025"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_026"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_027"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_028"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_029"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_030"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_031"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_Setup"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_SetupRequired"},
    {"package": "stage1/cohere/lint", "test": "TestShardsAgree_Union"}
  ],
  "survived": false
}
```
