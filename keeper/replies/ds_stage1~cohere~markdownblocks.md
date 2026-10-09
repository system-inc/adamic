All 13 gathered mutants retained clean catches outside the deletion set. The green baseline took 909.3 seconds wall time; mutant replays totaled 3558.1 seconds, stopping after clean failures. No diffs were stale and no replay panicked or remained broken; an environment interruption was recovered. Production edits were restored and [evidence](/workspace/adamic/review/test-defend/deletion-set/stage1-cohere-markdownblocks/report.json) pushed to `test-defend/deletion-set/stage1-cohere-markdownblocks`.

```json
{
  "package": "stage1/cohere/markdownblocks",
  "main": "7b9d4272c28f59530ab13daa5c49067e47933b06",
  "skipped": [
    "TestMarkdownLayout family",
    "TestMarkdownListLayout",
    "TestMarkdownListLayout family",
    "TestMarkdownStructureLayout family",
    "TestMarkdownTableLayout",
    "TestMarkdownTableLayout family",
    "TestMdastIdentifierWitnesses"
  ],
  "mutants": [
    {
      "mutant": "audit-leaf_composition_products-M1",
      "file_line": "stage1/cohere/markdownblocks/codeblocks.ts:20",
      "branch": "test-audit/stage1-cohere-markdownblocks-leaf_composition_products",
      "candidates_failed": ["TestMarkdownLayout family", "TestMarkdownListLayout family"],
      "still_caught_by": ["TestMarkdownWhitespaceLayout_021", "TestMarkdownWhitespaceLayout_023"],
      "stale": false
    },
    {
      "mutant": "audit-leaf_composition_products-M2",
      "file_line": "stage1/cohere/markdownblocks/htmlblocks.ts:15",
      "branch": "test-audit/stage1-cohere-markdownblocks-leaf_composition_products",
      "candidates_failed": ["TestMarkdownLayout family", "TestMarkdownListLayout family"],
      "still_caught_by": ["TestMarkdownQuoteLayout_009"],
      "stale": false
    },
    {
      "mutant": "audit-leaf_composition_products-M3",
      "file_line": "stage1/cohere/markdownblocks/root.ts:99",
      "branch": "test-audit/stage1-cohere-markdownblocks-leaf_composition_products",
      "candidates_failed": ["TestMarkdownLayout family", "TestMarkdownListLayout family"],
      "still_caught_by": ["TestMarkdownQuoteLayout_009"],
      "stale": false
    },
    {
      "mutant": "audit-mdast_errors-M1",
      "file_line": "stage1/cohere/markdownblocks/mdastCompile.ts:620",
      "branch": "test-audit/stage1-cohere-markdownblocks-mdast_errors",
      "candidates_failed": ["TestMdastIdentifierWitnesses"],
      "still_caught_by": ["TestNativeMdastConstruction"],
      "stale": false
    },
    {
      "mutant": "audit-structure_layout_shards-M1",
      "file_line": "stage1/cohere/markdownblocks/width.ts:23",
      "branch": "test-audit/stage1-cohere-markdownblocks-structure_layout_shards",
      "candidates_failed": ["TestMarkdownStructureLayout family", "TestMarkdownTableLayout family"],
      "still_caught_by": ["TestMarkdownUnicodeWidths"],
      "stale": false
    },
    {
      "mutant": "audit-structure_layout_shards-M2",
      "file_line": "stage1/cohere/markdownblocks/structure.ts:51",
      "branch": "test-audit/stage1-cohere-markdownblocks-structure_layout_shards",
      "candidates_failed": ["TestMarkdownStructureLayout family", "TestMarkdownTableLayout family"],
      "still_caught_by": ["TestMarkdownLeafComposition_002", "TestMarkdownLeafComposition_003"],
      "stale": false
    },
    {
      "mutant": "defend-leaf_composition_products-layout-D1",
      "file_line": "stage1/cohere/markdownblocks/codeblocks.ts:34",
      "branch": "test-defend/stage1-cohere-markdownblocks-leaf_composition_products",
      "candidates_failed": ["TestMarkdownLayout family", "TestMarkdownListLayout family"],
      "still_caught_by": ["TestMarkdownQuoteLayout_009"],
      "stale": false
    },
    {
      "mutant": "defend-leaf_composition_products-layout-D3",
      "file_line": "stage1/cohere/markdownblocks/root.ts:76",
      "branch": "test-defend/stage1-cohere-markdownblocks-leaf_composition_products",
      "candidates_failed": ["TestMarkdownLayout family", "TestMarkdownListLayout family"],
      "still_caught_by": ["TestMarkdownQuoteLayout_008"],
      "stale": false
    },
    {
      "mutant": "defend-mdast_errors-D2",
      "file_line": "stage1/cohere/markdownblocks/identifierCaseValues.ts:5",
      "branch": "test-defend/stage1-cohere-markdownblocks-mdast_errors",
      "candidates_failed": ["TestMdastIdentifierWitnesses"],
      "still_caught_by": ["TestMdastIdentifierScalars"],
      "stale": false
    },
    {
      "mutant": "defend-mdast_errors-D3",
      "file_line": "stage1/cohere/markdownblocks/identifierCaseValues.ts:13",
      "branch": "test-defend/stage1-cohere-markdownblocks-mdast_errors",
      "candidates_failed": ["TestMdastIdentifierWitnesses"],
      "still_caught_by": ["TestMdastIdentifierScalars"],
      "stale": false
    },
    {
      "mutant": "defend-mdast_errors-D4",
      "file_line": "stage1/cohere/markdownblocks/mdastCompile.ts:109",
      "branch": "test-defend/stage1-cohere-markdownblocks-mdast_errors",
      "candidates_failed": ["TestMdastIdentifierWitnesses"],
      "still_caught_by": ["TestNativeMdastConstruction"],
      "stale": false
    },
    {
      "mutant": "defend-structure_layout_shards-D3",
      "file_line": "stage1/cohere/markdownblocks/structure.ts:25",
      "branch": "test-defend/stage1-cohere-markdownblocks-structure_layout_shards",
      "candidates_failed": ["TestMarkdownStructureLayout family", "TestMarkdownTableLayout family"],
      "still_caught_by": ["TestMarkdownQuoteLayout_011"],
      "stale": false
    },
    {
      "mutant": "defend-structure_layout_shards-D4",
      "file_line": "stage1/cohere/markdownblocks/tables.ts:51",
      "branch": "test-defend/stage1-cohere-markdownblocks-structure_layout_shards",
      "candidates_failed": ["TestMarkdownStructureLayout family", "TestMarkdownTableLayout family"],
      "still_caught_by": [
        "TestMarkdownQuoteLayout_000",
        "TestMarkdownQuoteLayout_001",
        "TestMarkdownQuoteLayout_002",
        "TestMarkdownQuoteLayout_003",
        "TestMarkdownQuoteLayout_004"
      ],
      "stale": false
    }
  ],
  "keep": [],
  "deletable": [
    "TestMarkdownLayout family",
    "TestMarkdownListLayout",
    "TestMarkdownListLayout family",
    "TestMarkdownStructureLayout family",
    "TestMarkdownTableLayout",
    "TestMarkdownTableLayout family",
    "TestMdastIdentifierWitnesses"
  ],
  "notes": [
    "TestMarkdownListLayout and TestMarkdownTableLayout had no gathered mutants; neither guards behavior demonstrated by the supplied mutants.",
    "The skip expression was expanded to include the differently named TestMarkdownLayout members and family Union tests.",
    "Deletable is limited to the gathered mutant evidence. No tests were deleted."
  ]
}
```
