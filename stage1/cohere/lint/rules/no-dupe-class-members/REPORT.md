# Core no-dupe-class-members proof

Shared package: ../../helpers/classmembers. Go cohere pin
`7945d102a6c18dd36adf9114a758ce646e8b2359`; production oracle adapter delegates
unchanged to core.NoDupeClassMembers with its nil options adapter.

[Selected test](testdata/selected_test.go) selects all 37 unique upstream
source/rule/options combinations, both owned witnesses in selected and all-rule
mode, and applicable inherited rows: 126 manifest rows in total. Diagnostic
text, ranges, fixes, suggestions and final fixed text match byte-for-byte on
Node source, emitted JavaScript and ASan/UBSan native: 508,801 bytes.

Witnesses: [collisions](testdata/collisions.ts.txt) checks static/instance,
private/public, numeric/string, delimiter, computed and class-expression keys;
[accessors](testdata/accessors.ts.txt) checks first-accessor behavior, overloads,
abstract declarations and repeated fields. Both produce Go findings. The port
preserves Go's getter/setter/setter and setter/getter/getter non-reporting and
bodyless abstract-member exclusion; these are intentionally not corrected.

[Mutant](mutant.json) drops the static flag. It compiles, executes and disagrees
with Go on source Node, emitted JavaScript and sanitized native (case 86).
[Selected log](testdata/selected.log) records the differences. Selected run
PASS in 177.317 seconds. Core rule ownership audit: [CLAIM.md](CLAIM.md).
The TypeScript extension's earlier claim remains untouched.

Complete package gate evidence is recorded in
[the helper report](../../helpers/classmembers/REPORT.md).
