Starting commit: 16f436a16a8e3b9cd2343f449eb1b0b4577c135c

Code under test: lint main.ts run, Suggestion and SuggestionEdit constructors; reachable local Linter constructor, run, ancestry, node, walk, fixed, compareFindings, compareEdits; RuleContext constructor, report, start, node, enabled; Finding constructor; Settings constructor/load/read; generated RuleSet constructors/prepare/visit/finish and fixture Rule constructor/visit/create; written and parser/scanner dependencies. This is a static inventory, not an exhaustive dynamic call graph of parser internals.

Oracle: unchanged Go cohere with testdata/serialization/oracle.go adapter; hand-written expected field strings for shard 000; hand-written suite-construction and child-guard assertions. Node executes source or generated JS, it is not the expected lint answer.

All five planned edits were fixed before checking any mutant result. Four production mutants use the fixed menu. P1 is an entry empty-answer probe, excluded from verdict kills. No selector is available in source without adding runtime instrumentation: rebuild each port mutant, at most four.

Potential reachable local implementation declarations. Static overapproximation; parser/scanner dependencies are outside this package.
stage1/cohere/lint/main.ts:11: function run(row: string, countOnly: boolean): number {
stage1/cohere/lint/suggestions.a:6: constructor(start: number, end: number, text: string) {
stage1/cohere/lint/suggestions.a:16: constructor(id: string, message: string, edits: readonly SuggestionEdit[]) {
stage1/cohere/lint/lint.ts:20: function compareFindings(left: Finding, right: Finding): number {
stage1/cohere/lint/lint.ts:23: function compareEdits(left: Finding, right: Finding): number {
stage1/cohere/lint/lint.ts:56: constructor(
stage1/cohere/lint/lint.ts:75: run(): void {
stage1/cohere/lint/lint.ts:106: ancestry(index: number, parent: number): void {
stage1/cohere/lint/lint.ts:112: node(index: number): ParseNode {
stage1/cohere/lint/lint.ts:115: walk(index: number, parent: number, rules: RuleSet): void {
stage1/cohere/lint/lint.ts:121: fixed(): string {
stage1/cohere/lint/context.ts:15: function lineStartsOf(source: string): number[] {
stage1/cohere/lint/context.ts:33: function isFunctionKind(kind: string): boolean {
stage1/cohere/lint/context.ts:50: function isLiteralPartKind(kind: string): boolean {
stage1/cohere/lint/context.ts:64: export function space(character: string): boolean {
stage1/cohere/lint/context.ts:85: constructor(
stage1/cohere/lint/context.ts:108: node(index: number): ParseNode {
stage1/cohere/lint/context.ts:111: enabled(name: string): boolean {
stage1/cohere/lint/context.ts:114: start(index: number): number {
stage1/cohere/lint/context.ts:189: edit(index: number, text: string): SuggestionEdit {
stage1/cohere/lint/context.ts:192: line(position: number): number {
stage1/cohere/lint/context.ts:213: literalEnds(): readonly number[] {
stage1/cohere/lint/context.ts:221: literalSpans(index: number, ends: number[]): void {
stage1/cohere/lint/context.ts:230: child(index: number, position: number): number {
stage1/cohere/lint/context.ts:233: parent(index: number): number {
stage1/cohere/lint/context.ts:238: kind(index: number): string {
stage1/cohere/lint/context.ts:241: functionLike(index: number): boolean {
stage1/cohere/lint/context.ts:245: has(index: number, kind: string): boolean {
stage1/cohere/lint/context.ts:253: children(index: number, kind: string): number[] {
stage1/cohere/lint/context.ts:256: scanAt(position: number): string {
stage1/cohere/lint/context.ts:260: scanStart(): number {
stage1/cohere/lint/context.ts:263: scanEnd(): number {
stage1/cohere/lint/context.ts:266: scanKind(): string {
stage1/cohere/lint/context.ts:269: raw(index: number): string {
stage1/cohere/lint/context.ts:274: name(index: number): number {
stage1/cohere/lint/context.ts:295: memberName(index: number): number {
stage1/cohere/lint/context.ts:300: members(index: number): number[] {
stage1/cohere/lint/context.ts:314: arguments(index: number): number[] {
stage1/cohere/lint/context.ts:318: property(index: number): number {
stage1/cohere/lint/context.ts:322: questionDot(index: number): boolean {
stage1/cohere/lint/context.ts:325: initializer(index: number): number {
stage1/cohere/lint/context.ts:330: extendsClass(index: number): boolean {
stage1/cohere/lint/context.ts:333: comment(index: number): boolean {
stage1/cohere/lint/context.ts:345: unwrap(index: number): number {
stage1/cohere/lint/context.ts:352: literalType(index: number): string {
stage1/cohere/lint/context.ts:371: tokens(start: number, end: number): string {
stage1/cohere/lint/context.ts:379: signature(index: number): string {
stage1/cohere/lint/settings.ts:9: load(text: string): void {
stage1/cohere/lint/settings.ts:65: read(name: string, fallback: string): string {
stage1/cohere/lint/settings.ts:72: list(name: string, fallback: string[]): string[] {
stage1/cohere/lint/finding.ts:18: constructor(
internal/testguard/guard.go:22: func init() {
internal/testguard/guard.go:47: func Run(command *exec.Cmd, budget, ceiling time.Duration) error {
