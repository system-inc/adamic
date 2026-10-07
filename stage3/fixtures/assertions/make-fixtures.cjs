// Extract complete upstream functions; the support types and drivers are fixture scaffolding.
// NODE_PATH=<stock 6.0.3 modules> node make-fixtures.cjs <TypeScript checkout>
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const root = process.argv[2];
const directory = __dirname;
const specs = [];
function add(file, source, name, reason, support, driver) { specs.push({file,source,name,reason,support,driver}); }
const bang = 'the non-null assertion !';
const cast = 'type assertion sites';
add('01_map_call.a','core.ts','getOrUpdate',bang,'',`
const values = new Map<string, number>();
console.log(String(getOrUpdate(values, "id", () => 17)));
console.log(String(getOrUpdate(values, "id", () => 99)));
`);
add('02_code_point_call.a','scanner.ts','codePointAt',bang,'',`
console.log(String(codePointAt("identifier", 0)));
console.log(String(codePointAt("𐐀name", 0)));
`);
add('03_exports_field.a','checker.ts','hasExportAssignmentSymbol',bang,`
interface Symbol { exports?: Map<string, number>; }
const InternalSymbolName = { ExportEquals: "export=" } as const;
`,`
console.log(String(hasExportAssignmentSymbol({ exports: new Map<string, number>() })));
console.log(String(hasExportAssignmentSymbol({ exports: new Map<string, number>([["export=", 1]]) })));
`);
add('04_field_then_call.a','utilities.ts','createTokenRange',bang,`
type SyntaxKind = number;
interface TextRange { pos: number; end: number; }
function tokenToString(token: SyntaxKind): string | undefined { return token === 1 ? "=>" : "+"; }
function createRange(pos: number, end: number): TextRange { return { pos, end }; }
`,`
const range = createTokenRange(10, 1);
console.log(String(range.pos) + ":" + String(range.end));
`);
add('05_regex_call_index.a','scanner.ts','scanShebangTrivia',bang,`
const shebangTriviaRegex = /^#!.*/;
`,`
console.log(String(scanShebangTrivia("#!/usr/bin/node\\nlet x", 0)));
console.log(String(scanShebangTrivia("#!node", 0)));
`);
add('06_memoize_clear.a','core.ts','memoize',bang,'',`
let calls = 0;
const read = memoize(() => { calls++; return 42; });
console.log(String(read()));
console.log(String(read()));
console.log(String(calls));
`);
add('07_indexed_operation.a','transformers/generators.ts','writeOperation',bang,`
const OpCode = { Nop: 0, Statement: 1, Assign: 2, Break: 3, BreakWhenTrue: 4, BreakWhenFalse: 5, Yield: 6, YieldStar: 7, Return: 8, Throw: 9, Endfinally: 10 } as const;
interface Expression { readonly kind: 80; readonly text: string; }
interface Statement { readonly kind: 245; readonly expression: Expression; }
type Label = number;
let operations: number[] | undefined = [0, 1, 2, 10];
type OperationArguments = [Label] | [Label, Expression] | [Statement] | [Expression | undefined] | [Expression, Expression];
let operationArguments: OperationArguments[] | undefined = [[0], [{ kind: 245, expression: { kind: 80, text: "statement" } }], [{ kind: 80, text: "left" }, { kind: 80, text: "right" }], [0]];
let operationLocations: (number | undefined)[] | undefined = [0, 1, 2, 3];
let lastOperationWasAbrupt = false;
let lastOperationWasCompletion = false;
function tryEnterLabel(index: number): void {}
function tryEnterOrLeaveBlock(index: number): void {}
function writeEndfinally(): void { console.log("endfinally"); }
function writeStatement(statement: Statement): void { console.log(statement.expression.text); }
function writeAssign(left: Expression, right: Expression, location: number | undefined): void { console.log(left.text + "=" + right.text); }
function writeBreak(label: Label, location: number | undefined): void {}
function writeBreakWhenTrue(label: Label, expression: Expression, location: number | undefined): void {}
function writeBreakWhenFalse(label: Label, expression: Expression, location: number | undefined): void {}
function writeYield(expression: Expression, location: number | undefined): void {}
function writeYieldStar(expression: Expression, location: number | undefined): void {}
function writeReturn(expression: Expression, location: number | undefined): void {}
function writeThrow(expression: Expression, location: number | undefined): void {}
`,`
writeOperation(0);
writeOperation(1);
writeOperation(2);
writeOperation(3);
`);
add('08_optional_start.a','scanner.ts','setText',bang,`
let text = "";
let end = 0;
function resetTokenState(position: number): void { console.log(String(position) + ":" + String(end) + ":" + text); }
`,`
setText("abc", 0, 2);
setText("name", undefined, undefined);
setText(undefined, undefined, undefined);
`);
const nodeSupport = `
const SyntaxKind = { ObjectLiteralExpression: 211, ArrayLiteralExpression: 210, ParenthesizedExpression: 218, ThisKeyword: 110, SuperKeyword: 108, PropertyAccessExpression: 212, ElementAccessExpression: 213 } as const;
type SyntaxKind = number;
interface Node { readonly kind: SyntaxKind; }
interface ObjectLiteralExpression extends Node { readonly kind: 211; readonly properties: readonly number[]; }
interface ArrayLiteralExpression extends Node { readonly kind: 210; readonly elements: readonly number[]; }
interface ParenthesizedExpression extends Node { readonly kind: 218; readonly expression: Expression; }
type Expression = Node;
`;
add('09_interface_kind.a','utilities.ts','isEmptyObjectLiteral',cast,nodeSupport,`
const empty: ObjectLiteralExpression = { kind: 211, properties: [] };
const full: ObjectLiteralExpression = { kind: 211, properties: [1] };
const array: ArrayLiteralExpression = { kind: 210, elements: [] };
console.log(String(isEmptyObjectLiteral(empty)));
console.log(String(isEmptyObjectLiteral(full)));
console.log(String(isEmptyObjectLiteral(array)));
`);
add('10_union_kind.a','utilities.ts','isEmptyArrayLiteral',cast,`
const SyntaxKind = { ObjectLiteralExpression: 211, ArrayLiteralExpression: 210 } as const;
interface ObjectLiteralExpression { readonly kind: 211; readonly properties: readonly number[]; }
interface ArrayLiteralExpression { readonly kind: 210; readonly elements: readonly number[]; }
type Node = ObjectLiteralExpression | ArrayLiteralExpression;
`,`
console.log(String(isEmptyArrayLiteral({ kind: 210, elements: [] })));
console.log(String(isEmptyArrayLiteral({ kind: 210, elements: [1] })));
console.log(String(isEmptyArrayLiteral({ kind: 211, properties: [] })));
`);
add('11_parenthesized_kind.a','utilities.ts','unwrapParenthesizedExpression',cast,nodeSupport,`
const leaf: Node = { kind: 110 };
const inner: ParenthesizedExpression = { kind: 218, expression: leaf };
const outer: ParenthesizedExpression = { kind: 218, expression: inner };
console.log(String(unwrapParenthesizedExpression(outer).kind));
console.log(String(unwrapParenthesizedExpression(leaf).kind));
`);
add('12_union_target.a','utilities.ts','isThisProperty',cast,nodeSupport+`
interface PropertyAccessExpression extends Node { readonly kind: 212; readonly expression: Node; }
interface ElementAccessExpression extends Node { readonly kind: 213; readonly expression: Node; }
`,`
const property: PropertyAccessExpression = { kind: 212, expression: { kind: 110 } };
const element: ElementAccessExpression = { kind: 213, expression: { kind: 108 } };
console.log(String(isThisProperty(property)));
console.log(String(isThisProperty(element)));
console.log(String(isThisProperty({ kind: 110 })));
`);
add('13_flag_downcast.a','utilities.ts','getCheckFlags',cast,`
type CheckFlags = number;
const SymbolFlags = { Transient: 33554432 } as const;
interface Symbol { readonly flags: number; }
interface TransientSymbol extends Symbol { readonly links: { readonly checkFlags: CheckFlags; }; }
`,`
const transient: TransientSymbol = { flags: SymbolFlags.Transient, links: { checkFlags: 8 } };
console.log(String(getCheckFlags(transient)));
console.log(String(getCheckFlags({ flags: 0 })));
`);
add('14_structural_cache.a','checker.ts','getCachedIterationTypes',cast,`
interface Type { readonly flags: number; }
interface IterationTypes { readonly value: number; }
interface IterableOrIteratorType extends Type { iterationTypes?: IterationTypes; }
type MatchingKeys<T, U> = "iterationTypes";
`,`
const type: IterableOrIteratorType = { flags: 1, iterationTypes: { value: 7 } };
console.log(String(getCachedIterationTypes(type, "iterationTypes")?.value));
console.log(String(getCachedIterationTypes({ flags: 1 }, "iterationTypes")));
`);
add('15_mutable_view.a','utilities.ts','setTextRangePos',cast,`
interface ReadonlyTextRange { readonly pos: number; readonly end: number; }
interface TextRange { pos: number; end: number; }
`,`
const range: ReadonlyTextRange = { pos: 1, end: 9 };
console.log(String(setTextRangePos(range, 4).pos));
console.log(String(range.pos));
`);
add('16_unknown_chain.a','watchUtilities.ts','getFallbackOptions',cast,`
type WatchFileKind = number & { readonly __watchFileKind: unique symbol; };
const WatchFileKind = { PriorityPollingInterval: 1 as WatchFileKind } as const;
type PollingWatchKind = number & { readonly __pollingWatchKind: unique symbol; };
interface WatchOptions { readonly fallbackPolling?: PollingWatchKind; readonly watchFile?: WatchFileKind; }
`,`
console.log(String(getFallbackOptions(undefined).watchFile));
console.log(String(getFallbackOptions({ fallbackPolling: 2 as PollingWatchKind }).watchFile));
`);
add('17_brand_upcast.a','utilities.ts','isExternalModuleSymbol',cast,`
type __String = string & { readonly __escapedIdentifier: unique symbol; };
interface Symbol { readonly flags: number; readonly escapedName: __String; }
const SymbolFlags = { Module: 1536 } as const;
const CharacterCodes = { doubleQuote: 34 } as const;
`,`
console.log(String(isExternalModuleSymbol({ flags: 512, escapedName: '\"mod\"' as __String })));
console.log(String(isExternalModuleSymbol({ flags: 512, escapedName: 'namespace' as __String })));
`);
add('18_const_tuple.a','checker.ts',null,cast,`
interface NodeLinks { resolvedSignature: number; }
const cachedResolvedSignatures: (readonly [NodeLinks, number])[] = [];
const nodeLinks: NodeLinks = { resolvedSignature: 7 };
`,`
console.log(String(cachedResolvedSignatures.length));
console.log(String(cachedResolvedSignatures[0]?.[1]));
`);
add('19_identifier_kind.a','checker.ts','getIdentifierFromEntityNameExpression',cast,`
const SyntaxKind = { Identifier: 80, PropertyAccessExpression: 212 } as const;
interface Node { readonly kind: number; }
type Expression = Node;
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
interface PrivateIdentifier extends Node { readonly kind: 81; readonly escapedText: string; }
interface PropertyAccessExpression extends Node { readonly kind: 212; readonly name: Identifier; }
`,`
const id: Identifier = { kind: 80, escapedText: "name" };
const access: PropertyAccessExpression = { kind: 212, name: id };
console.log(getIdentifierFromEntityNameExpression(id)?.escapedText ?? "missing");
console.log(getIdentifierFromEntityNameExpression(access)?.escapedText ?? "missing");
console.log(getIdentifierFromEntityNameExpression({ kind: 110 })?.escapedText ?? "missing");
`);
add('20_type_flag_mask.a','utilities.ts','getObjectFlags',cast,`
type ObjectFlags = number;
const TypeFlags = { ObjectFlagsType: 403963917 } as const;
interface Type { readonly flags: number; }
interface ObjectFlagsType extends Type { readonly objectFlags: ObjectFlags; }
`,`
const object: ObjectFlagsType = { flags: 1048576, objectFlags: 2 };
console.log(String(getObjectFlags(object)));
console.log(String(getObjectFlags({ flags: 32 })));
`);
const manifest = [];
for (const s of specs) {
 const filename = path.join(root,'src/compiler',s.source);
 const text = fs.readFileSync(filename,'utf8');
 const f = ts.createSourceFile(filename,text,ts.ScriptTarget.Latest,true);
 let node;
 function visit(n) { if (ts.isFunctionDeclaration(n) && n.name?.text === s.name) node = n; ts.forEachChild(n,visit); }
 visit(f);
 let body, line, endLine;
 if (s.name === null) {
  line = 1992; endLine = line; body = text.split(/\r?\n/)[line-1].trim();
 } else {
  if (!node) throw Error(s.name);
  line = f.getLineAndCharacterOfPosition(node.getStart(f)).line+1;
  endLine = f.getLineAndCharacterOfPosition(node.end).line+1;
  body = node.getText(f).replace(/\r\n/g,'\n');
 }
 fs.writeFileSync(path.join(directory,s.file),`// From TypeScript 6.0.3, src/compiler/${s.source}:${line}\n// Census reason: ${s.reason}\n// Support types and driver are reduced scaffolding; the upstream function is unchanged.\n${s.support}\n${body}\n${s.driver}`);
 manifest.push({file:s.file,tsc:[`src/compiler/${s.source}:${line}`],reason:s.reason,source_end_line:endLine,function:s.name});
}
fs.writeFileSync(path.join(directory,'manifest.json'),JSON.stringify(manifest,null,2)+'\n');
console.log(`extracted ${manifest.length} fixtures`);
