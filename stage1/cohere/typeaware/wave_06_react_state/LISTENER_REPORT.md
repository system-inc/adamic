Built: rule.json kinds declarations for the three retained React claims, using the pinned checker numeric SourceFile kind 307.
Commit: this report and the three declarations are on codex/typeaware-wave-06, based on current main e8ba3d5d.
Checks: declarations checked against unchanged Go production listener tables and its generated numeric AST enum.
Mutant: changing SourceFile 307 to another numeric kind is rejected by the independent production listener comparison.
Not covered: native dispatch integration, numeric native parser nodes and the previously documented native React source adapter.

All three production Go rules listen only to SourceFile. Their lowering descends into nested functions itself; function-level subscriptions would lower nested functions twice without their enclosing context. The numeric declarations therefore contain only 307, the SourceFile value in the pinned typescript-go AST enum.

The current native ParseNode and the copy on origin/codex/lint-harness-dot-a still expose only kind: string. No numeric native kind field or shared rule.json consumer exists in this checkout. These declarations record the pinned numeric listener contract for the upcoming shared driver; they do not establish that the current native parser emits those numbers. No string-based relevance adapter or per-rule node refetch is added. Native execution migration needs the shared parser and driver contract.

The existing prepared-HIR validators consume the graph they are handed. Their instruction kind labels describe HIR operations, not parser syntax relevance. Source analysis continues to refuse explicitly because native source lowering, SSA, compilation-unit selection, JSX lowering and memo-scope handling are missing. CORE_REPORT.md contains the complete comparisons, sanitizer results and semantic mutants. No new claims, shared-file edits or runtime behavior changes are made.

After fetching all heads, main remains e8ba3d5d and the own remote remains 5a638b72. The existing implementation is already rebased and green on that main. The declaration-only continuation does not change compiled inputs, so the recorded implementation oracle results remain applicable.
