// Construction slice of lower.go / lower_expression.go. Other paths are explicit declines.
import { panic, utf8Length } from 'adamic';
import { HIRFunction, Instruction, BasicBlock, HIRArena, ConstructedHIR } from './core.ts';
import { SymbolSnapshot } from './symbol.ts';
import { ExportResolver, emptyOrigin } from './export_origin.ts';
import type { OriginInterface } from './export_origin.ts';
import { Scanner } from '../../typescript/scanner/scanner.ts';
import { Parser } from '../../typescript/parser/parser.ts';
import { written } from '../../typescript/parser/nodes.ts';
import type { PlaceInterface, ValueType, TerminalType, ArgumentInterface, ModuleExportOriginInterface, JsxTagInterface, JsxAttributeInterface, ArrayElementInterface, ObjectPropertyInterface, FunctionIndex, BlockIndex, DeclarationIndex, PatternIndex, PatternPropertyInterface, PatternElementInterface } from './core.ts';
import { constructHIR } from './graph.ts';
function literal(kind: string, text: string): string | undefined {
    if(['NumericLiteral', 'StringLiteral', 'BigIntLiteral', 'NoSubstitutionTemplateLiteral'].includes(kind)) { return `string:${written(text)}`; }
    if(kind === 'TrueKeyword') { return 'bool:true'; }
    if(kind === 'FalseKeyword') { return 'bool:false'; }
    if(kind === 'NullKeyword') { return 'nil'; }
    return undefined;
}
const operators = new Map<string, string>([
    ['PlusToken', '+'], ['MinusToken', '-'], ['AsteriskToken', '*'], ['SlashToken', '/'],
    ['PercentToken', '%'], ['AsteriskAsteriskToken', '**'], ['LessThanToken', '<'], ['GreaterThanToken', '>'],
    ['LessThanEqualsToken', '<='], ['GreaterThanEqualsToken', '>='], ['EqualsEqualsToken', '=='],
    ['ExclamationEqualsToken', '!='], ['EqualsEqualsEqualsToken', '==='], ['ExclamationEqualsEqualsToken', '!=='],
    ['AmpersandToken', '&'], ['BarToken', '|'], ['CaretToken', '^'], ['LessThanLessThanToken', '<<'],
    ['GreaterThanGreaterThanToken', '>>'], ['GreaterThanGreaterThanGreaterThanToken', '>>>'],
    ['InKeyword', 'in'], ['InstanceOfKeyword', 'instanceof'], ['ExclamationToken', '!'], ['TildeToken', '~'],
]);
const compounds = new Map<string, string>([
 ['AmpersandAmpersandEqualsToken', '&&'], ['BarBarEqualsToken', '||'], ['QuestionQuestionEqualsToken', '??'], ['PlusEqualsToken', '+'], ['MinusEqualsToken', '-'], ['AsteriskEqualsToken', '*'], ['SlashEqualsToken', '/'], ['PercentEqualsToken', '%'], ['AsteriskAsteriskEqualsToken', '**'], ['AmpersandEqualsToken', '&'], ['BarEqualsToken', '|'], ['CaretEqualsToken', '^'], ['LessThanLessThanEqualsToken', '<<'], ['GreaterThanGreaterThanEqualsToken', '>>'], ['GreaterThanGreaterThanGreaterThanEqualsToken', '>>>'],
]);
const logicals = new Map<string, string>([['AmpersandAmpersandToken', '&&'], ['BarBarToken', '||'], ['QuestionQuestionToken', '??']]);
function supportedTarget(parser: Parser, id: number): boolean {
 const node = parser.node(id);
 if(node.kind === 'ParenthesizedExpression') { return supportedTarget(parser, node.children[0] ?? -1); }
 if(node.kind === 'AsExpression' || node.kind === 'TypeAssertionExpression' || node.kind === 'SatisfiesExpression') { return true; }
 if(node.kind === 'ObjectLiteralExpression' || node.kind === 'ArrayLiteralExpression') { return true; }
 return node.kind === 'Identifier' || ((node.kind === 'PropertyAccessExpression' || node.kind === 'ElementAccessExpression') && supportedExpression(parser, id));
}
function supportedExpression(parser: Parser, id: number): boolean {
    const node = parser.node(id);
    if(node.kind === 'ThisKeyword' || node.kind === 'RegularExpressionLiteral') { return true; }
    if(node.kind === 'TemplateExpression') { return node.children.slice(1).every((child) => supportedExpression(parser, parser.node(child).children[0] ?? -1)); }
    if(node.kind === 'TaggedTemplateExpression') { return supportedExpression(parser, node.children[0] ?? -1) && supportedExpression(parser, node.children[node.children.length - 1] ?? -1); }
    if(node.kind === 'AsExpression' || node.kind === 'SatisfiesExpression' || node.kind === 'TypeAssertionExpression') { return supportedExpression(parser, node.children[node.kind === 'TypeAssertionExpression' ? 1 : 0] ?? -1); }
    if(node.kind === 'AwaitExpression' || node.kind === 'DeleteExpression') { return supportedExpression(parser, node.children[0] ?? -1); }
    if(node.kind === 'NonNullExpression') { return supportedExpression(parser, node.children[0] ?? -1); }
    if(node.kind === 'ArrayLiteralExpression') { return node.children.every((child) => { const element = parser.node(child); return element.kind === 'OmittedExpression' || supportedExpression(parser, element.kind === 'SpreadElement' ? element.children[0] ?? -1 : child); }); }
    if(node.kind === 'ObjectLiteralExpression') {
        return node.children.every((child) => {
            const member = parser.node(child);
            if(member.kind === 'SpreadAssignment') { return supportedExpression(parser, member.children[0] ?? -1); }
            if(member.kind === 'ShorthandPropertyAssignment') { return parser.node(member.children[0] ?? -1).kind === 'Identifier'; }
            if(['MethodDeclaration', 'GetAccessor', 'SetAccessor'].includes(member.kind)) { return supportedFunction(parser, child); }
            if(member.kind !== 'PropertyAssignment') { return true; }
            const name = parser.node(member.children[0] ?? -1);
            return (name.kind !== 'ComputedPropertyName' || supportedExpression(parser, name.children[0] ?? -1)) && supportedExpression(parser, member.children[1] ?? -1);
        });
    }
    if(node.kind === 'JsxExpression') { return node.children.every((child) => parser.node(child).kind === 'DotDotDotToken' || supportedExpression(parser, child)); }
    if(node.kind === 'JsxElement' || node.kind === 'JsxSelfClosingElement' || node.kind === 'JsxFragment') {
        if(node.kind !== 'JsxFragment') {
            const opening = node.kind === 'JsxElement' ? parser.node(node.children[0] ?? -1) : node;
            const tag = parser.node(opening.children[0] ?? -1);
            if(tag.kind === 'PropertyAccessExpression' && !supportedExpression(parser, opening.children[0] ?? -1)) { return false; }
            const attributes = parser.node(opening.children[opening.children.length - 1] ?? -1);
            for(const child of attributes.children) { const attribute = parser.node(child); const value = attribute.children[attribute.kind === 'JsxSpreadAttribute' ? 0 : 1]; if(value !== undefined && !supportedExpression(parser, value)) { return false; } }
        }
        return node.kind === 'JsxSelfClosingElement' || node.children.slice(1, node.children.length - 1).every((child) => parser.node(child).kind === 'JsxText' || supportedExpression(parser, child));
    }
    if(node.kind === 'FunctionExpression' || node.kind === 'ArrowFunction') { return supportedFunction(parser, id); }
    if(node.kind === 'Identifier' || literal(node.kind, node.text) !== undefined) { return true; }
    if(['ParenthesizedExpression', 'TypeOfExpression', 'VoidExpression'].includes(node.kind)) {
        return node.children.length === 1 && supportedExpression(parser, node.children[0] ?? -1);
    }
    if(node.kind === 'PropertyAccessExpression' || node.kind === 'ElementAccessExpression') {
        return supportedExpression(parser, node.children[0] ?? -1) && (node.kind === 'PropertyAccessExpression' || supportedExpression(parser, node.children[node.children.length - 1] ?? -1));
    }
    if(node.kind === 'CallExpression' || node.kind === 'NewExpression') {
        if(!supportedExpression(parser, node.children[0] ?? -1)) { return false; }
        const count = Math.max(0, node.list);
        for(const argument of node.children.slice(node.children.length - count)) {
            const arg = parser.node(argument);
            if(!supportedExpression(parser, arg.kind === 'SpreadElement' ? arg.children[0] ?? -1 : argument)) { return false; }
        }
        return true;
    }
    if(node.kind === 'PostfixUnaryExpression') { return supportedTarget(parser, node.children[0] ?? -1); }
    if(node.kind === 'PrefixUnaryExpression') {
        if(node.operator === 'PlusPlusToken' || node.operator === 'MinusMinusToken') { return supportedTarget(parser, node.children[0] ?? -1); }
        return ['PlusToken', 'MinusToken', 'ExclamationToken', 'TildeToken'].includes(node.operator) && supportedExpression(parser, node.children[0] ?? -1);
    }
    if(node.kind === 'ConditionalExpression') { return supportedExpression(parser, node.children[0] ?? -1) && supportedExpression(parser, node.children[2] ?? -1) && supportedExpression(parser, node.children[4] ?? -1); }
    if(node.kind === 'BinaryExpression') {
        const operator = parser.node(node.children[1] ?? -1).kind;
        if(operator === 'EqualsToken' || compounds.has(operator)) { return supportedTarget(parser, node.children[0] ?? -1) && supportedExpression(parser, node.children[2] ?? -1); }
        return (operator === 'CommaToken' || operators.has(operator) || logicals.has(operator)) && supportedExpression(parser, node.children[0] ?? -1) && supportedExpression(parser, node.children[2] ?? -1);
    }
    return true;
}
function declarationInitializer(parser: Parser, id: number): number | undefined {
    const node = parser.node(id);
    if(node.children.length < 2) { return undefined; }
    const previous = parser.node(node.children[node.children.length - 2] ?? -1);
    const scanner = new Scanner(parser.scanner.text);
    scanner.pos = previous.end; scanner.scan();
    return scanner.kind === 'EqualsToken' ? node.children[node.children.length - 1] : undefined;
}
function parameterBinding(parser: Parser, id: number): number {
    return parser.node(id).children.find((child) => ['Identifier', 'ObjectBindingPattern', 'ArrayBindingPattern'].includes(parser.node(child).kind)) ?? -1;
}
function supportedList(parser: Parser, id: number): boolean {
    const list = parser.node(id);
    if(list.kind !== 'VariableDeclarationList') { return false; }
    for(const child of list.children) { const declaration = parser.node(child); const initializer = declarationInitializer(parser, child); if((initializer !== undefined && !supportedExpression(parser, initializer))) { return false; } }
    return true;
}
// Recover for-header roles from existing child spans without modifying the parser.
// Scan only gaps between complete AST children: nested semicolons, templates and
// regular expressions are skipped with their child, never interpreted as separators.
function forHeaderSlots(parser: Parser, id: number): number[] {
    const node = parser.node(id); const slots: number[] = [-1, -1, -1, node.children.at(-1) ?? -1];
    const scanner = new Scanner(parser.scanner.text); scanner.pos = node.pos; let role = 0;
    for(let i = 0; i + 1 < node.children.length; i++) {
        const child = node.children[i] ?? -1; const part = parser.node(child);
        while(scanner.pos < part.pos) { if(scanner.scan() === 'SemicolonToken') { role++; } }
        if(role > 2) { panic('invalid for-header roles'); }
        slots[role] = child; scanner.pos = part.end;
    }
    return slots;
}
function supportedStatement(parser: Parser, id: number): boolean {
    const statement = parser.node(id);
    if(statement.kind === 'FunctionDeclaration') { return supportedFunction(parser, id); }
    if(statement.kind === 'Block') { return statement.children.every((child) => supportedStatement(parser, child)); }
    if(statement.kind === 'IfStatement') {
        return supportedExpression(parser, statement.children[0] ?? -1) && supportedStatement(parser, statement.children[1] ?? -1) && (statement.children[2] === undefined || supportedStatement(parser, statement.children[2] ?? -1));
    }
    if(statement.kind === 'DoStatement') { return supportedStatement(parser, statement.children[0] ?? -1) && supportedExpression(parser, statement.children[1] ?? -1); }
    if(statement.kind === 'ForStatement') {
        const slots = forHeaderSlots(parser, id); const init = slots[0] ?? -1;
        return (init < 0 || (parser.node(init).kind === 'VariableDeclarationList' ? supportedList(parser, init) : supportedExpression(parser, init))) && slots.slice(1, 3).every((child) => child < 0 || supportedExpression(parser, child)) && supportedStatement(parser, slots[3] ?? -1);
    }
    if(statement.kind === 'ForOfStatement' || statement.kind === 'ForInStatement') {
        const offset = parser.node(statement.children[0] ?? -1).kind === 'AwaitKeyword' ? 1 : 0;
        const init = statement.children[offset] ?? -1;
        return (parser.node(init).kind === 'VariableDeclarationList' ? supportedList(parser, init) : supportedTarget(parser, init)) && supportedExpression(parser, statement.children[offset + 1] ?? -1) && supportedStatement(parser, statement.children[offset + 2] ?? -1);
    }
    if(statement.kind === 'SwitchStatement') { return supportedExpression(parser, statement.children[0] ?? -1) && parser.node(statement.children[1] ?? -1).children.every((child) => { const clause = parser.node(child); return (clause.kind === 'DefaultClause' || supportedExpression(parser, clause.children[0] ?? -1)) && clause.children.slice(clause.kind === 'CaseClause' ? 1 : 0).every((id) => supportedStatement(parser, id)); }); }
    if(statement.kind === 'LabeledStatement') { return supportedStatement(parser, statement.children[1] ?? -1); }
    if(statement.kind === 'TryStatement') { return statement.children.every((child) => { const part = parser.node(child); if(part.kind !== 'CatchClause') { return supportedStatement(parser, child); } const binding = part.children.length > 1 ? parser.node(part.children[0] ?? -1) : undefined; return supportedStatement(parser, part.children[part.children.length - 1] ?? -1); }); }
    if(statement.kind === 'WhileStatement') { return supportedExpression(parser, statement.children[0] ?? -1) && supportedStatement(parser, statement.children[1] ?? -1); }
    if(statement.kind === 'BreakStatement' || statement.kind === 'ContinueStatement') { return statement.children.length <= 1; }
    if(statement.kind === 'VariableStatement') { return supportedList(parser, statement.children[0] ?? -1); }
    if(['ExpressionStatement', 'ReturnStatement', 'ThrowStatement'].includes(statement.kind)) { return statement.children.every((child) => supportedExpression(parser, child)); }
    return true;
}
function supportedFunction(parser: Parser, root: number): boolean {
    const node = parser.node(root);
    const body = node.children[node.children.length - 1] ?? -1;
    for(const child of node.children) { if(parser.node(child).kind === 'Parameter' && parameterBinding(parser, child) < 0) { return false; } }
    return parser.node(body).kind === 'Block' ? supportedStatement(parser, body) : node.kind === 'ArrowFunction' && supportedExpression(parser, body);
}
class StraightLineBuilder {
    readonly parser: Parser;
    readonly source: string;
    readonly fn: HIRFunction;
    readonly arena: HIRArena;
    readonly functionIndex: FunctionIndex;
    current: BasicBlock | undefined;
    readonly jumps: { readonly label: string; readonly breakBlock: BlockIndex; readonly continueBlock?: BlockIndex }[] = [];
    readonly symbols: SymbolSnapshot | undefined;
    readonly enclosing: StraightLineBuilder | undefined;
    readonly captured: Map<number, PlaceInterface> = new Map<number, PlaceInterface>();
    readonly captureSymbols: number[] = [];
    readonly contextual: Set<number> = new Set<number>();
    readonly locals: Map<number, PlaceInterface> = new Map<number, PlaceInterface>();
    constructor(parser: Parser, source: string, fn: HIRFunction, symbols: SymbolSnapshot | undefined, enclosing: StraightLineBuilder | undefined, arena: HIRArena, functionIndex: FunctionIndex) { this.arena = arena; this.functionIndex = functionIndex; this.parser = parser; this.source = source; this.fn = fn; this.symbols = symbols; this.enclosing = enclosing; this.current = fn.block(fn.entry); }
    identity(id: number): number {
        const node = this.parser.node(id);
        if(this.symbols === undefined) { return 0; }
        return this.symbols.read(this.byte(node.pos), this.byte(node.end)).identity;
    }
    findContext(root: number): void {
        if(this.symbols === undefined) { return; }
        const usage = new Map<number, { reassigned: boolean; referenced: boolean; innerWrite: boolean }>();
        this.contextUsage(root, 0, usage);
        for(const [symbol, entry] of usage) { if(symbol !== 0 && (entry.innerWrite || (entry.reassigned && entry.referenced))) { this.contextual.add(symbol); } }
    }
    contextUsage(id: number, depth: number, usage: Map<number, { reassigned: boolean; referenced: boolean; innerWrite: boolean }>): void {
            const node = this.parser.node(id);
            let target = -1;
            if(node.kind === 'BinaryExpression' && (this.parser.node(node.children[1] ?? -1).kind === 'EqualsToken' || compounds.has(this.parser.node(node.children[1] ?? -1).kind))) { target = node.children[0] ?? -1; }
            if((node.kind === 'PrefixUnaryExpression' || node.kind === 'PostfixUnaryExpression') && (node.operator === 'PlusPlusToken' || node.operator === 'MinusMinusToken')) { target = node.children[0] ?? -1; }
            if(target >= 0 && this.parser.node(target).kind === 'Identifier') {
                const symbol = this.identity(target);
                const entry = usage.get(symbol) ?? { reassigned: false, referenced: false, innerWrite: false };
                entry.reassigned = true; if(depth > 0) { entry.innerWrite = true; } usage.set(symbol, entry);
            }
            if(depth > 0 && node.kind === 'Identifier') {
                const symbol = this.identity(id);
                const entry = usage.get(symbol) ?? { reassigned: false, referenced: false, innerWrite: false };
                entry.referenced = true; usage.set(symbol, entry);
            }
            const next = ['FunctionDeclaration', 'FunctionExpression', 'ArrowFunction', 'MethodDeclaration', 'GetAccessor', 'SetAccessor', 'Constructor'].includes(node.kind) ? depth + 1 : depth;
            for(const child of node.children) { this.contextUsage(child, next, usage); }
    }
    captureOf(symbol: number): PlaceInterface | undefined {
        if(symbol === 0 || this.enclosing === undefined) { return undefined; }
        const old = this.captured.get(symbol); if(old !== undefined) { return old; }
        let scope: StraightLineBuilder | undefined = this.enclosing;
        let found = false;
        while(scope !== undefined) { if(scope.locals.has(symbol)) { found = true; break; } scope = scope.enclosing; }
        if(!found) { return undefined; }
        let name = '';
        if(this.symbols !== undefined) { for(const facts of this.symbols.symbols.values()) { if(facts.identity === symbol) { name = facts.name; break; } } }
        const place = this.fn.named(name, 0, 0);
        this.captured.set(symbol, place); this.captureSymbols.push(symbol); this.fn.context.push(place);
        return place;
    }
    nested(id: number): PlaceInterface {
        const builder = lowerFunction(this.parser, id, this.source, this.symbols, this, this.arena) ?? panic('supported nested function declined');
        const captures: PlaceInterface[] = [];
        for(const symbol of builder.captureSymbols) {
            const place = this.locals.get(symbol) ?? this.captureOf(symbol);
            captures.push(place === undefined ? { identifier: this.fn.returns.identifier, effect: '<unknown>', reactive: false, start: 0, end: 0 } : { identifier: place.identifier, effect: '<unknown>', reactive: false, start: 0, end: 0 });
        }
        const functionId = this.fn.functions.length;
        this.fn.functions.push(builder.functionIndex);
        return this.emit({ kind: 'FunctionExpression', functionReference: { index: builder.functionIndex, ordinal: functionId }, captures }, id, undefined);
    }
    pattern(id: number): PatternIndex {
        const node = this.parser.node(id);
        if(node.kind === 'Identifier') { return this.fn.addPattern({kind: 'Place', place: this.bind(id)}); }
        if(node.kind === 'ObjectBindingPattern' || node.kind === 'ArrayBindingPattern') {
            const properties: PatternPropertyInterface[] = []; const elements: PatternElementInterface[] = []; let rest: PlaceInterface | undefined;
            for(const child of node.children) {
                const element = this.parser.node(child); const parts = element.children.filter((part) => this.parser.node(part).kind !== 'DotDotDotToken');
                if(element.kind === 'OmittedExpression' || parts.length === 0) { elements.push({value: undefined,defaultValue: undefined}); continue; }
                const first = parts[0] ?? -1; const scanner = new Scanner(this.parser.scanner.text); scanner.pos = this.parser.node(first).end; scanner.scan();
                const renamed = scanner.kind === 'ColonToken'; const name = parts[renamed ? 1 : 0] ?? -1;
                if(element.children.some((part) => this.parser.node(part).kind === 'DotDotDotToken')) { if(this.parser.node(name).kind === 'Identifier') { rest = this.bind(name); } continue; }
                const propertyName = renamed ? this.parser.node(first) : this.parser.node(name);
                const computedKey = renamed && propertyName.kind === 'ComputedPropertyName' ? this.expression(propertyName.children[0] ?? -1) : undefined;
                const initializer = declarationInitializer(this.parser,child); const defaultValue = initializer === undefined ? undefined : this.expression(initializer);
                const value = this.pattern(name);
                if(node.kind === 'ObjectBindingPattern') { properties.push({key: computedKey === undefined && (renamed || propertyName.kind === 'Identifier') ? propertyName.text : '',computedKey,defaultValue,value}); }
                else { elements.push({value,defaultValue}); }
            }
            return node.kind === 'ObjectBindingPattern' ? this.fn.addPattern({kind: 'Object',properties,rest}) : this.fn.addPattern({kind: 'Array',elements,rest});
        }
        if(node.kind === 'ObjectLiteralExpression') {
            const properties: PatternPropertyInterface[] = []; let rest: PlaceInterface | undefined;
            for(const child of node.children) {
                const member = this.parser.node(child); const name = member.children[0] ?? -1;
                if(member.kind === 'SpreadAssignment') { if(this.parser.node(name).kind === 'Identifier') { rest = this.bind(name); } continue; }
                if(member.kind !== 'PropertyAssignment' && member.kind !== 'ShorthandPropertyAssignment') { continue; }
                const value = this.pattern(member.kind === 'PropertyAssignment' ? member.children[1] ?? -1 : name);
                const keyNode = this.parser.node(name); const computedKey = keyNode.kind === 'ComputedPropertyName' ? this.expression(keyNode.children[0] ?? -1) : undefined;
                const initial = member.kind === 'ShorthandPropertyAssignment' ? member.children[member.children.length - 1] : undefined;
                const defaultValue = initial !== undefined && initial !== name ? this.expression(initial) : undefined;
                properties.push({key: computedKey === undefined ? keyNode.text : '',computedKey,defaultValue,value});
            }
            return this.fn.addPattern({kind: 'Object',properties,rest});
        }
        if(node.kind === 'ArrayLiteralExpression') {
            const elements: PatternElementInterface[] = []; let rest: PlaceInterface | undefined;
            for(const child of node.children) {
                const item = this.parser.node(child);
                if(item.kind === 'OmittedExpression') { elements.push({value: undefined,defaultValue: undefined}); continue; }
                if(item.kind === 'SpreadElement') { const target = item.children[0] ?? -1; if(this.parser.node(target).kind === 'Identifier') { rest = this.bind(target); } continue; }
                const value = this.pattern(item.kind === 'BinaryExpression' ? item.children[0] ?? -1 : child);
                const defaultValue = item.kind === 'BinaryExpression' ? this.expression(item.children[2] ?? -1) : undefined;
                elements.push({value,defaultValue});
            }
            return this.fn.addPattern({kind: 'Array',elements,rest});
        }
        if(node.kind === 'PropertyAccessExpression' || node.kind === 'ElementAccessExpression') { return this.fn.addPattern({kind: 'Place',place: this.expression(id)}); }
        return this.fn.addPattern({kind: 'Place',place: this.temporary(id)});
    }
    bind(id: number): PlaceInterface {
        const node = this.parser.node(id);
        const symbol = this.identity(id);
        const old = this.locals.get(symbol);
        const declaration = old === undefined || symbol === 0 ? undefined : this.fn.identifier(old.identifier).declaration;
        const place = this.fn.named(node.text, this.byte(node.pos), this.byte(node.end), declaration);
        this.fn.identifier(place.identifier).nodeIndex = id;
        if(symbol !== 0) { this.locals.set(symbol, place); }
        return place;
    }
    loadIdentifier(id: number): PlaceInterface {
        const node = this.parser.node(id);
        if(node.text === 'undefined') { return this.emit({ kind: 'Primitive', literal: 'nil' }, id, undefined); }
        const local = this.locals.get(this.identity(id));
        if(local !== undefined) {
            const place: PlaceInterface = { identifier: local.identifier, effect: '<unknown>', reactive: false, start: this.byte(node.pos), end: this.byte(node.end) };
            return this.emit({ kind: 'LoadLocal', place }, id, undefined);
        }
        const capture = this.captureOf(this.identity(id));
        if(capture !== undefined) { return this.emit({ kind: 'LoadContext', place: { identifier: capture.identifier, effect: '<unknown>', reactive: false, start: this.byte(node.pos), end: this.byte(node.end) } }, id, undefined); }
        let bindingKind = 0; let source = ''; let imported = '';
        if(this.symbols !== undefined) {
            const symbol = this.symbols.read(this.byte(node.pos), this.byte(node.end));
            for(const declaration of symbol.declarations) {
                if(!declaration.sameSource) { continue; }
                if(['ImportSpecifier', 'ImportClause', 'NamespaceImport'].includes(declaration.kind)) {
                    if(declaration.module !== '') {
                        source = declaration.module;
                        bindingKind = declaration.kind === 'ImportSpecifier' ? 4 : declaration.kind === 'NamespaceImport' ? 3 : 2;
                        if(declaration.kind === 'ImportSpecifier') { imported = declaration.property === '' ? declaration.name : declaration.property; }
                        break;
                    }
                }
                else { bindingKind = 1; }
            }
        }
        return this.emit({ kind: 'LoadGlobal', name: node.text, bindingKind, source, imported }, id, undefined);
    }
    assign(id: number, value: PlaceInterface): PlaceInterface {
        const node = this.parser.node(id);
        if(node.kind === 'ParenthesizedExpression') { return this.assign(node.children[0] ?? -1, value); }
        if(node.kind === 'Identifier') {
            const symbol = this.identity(id);
            if(this.locals.has(symbol)) {
                const place = this.bind(id);
                if(this.contextual.has(symbol)) { this.fn.contextDeclarations.add(this.fn.identifier(place.identifier).declaration); }
                this.emit({ kind: this.contextual.has(symbol) ? 'StoreContext' : 'StoreLocal', lvalue: place, value, declarationKind: 2 }, id, undefined);
            }
            else {
                const capture = this.captureOf(symbol);
                if(capture === undefined) { this.emit({ kind: 'StoreGlobal', name: node.text, value }, id, undefined); }
                else { this.emit({ kind: 'StoreContext', lvalue: { identifier: capture.identifier, effect: '<unknown>', reactive: false, start: this.byte(node.pos), end: this.byte(node.end) }, value, declarationKind: 2 }, id, undefined); }
            }
        }
        else if(node.kind === 'ObjectLiteralExpression' || node.kind === 'ArrayLiteralExpression') { const pattern = this.pattern(id); this.emit({kind: 'Destructure', lvaluePattern: pattern, pattern, value, declarationKind: 2},id,undefined); }
        else if(node.kind !== 'PropertyAccessExpression' && node.kind !== 'ElementAccessExpression') { this.emit({kind: 'UnsupportedNode', nodeKind: 'Kind' + node.kind, nodePos: this.byte(node.pos), nodeEnd: this.byte(node.end), reason: 'assignment target'}, id, undefined); }
        else {
            const object = this.expression(node.children[0] ?? -1);
            if(node.kind === 'PropertyAccessExpression') { this.emit({ kind: 'PropertyStore', object, property: this.parser.node(node.children[1] ?? -1).text, value }, id, undefined); }
            else { this.emit({ kind: 'ComputedStore', object, property: this.expression(node.children[1] ?? -1), value }, id, undefined); }
        }
        return value;
    }
    declarations(id: number): void {
        this.declarationList(this.parser.node(id).children[0] ?? -1);
    }
    declarationList(id: number): void {
        const list = this.parser.node(id);
        const declarationKind = (Number.parseInt(list.semantic, 10) & 2) !== 0 ? 0 : 1;
        for(const declarationId of list.children) {
            const declaration = this.parser.node(declarationId);
            const name = declaration.children[0] ?? -1;
            const initializer = declarationInitializer(this.parser, declarationId);
            if(initializer === undefined) { if(this.parser.node(name).kind === 'Identifier') { this.emit({ kind: 'DeclareLocal', lvalue: this.bind(name), declarationKind }, declarationId, undefined); } }
            else {
                const value = this.expression(initializer);
                if(this.parser.node(name).kind === 'Identifier') { this.emit({ kind: 'StoreLocal', lvalue: this.bind(name), value, declarationKind }, declarationId, undefined); } else { const pattern = this.pattern(name); this.emit({kind: 'Destructure', lvaluePattern: pattern, pattern, value, declarationKind}, declarationId, undefined); }
            }
        }
    }
    ensureBlock(): BasicBlock {
        if(this.current === undefined) { this.current = this.fn.newBlock('block'); }
        return this.current;
    }
    close(terminal: TerminalType): void {
        if(this.current !== undefined) { this.current.terminal = terminal; this.current = undefined; }
    }
    jump(block: BlockIndex, variant: number): void { this.close({ kind: 'Goto', block, variant }); }
    statements(ids: readonly number[]): void { for(const id of ids) { this.statement(id); } }
    statement(id: number, label: string = ''): void {
        const node = this.parser.node(id);
        if(node.kind === 'FunctionDeclaration') {
            const value = this.nested(id);
            const nameId = node.children.find((child) => this.parser.node(child).kind === 'Identifier');
            if(nameId !== undefined) { const binding = this.bind(nameId); const place: PlaceInterface = { identifier: binding.identifier, effect: '<unknown>', reactive: false, start: this.byte(node.pos), end: this.byte(node.end) }; this.emit({ kind: 'StoreLocal', lvalue: place, value, declarationKind: 6 }, id, undefined); }
        }
        else if(node.kind === 'Block') { this.statements(node.children); }
        else if(node.kind === 'VariableStatement') { this.declarations(id); }
        else if(node.kind === 'IfStatement') {
            const test = this.expression(node.children[0] ?? -1);
            const consequent = this.fn.newBlock('block'); const fallthrough = this.fn.newBlock('block');
            const alternate = node.children[2] === undefined ? fallthrough : this.fn.newBlock('block');
            this.close({ kind: 'If', testPlace: test, consequent: consequent.id, alternate: alternate.id, fallthrough: fallthrough.id });
            this.current = consequent; this.statement(node.children[1] ?? -1); this.jump(fallthrough.id, 0);
            if(node.children[2] !== undefined) { this.current = alternate; this.statement(node.children[2] ?? -1); this.jump(fallthrough.id, 0); }
            this.current = fallthrough;
        }
        else if(node.kind === 'WhileStatement') {
            const test = this.fn.newBlock('block'); const loop = this.fn.newBlock('loop'); const fallthrough = this.fn.newBlock('block');
            this.close({ kind: 'While', testBlock: test.id, loop: loop.id, fallthrough: fallthrough.id });
            this.current = test;
            const value = this.expression(node.children[0] ?? -1);
            this.close({ kind: 'Branch', testPlace: value, consequent: loop.id, alternate: fallthrough.id, fallthrough: fallthrough.id });
            this.current = loop; this.jumps.push({ label, breakBlock: fallthrough.id, continueBlock: test.id });
            this.statement(node.children[1] ?? -1); this.jumps.pop(); this.jump(test.id, 1);
            this.current = fallthrough;
        }
        else if(['DoStatement', 'ForStatement', 'ForOfStatement', 'ForInStatement', 'SwitchStatement', 'LabeledStatement', 'TryStatement'].includes(node.kind)) { this.flow(id, label); }
        else if(node.kind === 'BreakStatement' || node.kind === 'ContinueStatement') {
            const name = node.children[0] === undefined ? '' : this.parser.node(node.children[0] ?? -1).text;
            const target = this.lookupJump(name, node.kind === 'ContinueStatement');
            if(target === undefined) { this.close({ kind: 'Unsupported' }); }
            else { this.jump(node.kind === 'BreakStatement' ? target.breakBlock : target.continueBlock ?? panic('invalid continue target'), node.kind === 'BreakStatement' ? 0 : 1); }
        }
        else if(node.kind === 'ReturnStatement') {
            const expressionId = node.children[0];
            const result = expressionId === undefined ? undefined : this.expression(expressionId);
            const value: ValueType = result === undefined ? { kind: 'Primitive', literal: 'nil' } : { kind: 'LoadLocal', place: result };
            this.emit(value, id, this.fn.returns);
            this.close({ kind: 'Return', value: this.fn.returns });
        }
        else if(node.kind === 'ThrowStatement') { this.close({ kind: 'Throw', value: this.expression(node.children[0] ?? -1) }); }
        else if(node.kind === 'ExpressionStatement') { this.expression(node.children[0] ?? -1); }
        else if(node.kind === 'DebuggerStatement') { this.emit({ kind: 'Debugger' }, id, undefined); }
        else if(!['EmptyStatement', 'InterfaceDeclaration', 'TypeAliasDeclaration'].includes(node.kind)) { this.unsupported(id); }
    }
    lookupJump(label: string, continuing: boolean): { readonly label: string; readonly breakBlock: BlockIndex; readonly continueBlock?: BlockIndex } | undefined {
        for(let index = this.jumps.length - 1; index >= 0; index--) {
            const target = this.jumps[index] ?? panic('missing jump');
            if(continuing && target.continueBlock === undefined) { if(label !== '' && target.label === label) { return undefined; } continue; }
            if(label === '' || target.label === label) { return target; }
        }
        return undefined;
    }
    forBinding(id: number, value: PlaceInterface): void {
        const node = this.parser.node(id);
        if(node.kind !== 'VariableDeclarationList') { this.assign(id, value); return; }
        for(const child of node.children) { const name = this.parser.node(child).children[0] ?? -1; const declarationKind = (Number.parseInt(node.semantic,10) & 2) !== 0 ? 0 : 1; if(this.parser.node(name).kind === 'Identifier') { this.emit({ kind: 'StoreLocal', lvalue: this.bind(name), value, declarationKind }, name, undefined); } else { const pattern = this.pattern(name); this.emit({kind: 'Destructure', lvaluePattern: pattern, pattern, value, declarationKind},name,undefined); } }
    }
    flow(id: number, label: string): void {
        const node = this.parser.node(id);
        if(node.kind === 'LabeledStatement') {
            const name = this.parser.node(node.children[0] ?? -1).text; const body = node.children[1] ?? -1;
            if(['WhileStatement', 'DoStatement', 'ForStatement', 'ForOfStatement', 'ForInStatement'].includes(this.parser.node(body).kind)) { this.statement(body, name); return; }
            const block = this.fn.newBlock('block'); const fallthrough = this.fn.newBlock('block');
            this.close({ kind: 'Label', block: block.id, fallthrough: fallthrough.id }); this.current = block;
            this.jumps.push({ label: name, breakBlock: fallthrough.id }); this.statement(body); this.jumps.pop();
            this.jump(fallthrough.id, 0); this.current = fallthrough; return;
        }
        if(node.kind === 'SwitchStatement') {
            const test = this.expression(node.children[0] ?? -1); const clauses = this.parser.node(node.children[1] ?? -1).children;
            const fallthrough = this.fn.newBlock('block'); const blocks = clauses.map((_child) => this.fn.newBlock('block').id);
            const cases: { test: PlaceInterface | undefined; readonly block: BlockIndex }[] = [];
            for(let index = 0; index < clauses.length; index++) { const clause = this.parser.node(clauses[index] ?? -1); cases.push({ test: clause.kind === 'DefaultClause' ? undefined : this.expression(clause.children[0] ?? -1), block: blocks[index] ?? panic('missing case block') }); }
            this.close({ kind: 'Switch', testPlace: test, cases, fallthrough: fallthrough.id }); this.jumps.push({ label, breakBlock: fallthrough.id });
            for(let index = 0; index < clauses.length; index++) { const clause = this.parser.node(clauses[index] ?? -1); this.current = this.fn.block(blocks[index] ?? panic('missing case block')); this.statements(clause.children.slice(clause.kind === 'CaseClause' ? 1 : 0)); this.jump(blocks[index + 1] ?? fallthrough.id, 0); }
            this.jumps.pop(); this.current = fallthrough; return;
        }
        if(node.kind === 'TryStatement') { this.tryStatement(id); return; }
        if(node.kind === 'DoStatement') {
            const loop = this.fn.newBlock('loop'); const test = this.fn.newBlock('block'); const fallthrough = this.fn.newBlock('block');
            this.close({ kind: 'DoWhile', loop: loop.id, testBlock: test.id, fallthrough: fallthrough.id }); this.current = loop;
            this.jumps.push({ label, breakBlock: fallthrough.id, continueBlock: test.id }); this.statement(node.children[0] ?? -1); this.jumps.pop();
            this.jump(test.id, 1); this.current = test;
            const value = this.expression(node.children[1] ?? -1); this.close({ kind: 'Branch', testPlace: value, consequent: loop.id, alternate: fallthrough.id, fallthrough: fallthrough.id }); this.current = fallthrough; return;
        }
        if(node.kind === 'ForStatement') {
            const slots = forHeaderSlots(this.parser, id);
            const init = this.fn.newBlock('block'); const test = this.fn.newBlock('block'); const loop = this.fn.newBlock('loop'); const fallthrough = this.fn.newBlock('block');
            const increment = slots[2] ?? -1; const update = increment < 0 ? undefined : this.fn.newBlock('block');
            if(update === undefined) { this.close({ kind: 'For', init: init.id, testBlock: test.id, loop: loop.id, fallthrough: fallthrough.id }); } else { this.close({ kind: 'For', init: init.id, testBlock: test.id, loop: loop.id, update: update.id, fallthrough: fallthrough.id }); } this.current = init;
            const initializer = slots[0] ?? -1; if(initializer >= 0) { if(this.parser.node(initializer).kind === 'VariableDeclarationList') { this.declarationList(initializer); } else { this.expression(initializer); } }
            this.jump(test.id, 0); this.current = test; const condition = slots[1] ?? -1;
            if(condition < 0) { this.jump(loop.id, 0); } else { const value = this.expression(condition); this.close({ kind: 'Branch', testPlace: value, consequent: loop.id, alternate: fallthrough.id, fallthrough: fallthrough.id }); }
            const continueBlock = update?.id ?? test.id; this.current = loop; this.jumps.push({ label, breakBlock: fallthrough.id, continueBlock }); this.statement(slots[3] ?? -1); this.jumps.pop();
            this.jump(continueBlock, 1); if(update !== undefined) { this.current = update; this.expression(increment); this.jump(test.id, 0); }
            this.current = fallthrough; return;
        }
        const of = node.kind === 'ForOfStatement'; const offset = this.parser.node(node.children[0] ?? -1).kind === 'AwaitKeyword' ? 1 : 0;
        const init = this.fn.newBlock(of ? 'loop' : 'block'); const test = of ? this.fn.newBlock('loop') : init; const loop = this.fn.newBlock('loop'); const fallthrough = this.fn.newBlock('block');
        if(of) { this.close({ kind: 'ForOf', init: init.id, testBlock: test.id, loop: loop.id, fallthrough: fallthrough.id }); } else { this.close({ kind: 'ForIn', init: init.id, loop: loop.id, fallthrough: fallthrough.id }); }
        this.current = init; const collectionId = node.children[offset + 1] ?? -1; const collection = this.expression(collectionId);
        const iterator = of ? this.emit({ kind: 'GetIterator', value: collection }, collectionId, undefined) : undefined;
        let next: PlaceInterface;
        if(iterator !== undefined) { this.jump(test.id, 0); this.current = test; next = this.emit({ kind: 'IteratorNext', iterator, collection }, id, undefined); }
        else { next = this.emit({ kind: 'NextPropertyOf', value: collection }, id, undefined); }
        this.close({ kind: 'Branch', testPlace: next, consequent: loop.id, alternate: fallthrough.id, fallthrough: fallthrough.id });
        this.current = loop; this.forBinding(node.children[offset] ?? -1, next); this.jumps.push({ label, breakBlock: fallthrough.id, continueBlock: test.id }); this.statement(node.children[offset + 2] ?? -1); this.jumps.pop();
        this.jump(test.id, 1); this.current = fallthrough;
    }
    tryStatement(id: number): void {
        const node = this.parser.node(id); const body = this.fn.newBlock('block'); const fallthrough = this.fn.newBlock('block');
        const catchId = node.children.find((child) => this.parser.node(child).kind === 'CatchClause');
        const finallyId = node.children.slice(1).find((child) => this.parser.node(child).kind === 'Block');
        const finalBlock = finallyId === undefined ? undefined : this.fn.newBlock('block'); const normal = finalBlock ?? fallthrough; const handler = this.fn.newBlock('catch');
        let binding: PlaceInterface | undefined;
        if(catchId !== undefined) { const clause = this.parser.node(catchId); if(clause.children.length > 1) { const name = this.parser.node(clause.children[0] ?? -1).children[0] ?? -1; if(this.parser.node(name).kind === 'Identifier') { binding = this.bind(name); } } }
        this.close({ kind: 'Try', block: body.id, handler: handler.id, handlerBinding: binding, fallthrough: fallthrough.id }); this.current = body; this.statement(node.children[0] ?? -1); this.jump(normal.id, 2);
        this.current = handler;
        if(catchId !== undefined) { const clause = this.parser.node(catchId); if(clause.children.length > 1 && binding === undefined) { const name = this.parser.node(clause.children[0] ?? -1).children[0] ?? -1; const nameNode = this.parser.node(name); const temporary = this.temporary(name); const pattern = this.pattern(name); this.emit({kind: 'Destructure', lvaluePattern: pattern, pattern, value: temporary, declarationKind: 3},name,undefined); } this.statement(clause.children[clause.children.length - 1] ?? -1); this.jump(normal.id, 2); }
        else if(finalBlock !== undefined) { this.jump(finalBlock.id, 2); } else { this.close({ kind: 'Unreachable' }); }
        if(finalBlock !== undefined) { this.current = finalBlock; this.statement(finallyId ?? -1); this.jump(fallthrough.id, 2); }
        this.current = fallthrough;
    }
    valueBranch(id: number, logical: string | undefined): PlaceInterface {
        const node = this.parser.node(id);
        const result = this.temporary(id);
        const testBlock = this.fn.newBlock('value'); const fallthrough = this.fn.newBlock('block');
        if(logical === undefined) { this.close({ kind: 'Ternary', testBlock: testBlock.id, fallthrough: fallthrough.id }); }
        else { this.close({ kind: 'Logical', operator: logical, testBlock: testBlock.id, fallthrough: fallthrough.id }); }
        this.current = testBlock;
        const left = this.expression(node.children[0] ?? -1);
        const first = this.fn.newBlock('value'); const second = this.fn.newBlock('value');
        const swapped = logical === '||' || logical === '??';
        // A logical's first arm is the short-circuit arm; a ternary's is the true arm.
        const consequent = logical === undefined || swapped ? first : second;
        const alternate = logical === undefined || swapped ? second : first;
        this.close({ kind: 'Branch', testPlace: left, consequent: consequent.id, alternate: alternate.id, fallthrough: fallthrough.id });
        const shared = this.fn.identifier(result.identifier).declaration;
        this.current = first;
        const firstNode = node.children[logical === undefined ? 2 : 0] ?? -1;
        const firstValue = logical === undefined ? this.expression(firstNode) : left;
        const firstPlace = logical === undefined ? result : this.temporary(firstNode,shared);
        this.emit({ kind: 'LoadLocal', place: firstValue }, firstNode, firstPlace); this.jump(fallthrough.id, 0);
        this.current = second;
        const secondNode = node.children[logical === undefined ? 4 : 2] ?? -1;
        const secondValue = this.expression(secondNode);
        const secondPlace = logical === undefined ? result : this.temporary(secondNode,shared);
        this.emit({ kind: 'LoadLocal', place: secondValue }, secondNode, secondPlace); this.jump(fallthrough.id, 0);
        this.current = fallthrough;
        return result;
    }
    temporary(id: number, declaration: DeclarationIndex | undefined = undefined): PlaceInterface {
        if(id < 0) { return this.fn.temporary(0,0,declaration); }
        const node = this.parser.node(id); const place = this.fn.temporary(this.byte(node.pos),this.byte(node.end),declaration);
        this.fn.identifier(place.identifier).nodeIndex = id; return place;
    }
    byte(index: number): number { return utf8Length(this.source.slice(0, index)); }
    emit(value: ValueType, id: number, target: PlaceInterface | undefined): PlaceInterface {
        const node = this.parser.node(id);
        const start = this.byte(node.pos);
        const end = this.byte(node.end);
        const place = target ?? this.fn.temporary(start, end);
        if(target === undefined) { this.fn.identifier(place.identifier).nodeIndex = id; }
        const block = this.ensureBlock();
        block.instructions.push(this.fn.emit(place, value, start, end));
        return place;
    }
    resolvedOrigin(id: number, resolver: ExportResolver): OriginInterface {
        const node = this.parser.node(id);
        if(['ParenthesizedExpression', 'AsExpression', 'TypeAssertionExpression', 'NonNullExpression', 'SatisfiesExpression'].includes(node.kind)) { return this.resolvedOrigin(node.children[0] ?? -1, resolver); }
        if(node.kind === 'Identifier') { if(this.symbols === undefined) { return emptyOrigin; } const index = this.symbols.read(this.byte(node.pos), this.byte(node.end)).identity; return resolver.symbol(resolver.graph.symbolReference(index)); }
        if(node.kind === 'PropertyAccessExpression' || node.kind === 'ElementAccessExpression') {
            const receiver = this.resolvedOrigin(node.children[0] ?? -1, resolver); const property = this.parser.node(node.children[node.children.length - 1] ?? -1);
            if(receiver.module !== undefined && (receiver.name === '*' || receiver.name === 'default') && property.text !== '' && (node.kind === 'PropertyAccessExpression' || property.kind === 'StringLiteral' || property.kind === 'NoSubstitutionTemplateLiteral')) { return resolver.export(receiver.module, property.text); }
        }
        return emptyOrigin;
    }
    // Resolve direct imports and same-file const aliases from compiler symbol identities.
    // Barrel export graphs remain a separate construction seam.
    exportOrigin(id: number, active: Set<number>): ModuleExportOriginInterface {
        if(this.symbols?.graph !== undefined) {
            const resolver = new ExportResolver(this.symbols.graph, 'react'); const origin = this.resolvedOrigin(id, resolver);
            return origin.module !== undefined && resolver.graph.module(origin.module).name === 'react' ? { module: 'react', exported: origin.name } : { module: '', exported: '' };
        }
        const empty: ModuleExportOriginInterface = { module: '', exported: '' };
        if(this.symbols === undefined) { return empty; }
        const node = this.parser.node(id);
        if(node.kind === 'ParenthesizedExpression') { return this.exportOrigin(node.children[0] ?? -1, active); }
        if(node.kind === 'PropertyAccessExpression' || node.kind === 'ElementAccessExpression') {
            const receiver = this.exportOrigin(node.children[0] ?? -1, active);
            const property = this.parser.node(node.children[node.children.length - 1] ?? -1);
            if(receiver.module !== '' && (receiver.exported === '*' || receiver.exported === 'default') && (node.kind === 'PropertyAccessExpression' || property.kind === 'StringLiteral')) {
                return { module: receiver.module, exported: property.text };
            }
            return empty;
        }
        if(node.kind !== 'Identifier') { return empty; }
        const symbol = this.symbols.read(this.byte(node.pos), this.byte(node.end));
        if(active.has(symbol.identity)) { return empty; }
        active.add(symbol.identity);
        let result = empty;
        for(const declaration of symbol.declarations) {
            if(['ImportSpecifier', 'ImportClause', 'NamespaceImport'].includes(declaration.kind) && declaration.module === 'react') {
                result = { module: 'react', exported: declaration.kind === 'NamespaceImport' ? '*' : declaration.kind === 'ImportClause' ? 'default' : declaration.property === '' ? declaration.name : declaration.property };
                break;
            }
            if(declaration.kind === 'VariableDeclaration' && declaration.sameSource) {
                const index = this.parser.nodes.findIndex((candidate) => candidate.kind === 'VariableDeclaration' && this.byte(candidate.pos) === declaration.start && this.byte(candidate.end) === declaration.end);
                if(index >= 0) {
                    const parent = this.parser.nodes.find((candidate) => candidate.kind === 'VariableDeclarationList' && candidate.children.includes(index));
                    const declarationNode = this.parser.node(index);
                    const initializer = declarationNode.children.length > 1 ? declarationNode.children[declarationNode.children.length - 1] : undefined;
                    if(parent !== undefined && (Number.parseInt(parent.semantic, 10) & 2) !== 0 && initializer !== undefined) { result = this.exportOrigin(initializer, active); if(result.module !== '') { break; } }
                }
            }
        }
        active.delete(symbol.identity);
        return result;
    }
    arguments(id: number): ArgumentInterface[] {
        const node = this.parser.node(id);
        const args: ArgumentInterface[] = [];
        const count = Math.max(0, node.list);
        for(const argument of node.children.slice(node.children.length - count)) {
            const child = this.parser.node(argument);
            args.push({ place: this.expression(child.kind === 'SpreadElement' ? child.children[0] ?? -1 : argument), spread: child.kind === 'SpreadElement' });
        }
        return args;
    }
    call(id: number): PlaceInterface {
        const node = this.parser.node(id);
        const calleeId = node.children[0] ?? -1;
        if(node.kind === 'NewExpression') {
            const callee = this.expression(calleeId);
            return this.emit({ kind: 'NewExpression', callee, args: this.arguments(id) }, id, undefined);
        }
        const origin = this.exportOrigin(calleeId, new Set<number>());
        const optional = node.children.some((child) => this.parser.node(child).kind === 'QuestionDotToken');
        const callee = this.parser.node(calleeId);
        if(callee.kind === 'PropertyAccessExpression' || callee.kind === 'ElementAccessExpression') {
            const receiver = this.expression(callee.children[0] ?? -1);
            const propertyId = callee.children[callee.children.length - 1] ?? -1;
            const property = callee.kind === 'PropertyAccessExpression' ? this.emit({ kind: 'Primitive', literal: `string:${written(this.parser.node(propertyId).text)}` }, propertyId, undefined) : this.expression(propertyId);
            return this.emit({ kind: 'MethodCall', receiver, property, args: this.arguments(id), optional, origin }, id, undefined);
        }
        const place = this.expression(calleeId);
        return this.emit({ kind: 'CallExpression', callee: place, args: this.arguments(id), optional, origin }, id, undefined);
    }
    aggregate(id: number): PlaceInterface {
        const node = this.parser.node(id);
        if(node.kind === 'ArrayLiteralExpression') {
            const elements: ArrayElementInterface[] = [];
            for(const child of node.children) {
                const element = this.parser.node(child);
                if(element.kind === 'OmittedExpression') { elements.push({ place: { identifier: this.fn.returns.identifier, effect: '<unknown>', reactive: false, start: 0, end: 0 }, spread: false, hole: true }); }
                else { elements.push({ place: this.expression(element.kind === 'SpreadElement' ? element.children[0] ?? -1 : child), spread: element.kind === 'SpreadElement', hole: false }); }
            }
            return this.emit({ kind: 'ArrayExpression', elements }, id, undefined);
        }
        const properties: ObjectPropertyInterface[] = [];
        for(const child of node.children) {
            const member = this.parser.node(child);
            if(member.kind === 'SpreadAssignment') { properties.push({ key: '', computedKey: undefined, value: this.expression(member.children[0] ?? -1), spread: true }); continue; }
            const method = ['MethodDeclaration', 'GetAccessor', 'SetAccessor'].includes(member.kind);
            const nameId = method ? member.children.find((part) => ['Identifier', 'StringLiteral', 'NumericLiteral', 'ComputedPropertyName', 'PrivateIdentifier', 'NoSubstitutionTemplateLiteral'].includes(this.parser.node(part).kind)) ?? -1 : member.children[0] ?? -1; const name = this.parser.node(nameId);
            // Go lowers the initializer before a computed key, including effects and block splits.
            let value: PlaceInterface;
            if(method) {
                const builder = lowerFunction(this.parser, child, this.source, this.symbols, this, this.arena) ?? panic('object method declined');
                const ordinal = this.fn.functions.length; this.fn.functions.push(builder.functionIndex);
                value = this.emit({ kind: 'ObjectMethod', key: name.kind === 'ComputedPropertyName' ? '' : name.text, functionReference: { index: builder.functionIndex, ordinal } }, child, undefined);
            } else if(member.kind === 'PropertyAssignment' || member.kind === 'ShorthandPropertyAssignment') { value = this.expression(member.kind === 'ShorthandPropertyAssignment' ? nameId : member.children[1] ?? -1); }
            else { properties.push({key: '', computedKey: undefined, value: this.emit({kind: 'UnsupportedNode', nodeKind: 'Kind' + member.kind, nodePos: this.byte(member.pos), nodeEnd: this.byte(member.end), reason: 'object member'}, child, undefined), spread: false}); continue; }
            const computedKey = name.kind === 'ComputedPropertyName' ? this.expression(name.children[0] ?? -1) : undefined;
            properties.push({ key: computedKey === undefined ? name.text : '', computedKey, value, spread: false });
        }
        return this.emit({ kind: 'ObjectExpression', properties }, id, undefined);
    }
    jsxChildren(children: readonly number[]): PlaceInterface[] {
        const places: PlaceInterface[] = [];
        for(const id of children) {
            const child = this.parser.node(id);
            if(child.kind === 'JsxText') { if(child.semantic !== '1') { places.push(this.emit({ kind: 'JsxText', text: child.text }, id, undefined)); } }
            else if(child.kind === 'JsxExpression') { const inner = child.children.find((value) => this.parser.node(value).kind !== 'DotDotDotToken'); if(inner !== undefined) { places.push(this.expression(inner)); } }
            else { places.push(this.expression(id)); }
        }
        return places;
    }
    jsx(id: number): PlaceInterface {
        const node = this.parser.node(id);
        const children: number[] = node.kind === 'JsxSelfClosingElement' ? [] : node.children.slice(1, node.children.length - 1);
        if(node.kind === 'JsxFragment') { return this.emit({ kind: 'JsxFragment', children: this.jsxChildren(children) }, id, undefined); }
        const opening = node.kind === 'JsxElement' ? this.parser.node(node.children[0] ?? -1) : node;
        const tagId = opening.children[0] ?? -1; const name = this.parser.node(tagId); let tag: JsxTagInterface = { name: 'unknown', place: undefined };
        if(name.kind === 'Identifier') { if(name.text !== '' && name.text.charAt(0) >= 'a' && name.text.charAt(0) <= 'z') { tag = { name: name.text, place: undefined }; } else { tag = { name: '', place: this.loadIdentifier(tagId) }; } }
        else if(name.kind === 'PropertyAccessExpression') { tag = { name: '', place: this.expression(tagId) }; }
        const attributes = this.parser.node(opening.children[opening.children.length - 1] ?? -1); const props: JsxAttributeInterface[] = [];
        for(const child of attributes.children) {
            const attribute = this.parser.node(child);
            if(attribute.kind === 'JsxSpreadAttribute') { props.push({ name: '', value: this.expression(attribute.children[0] ?? -1), spread: true }); }
            else {
                const name = this.parser.node(attribute.children[0] ?? -1); const initializer = attribute.children[1];
                const value = initializer === undefined ? this.emit({ kind: 'Primitive', literal: 'bool:true' }, child, undefined) : this.expression(initializer);
                props.push({ name: name.kind === 'Identifier' ? name.text : '', value, spread: false });
            }
        }
        return this.emit({ kind: 'JsxExpression', tag, props, children: this.jsxChildren(children) }, id, undefined);
    }
    optionalLink(id: number): boolean {
        let cursor = id;
        while(cursor >= 0) { const node = this.parser.node(cursor); if(node.kind !== 'PropertyAccessExpression' && node.kind !== 'ElementAccessExpression') { return false; } if(node.children.some((child) => this.parser.node(child).kind === 'QuestionDotToken')) { return true; } cursor = node.children[0] ?? -1; }
        return false;
    }
    optionalChain(id: number, parentAlternate: BasicBlock | undefined): PlaceInterface {
        const node = this.parser.node(id);
        const result = this.temporary(id);
        const continuation = this.fn.newBlock(this.current?.kind ?? 'value');
        let alternate = parentAlternate;
        if(alternate === undefined) {
            alternate = this.fn.newBlock('value'); const saved = this.current; this.current = alternate;
            const empty = this.emit({ kind: 'Primitive', literal: 'nil' }, id, undefined);
            this.emit({ kind: 'StoreLocal', lvalue: result, value: empty, declarationKind: 0 }, id, undefined); this.jump(continuation.id, 0); this.current = saved;
        }
        const test = this.fn.newBlock('value');
        // gap 1 (GAPS.md): keep presence beside the value until mixed-union lowering lands.
        this.close({ kind: 'Optional', optionalFlag: { present: true, value: node.children.some((child) => this.parser.node(child).kind === 'QuestionDotToken') }, testBlock: test.id, fallthrough: continuation.id }); this.current = test;
        const objectId = node.children[0] ?? -1;
        const object = this.optionalLink(objectId) ? this.optionalChain(objectId, alternate) : this.expression(objectId);
        const consequent = this.fn.newBlock('value'); this.close({ kind: 'Branch', testPlace: object, consequent: consequent.id, alternate: alternate.id, fallthrough: continuation.id }); this.current = consequent;
        const key = node.children[node.children.length - 1] ?? -1;
        const loaded = node.kind === 'PropertyAccessExpression' ? this.emit({ kind: 'PropertyLoad', object, property: this.parser.node(key).text, optional: false }, id, undefined) : this.emit({ kind: 'ComputedLoad', object, property: this.expression(key), optional: false }, id, undefined);
        this.emit({ kind: 'StoreLocal', lvalue: result, value: loaded, declarationKind: 0 }, id, undefined); this.jump(continuation.id, 0); this.current = continuation;
        return result;
    }
    template(id: number): PlaceInterface {
        const node = this.parser.node(id);
        const tag = node.kind === 'TaggedTemplateExpression' ? this.expression(node.children[0] ?? -1) : undefined;
        const template = tag === undefined ? node : this.parser.node(node.children[node.children.length - 1] ?? -1);
        const quasis: string[] = [];
        const subexprs: PlaceInterface[] = [];
        if(template.kind === 'TemplateExpression') {
            quasis.push(this.parser.node(template.children[0] ?? -1).text);
            for(const child of template.children.slice(1)) { const span = this.parser.node(child); subexprs.push(this.expression(span.children[0] ?? -1)); quasis.push(this.parser.node(span.children[1] ?? -1).text); }
        } else { quasis.push(template.text); }
        return tag === undefined ? this.emit({ kind: 'TemplateLiteral', quasis, subexprs }, id, undefined) : this.emit({ kind: 'TaggedTemplateExpression', tag, quasis, subexprs }, id, undefined);
    }
    unsupported(id: number): PlaceInterface {
        const node = this.parser.node(id);
        return this.emit({ kind: 'UnsupportedNode', nodeKind: 'Kind' + node.kind, nodePos: this.byte(node.pos), nodeEnd: this.byte(node.end), reason: 'Kind' + node.kind }, id, undefined);
    }
    expression(id: number): PlaceInterface {
        const node = this.parser.node(id);
        if(node.kind === 'AsExpression' || node.kind === 'SatisfiesExpression' || node.kind === 'TypeAssertionExpression') { const type = this.parser.node(node.children[node.kind === 'TypeAssertionExpression' ? 0 : 1] ?? -1); return this.emit({ kind: 'TypeCastExpression', value: this.expression(node.children[node.kind === 'TypeAssertionExpression' ? 1 : 0] ?? -1), nodeKind: type.kind, nodePos: this.byte(type.pos), nodeEnd: this.byte(type.end) }, id, undefined); }
        if(node.kind === 'AwaitExpression') { return this.emit({ kind: 'Await', value: this.expression(node.children[0] ?? -1) }, id, undefined); }
        if(node.kind === 'DeleteExpression') { const targetId = node.children[0] ?? -1; const target = this.parser.node(targetId); if(target.kind === 'PropertyAccessExpression') { return this.emit({ kind: 'PropertyDelete', object: this.expression(target.children[0] ?? -1), property: this.parser.node(target.children[target.children.length - 1] ?? -1).text }, id, undefined); } if(target.kind === 'ElementAccessExpression') { return this.emit({ kind: 'ComputedDelete', object: this.expression(target.children[0] ?? -1), property: this.expression(target.children[target.children.length - 1] ?? -1) }, id, undefined); } this.expression(targetId); return this.emit({ kind: 'Primitive', literal: 'bool:true' }, id, undefined); }
        if(node.kind === 'MetaProperty') { return this.emit({ kind: 'MetaProperty', meta: 'import', property: 'meta' }, id, undefined); }
        if(node.kind === 'SpreadElement') { return this.expression(node.children[0] ?? -1); }
        if(node.kind === 'ThisKeyword') { return this.emit({ kind: 'LoadGlobal', name: 'this', bindingKind: 0, source: '', imported: '' }, id, undefined); }
        if(node.kind === 'RegularExpressionLiteral') { const slash = node.text.lastIndexOf('/'); return this.emit({ kind: 'RegExpLiteral', pattern: slash <= 0 ? node.text : node.text.slice(1, slash), flags: slash <= 0 ? '' : node.text.slice(slash + 1) }, id, undefined); }
        if(node.kind === 'TemplateExpression' || node.kind === 'TaggedTemplateExpression') { return this.template(id); }
        if(node.kind === 'NonNullExpression') { return this.expression(node.children[0] ?? -1); }
        if(node.kind === 'ArrayLiteralExpression' || node.kind === 'ObjectLiteralExpression') { return this.aggregate(id); }
        if(node.kind === 'JsxElement' || node.kind === 'JsxSelfClosingElement' || node.kind === 'JsxFragment') { return this.jsx(id); }
        if(node.kind === 'JsxExpression') { const inner = node.children.find((child) => this.parser.node(child).kind !== 'DotDotDotToken'); return inner === undefined ? this.emit({ kind: 'Primitive', literal: 'nil' }, id, undefined) : this.expression(inner); }
        if(node.kind === 'FunctionExpression' || node.kind === 'ArrowFunction') { return this.nested(id); }
        const primitive = literal(node.kind, node.text);
        if(primitive !== undefined) { return this.emit({ kind: 'Primitive', literal: primitive }, id, undefined); }
        if(node.kind === 'CallExpression' || node.kind === 'NewExpression') { return this.call(id); }
        if(node.kind === 'ConditionalExpression') { return this.valueBranch(id, undefined); }
        if(node.kind === 'Identifier') { return this.loadIdentifier(id); }
        if(node.kind === 'PropertyAccessExpression' || node.kind === 'ElementAccessExpression') {
            if(node.kind === 'PropertyAccessExpression' && this.optionalLink(id)) { return this.optionalChain(id, undefined); }
            const object = this.expression(node.children[0] ?? -1);
            if(node.kind === 'PropertyAccessExpression') { return this.emit({ kind: 'PropertyLoad', object, property: this.parser.node(node.children[node.children.length - 1] ?? -1).text, optional: node.children.some((child) => this.parser.node(child).kind === 'QuestionDotToken') }, id, undefined); }
            return this.emit({ kind: 'ComputedLoad', object, property: this.expression(node.children[node.children.length - 1] ?? -1), optional: node.children.some((child) => this.parser.node(child).kind === 'QuestionDotToken') }, id, undefined);
        }
        if(node.kind === 'PostfixUnaryExpression' || (node.kind === 'PrefixUnaryExpression' && (node.operator === 'PlusPlusToken' || node.operator === 'MinusMinusToken'))) {
            const operand = node.children[0] ?? -1;
            const current = this.expression(operand);
            const operation = node.operator === 'PlusPlusToken' ? '++' : '--';
            const result = this.emit({ kind: node.kind === 'PrefixUnaryExpression' ? 'PrefixUpdate' : 'PostfixUpdate', lvalue: current, operation, value: current }, id, undefined);
            this.assign(operand, result);
            return result;
        }
        if(node.kind === 'ParenthesizedExpression') { return this.expression(node.children[0] ?? -1); }
        if(node.kind === 'BinaryExpression') {
            const operator = this.parser.node(node.children[1] ?? -1).kind;
            const logical = logicals.get(operator);
            if(logical !== undefined) { return this.valueBranch(id, logical); }
            if(operator === 'EqualsToken') { return this.assign(node.children[0] ?? -1, this.expression(node.children[2] ?? -1)); }
            const left = this.expression(node.children[0] ?? -1);
            const right = this.expression(node.children[2] ?? -1);
            if(operator === 'CommaToken') { return right; }
            const compound = compounds.get(operator);
            if(compound !== undefined) {
                const updated = this.emit({ kind: 'BinaryExpression', left, operator: compound, right }, id, undefined);
                return this.assign(node.children[0] ?? -1, updated);
            }
            return this.emit({ kind: 'BinaryExpression', left, operator: operators.get(operator) ?? panic('unsupported binary'), right }, id, undefined);
        }
        if(!['TypeOfExpression', 'VoidExpression', 'PrefixUnaryExpression'].includes(node.kind)) { return this.unsupported(id); }
        const value = this.expression(node.children[0] ?? -1);
        const operator = node.kind === 'TypeOfExpression' ? 'typeof' : node.kind === 'VoidExpression' ? 'void' : operators.get(node.operator) ?? panic('unsupported unary');
        return this.emit({ kind: 'UnaryExpression', operator, value }, id, undefined);
    }
}
function lowerFunction(parser: Parser, root: number, source: string, symbols: SymbolSnapshot | undefined, enclosing: StraightLineBuilder | undefined, arena: HIRArena): StraightLineBuilder | undefined {
    const node = parser.node(root);
    let name = '';
    let bodyId = -1;
    let conciseId = -1;
    const parameters: number[] = [];
    const body = node.children[node.children.length - 1] ?? -1;
    if(parser.node(body).kind === 'Block') { bodyId = body; }
    else if(node.kind === 'ArrowFunction') { conciseId = body; }
    for(const id of node.children.slice(0, node.children.length - 1)) {
        const child = parser.node(id);
        if(['Identifier', 'StringLiteral', 'NumericLiteral', 'PrivateIdentifier', 'NoSubstitutionTemplateLiteral'].includes(child.kind) && node.kind !== 'ArrowFunction') { name = child.text; }
        if(child.kind === 'Parameter') { const binding = parameterBinding(parser, id); parameters.push(binding); }
    }
    if(bodyId < 0 && conciseId < 0) { return undefined; }
    if(name === '') {
        const parent = parser.nodes.find((candidate) => candidate.children.includes(root));
        if(parent?.kind === 'VariableDeclaration') {
            const binding = parser.node(parent.children[0] ?? -1);
            if(binding.kind === 'Identifier') { name = binding.text; }
        }
    }
    const statements: number[] = bodyId < 0 ? [] : parser.node(bodyId).children;
    for(const id of statements) { if(!supportedStatement(parser, id)) { return undefined; } }
    const functionIndex = arena.create(name);
    const fn = arena.read(functionIndex);
    fn.nodeIndex = root;
    fn.isAsync = node.children.some((child) => parser.node(child).kind === 'AsyncKeyword');
    fn.isGenerator = node.children.some((child) => parser.node(child).kind === 'AsteriskToken');
    const builder = new StraightLineBuilder(parser, source, fn, symbols, enclosing, arena, functionIndex);
    for(const id of parameters) {
        const parameter = parser.node(id);
        if(parameter.kind === 'Identifier') { fn.params.push(builder.bind(id)); } else { const temporary = builder.temporary(id); fn.params.push(temporary); const pattern = builder.pattern(id); builder.emit({kind: 'Destructure', lvaluePattern: pattern, pattern, value: temporary, declarationKind: 1}, id, undefined); }
    }
    builder.findContext(bodyId >= 0 ? bodyId : conciseId);
    builder.statements(statements);
    if(conciseId >= 0) { builder.close({ kind: 'Return', value: builder.expression(conciseId) }); }
    else { builder.close({ kind: 'Return', value: fn.returns }); }
    return builder;
}
// Construct once; the file entry cache owns the parser and checker facts.
export function lowerParsedFunction(parser: Parser, root: number, source: string, symbols: SymbolSnapshot | undefined, cloneBeforeSSA: boolean = false, construct: boolean = true): ConstructedHIR | undefined {
    if(!['FunctionDeclaration', 'FunctionExpression', 'ArrowFunction', 'MethodDeclaration', 'GetAccessor', 'SetAccessor', 'Constructor'].includes(parser.node(root).kind) || !supportedFunction(parser, root)) { return undefined; }
    const arena = new HIRArena();
    const builder = lowerFunction(parser, root, source, symbols, undefined, arena);
    if(builder === undefined) { return undefined; }
    if(construct) { constructHIR(arena, builder.functionIndex, cloneBeforeSSA); }
    return new ConstructedHIR(arena, builder.functionIndex);
}
// Corpus positions are UTF-8 offsets; path selects a nested graph constructed with its parent.
export function lowerSourceAt(source: string, start: number, end: number, symbols: SymbolSnapshot | undefined = undefined, path: string = '', filePath: string = '/test.tsx', cloneBeforeSSA: boolean = false): ConstructedHIR | undefined {
    const parser = new Parser(source, filePath); parser.file();
    let root = -1;
    for(let index = 0; index < parser.nodes.length; index++) {
        const candidate = parser.node(index);
        if(['FunctionDeclaration', 'FunctionExpression', 'ArrowFunction', 'MethodDeclaration', 'GetAccessor', 'SetAccessor', 'Constructor'].includes(candidate.kind) && (start < 0 || (utf8Length(source.slice(0, candidate.pos)) === start && utf8Length(source.slice(0, candidate.end)) === end))) { root = index; break; }
    }
    if(root < 0) { return undefined; }
    const graph = lowerParsedFunction(parser, root, source, symbols, cloneBeforeSSA);
    if(graph === undefined) { return undefined; }
    const arena = graph.arena;
    let index = graph.root;
    if(path !== '') { for(const part of path.split(',')) { index = arena.read(index).functions[Number.parseInt(part, 10)] ?? panic('missing nested path'); } }
    return new ConstructedHIR(arena, index);
}
export function lowerSource(source: string): ConstructedHIR | undefined { return lowerSourceAt(source, -1, -1); }
