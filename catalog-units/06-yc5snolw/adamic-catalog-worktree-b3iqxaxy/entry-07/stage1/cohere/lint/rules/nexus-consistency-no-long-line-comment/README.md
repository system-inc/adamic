# Nexus long line comment

Registered port of batch3:rules/no_long_line_comment.ts from codex/stage1-lint-batch3 fa9781c2. The SourceFile listener takes the node supplied by the registry. Options field 5 is decoded into upstream ConsistencyNoLongLineCommentOptions. Positive maximumLineCount overrides four; zero, negative, null and absent options keep the default.

Comments form a run only when they are standalone, consecutive, in the same byte column and are not directives. Long runs are replaced with a block comment preserving each line's prose and indentation. A star-slash in the run withholds the automatic fix. Messages retain upstream wording verbatim. Findings use the shared UTF-16 range API and wire conversion.

RuleContext.comments(node,index) owns reusable parser-anchored comment collection, cached for the immutable source file. comment_ranges.a is the batch's range model and literal-aware scanner. RuleContext.trim supplies Go strings.TrimSpace behavior. These helpers are shared context support as requested; no dispatch, registration generator or test harness edits are needed.

The threshold witness carries maximumLineCount:2 in its options sidecar and fires on three lines, while its all-rule row uses defaults. See REPORT.md and evidence for exact comparisons and the options-specific caught mutant.
