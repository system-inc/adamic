// Construction slice of lower.go / lower_expression.go. Other paths are explicit declines.
import { panic, utf8Length } from 'adamic';
import { HIRFunction, Instruction, BasicBlock, HIRArena, ConstructedHIR, blockIndex } from './core.ts';
import { Parser } from '../../typescript/parser/parser.ts';
import { written } from '../../typescript/parser/nodes.ts';
import type { PlaceInterface, ValueType, TerminalType, ArgumentInterface, ModuleExportOriginInterface, JsxTagInterface, JsxAttributeInterface, FunctionIndex, BlockIndex } from './core.ts';
import { SymbolSnapshot } from './symbol.ts';
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
 return node.kind === 'Identifier' || ((node.kind === 'PropertyAccessExpression' || node.kind === 'ElementAccessExpression') && supportedExpression(parser, id));
}
function supportedExpression(parser: Parser, id: number): boolean {
    const node = parser.node(id);
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
        return !node.optional && node.children.length === 2 && supportedExpression(parser, node.children[0] ?? -1) && (node.kind === 'PropertyAccessExpression' || supportedExpression(parser, node.children[1] ?? -1));
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
    return false;
}
function supportedList(parser: Parser, id: number): boolean {
    const list = parser.node(id);
    if(list.kind !== 'VariableDeclarationList') { return false; }
    for(const child of list.children) { const declaration = parser.node(child); if(declaration.children.length > 2 || parser.node(declaration.children[0] ?? -1).kind !== 'Identifier' || (declaration.children[1] !== undefined && !supportedExpression(parser, declaration.children[1] ?? -1))) { return false; } }
    return true;
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
        const init = statement.slots[0] ?? -1;
        return (init < 0 || (parser.node(init).kind === 'VariableDeclarationList' ? supportedList(parser, init) : supportedExpression(parser, init))) && statement.slots.slice(1, 3).every((child) => child < 0 || supportedExpression(parser, child)) && supportedStatement(parser, statement.slots[3] ?? -1);
    }
    if(statement.kind === 'ForOfStatement' || statement.kind === 'ForInStatement') {
        const offset = parser.node(statement.children[0] ?? -1).kind === 'AwaitKeyword' ? 1 : 0;
        const init = statement.children[offset] ?? -1;
        return (parser.node(init).kind === 'VariableDeclarationList' ? supportedList(parser, init) : supportedTarget(parser, init)) && supportedExpression(parser, statement.children[offset + 1] ?? -1) && supportedStatement(parser, statement.children[offset + 2] ?? -1);
    }
    if(statement.kind === 'SwitchStatement') { return supportedExpression(parser, statement.children[0] ?? -1) && parser.node(statement.children[1] ?? -1).children.every((child) => { const clause = parser.node(child); return (clause.kind === 'DefaultClause' || supportedExpression(parser, clause.children[0] ?? -1)) && clause.children.slice(clause.kind === 'CaseClause' ? 1 : 0).every((id) => supportedStatement(parser, id)); }); }
    if(statement.kind === 'LabeledStatement') { return supportedStatement(parser, statement.children[1] ?? -1); }
    if(statement.kind === 'TryStatement') { return statement.children.every((child) => { const part = parser.node(child); if(part.kind !== 'CatchClause') { return supportedStatement(parser, child); } const binding = part.children.length > 1 ? parser.node(part.children[0] ?? -1) : undefined; return (binding === undefined || (binding.children.length === 1 && parser.node(binding.children[0] ?? -1).kind === 'Identifier')) && supportedStatement(parser, part.children[part.children.length - 1] ?? -1); }); }
    if(statement.kind === 'WhileStatement') { return supportedExpression(parser, statement.children[0] ?? -1) && supportedStatement(parser, statement.children[1] ?? -1); }
    if(statement.kind === 'BreakStatement' || statement.kind === 'ContinueStatement') { return statement.children.length <= 1; }
        if(statement.kind === 'VariableStatement') {
            const list = parser.node(statement.children[0] ?? -1);
            if(list.kind !== 'VariableDeclarationList') { return false; }
            for(const declarationId of list.children) {
                const declaration = parser.node(declarationId);
                if(declaration.children.length > 2 || parser.node(declaration.children[0] ?? -1).kind !== 'Identifier') { return false; }
                if(declaration.children[1] !== undefined && !supportedExpression(parser, declaration.children[1] ?? -1)) { return false; }
            }
            return true;
        }
    if(!['ExpressionStatement', 'ReturnStatement', 'ThrowStatement', 'EmptyStatement'].includes(statement.kind) || statement.children.length > 1) { return false; }
    return statement.children.every((child) => supportedExpression(parser, child));
}
function supportedFunction(parser: Parser, root: number): boolean {
    const node = parser.node(root);
    let body = false;
    for(const id of node.children) {
        const child = parser.node(id);
        if(child.kind === 'Identifier' && node.kind !== 'ArrowFunction') { continue; }
        if(child.kind === 'EqualsGreaterThanToken' && node.kind === 'ArrowFunction') { continue; }
        if(child.kind === 'Parameter' && child.children.length === 1 && parser.node(child.children[0] ?? -1).kind === 'Identifier') { continue; }
        if(child.kind === 'Block') { if(!supportedStatement(parser, id)) { return false; } body = true; }
        else if(node.kind === 'ArrowFunction' && supportedExpression(parser, id)) { body = true; }
        else { return false; }
    }
    return body;
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
            const next = ['FunctionDeclaration', 'FunctionExpression', 'ArrowFunction'].includes(node.kind) ? depth + 1 : depth;
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
            captures.push(place === undefined ? { identifier: 0, effect: '<unknown>', reactive: false, start: 0, end: 0 } : { identifier: place.identifier, effect: '<unknown>', reactive: false, start: 0, end: 0 });
        }
        const functionId = this.fn.functions.length;
        this.fn.functions.push(builder.functionIndex);
        return this.emit({ kind: 'FunctionExpression', functionReference: { index: builder.functionIndex, ordinal: functionId }, captures }, id, undefined);
    }
    bind(id: number): PlaceInterface {
        const node = this.parser.node(id);
        const symbol = this.identity(id);
        const old = this.locals.get(symbol);
        const declaration = old === undefined || symbol === 0 ? 0 : (this.fn.identifiers[old.identifier] ?? panic('missing local')).declaration;
        const place = this.fn.named(node.text, this.byte(node.pos), this.byte(node.end), declaration);
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
                if(this.contextual.has(symbol)) { this.fn.contextDeclarations.add((this.fn.identifiers[place.identifier] ?? panic('missing context identifier')).declaration); }
                this.emit({ kind: this.contextual.has(symbol) ? 'StoreContext' : 'StoreLocal', lvalue: place, value, declarationKind: 2 }, id, undefined);
            }
            else {
                const capture = this.captureOf(symbol);
                if(capture === undefined) { this.emit({ kind: 'StoreGlobal', name: node.text, value }, id, undefined); }
                else { this.emit({ kind: 'StoreContext', lvalue: { identifier: capture.identifier, effect: '<unknown>', reactive: false, start: this.byte(node.pos), end: this.byte(node.end) }, value, declarationKind: 2 }, id, undefined); }
            }
        }
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
        const declarationKind = list.semantic === '2' ? 0 : 1;
        for(const declarationId of list.children) {
            const declaration = this.parser.node(declarationId);
            const name = declaration.children[0] ?? -1;
            const initializer = declaration.children[1];
            if(initializer === undefined) { this.emit({ kind: 'DeclareLocal', lvalue: this.bind(name), declarationKind }, declarationId, undefined); }
            else {
                const value = this.expression(initializer);
                this.emit({ kind: 'StoreLocal', lvalue: this.bind(name), value, declarationKind }, declarationId, undefined);
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
        for(const child of node.children) { const name = this.parser.node(child).children[0] ?? -1; this.emit({ kind: 'StoreLocal', lvalue: this.bind(name), value, declarationKind: node.semantic === '2' ? 0 : 1 }, name, undefined); }
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
            const init = this.fn.newBlock('block'); const test = this.fn.newBlock('block'); const loop = this.fn.newBlock('loop'); const fallthrough = this.fn.newBlock('block');
            const increment = node.slots[2] ?? -1; const update = increment < 0 ? undefined : this.fn.newBlock('block');
            if(update === undefined) { this.close({ kind: 'For', init: init.id, testBlock: test.id, loop: loop.id, fallthrough: fallthrough.id }); } else { this.close({ kind: 'For', init: init.id, testBlock: test.id, loop: loop.id, update: update.id, fallthrough: fallthrough.id }); } this.current = init;
            const initializer = node.slots[0] ?? -1; if(initializer >= 0) { if(this.parser.node(initializer).kind === 'VariableDeclarationList') { this.declarationList(initializer); } else { this.expression(initializer); } }
            this.jump(test.id, 0); this.current = test; const condition = node.slots[1] ?? -1;
            if(condition < 0) { this.jump(loop.id, 0); } else { const value = this.expression(condition); this.close({ kind: 'Branch', testPlace: value, consequent: loop.id, alternate: fallthrough.id, fallthrough: fallthrough.id }); }
            const continueBlock = update?.id ?? test.id; this.current = loop; this.jumps.push({ label, breakBlock: fallthrough.id, continueBlock }); this.statement(node.slots[3] ?? -1); this.jumps.pop();
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
        if(catchId !== undefined) { const clause = this.parser.node(catchId); if(clause.children.length > 1) { binding = this.bind(this.parser.node(clause.children[0] ?? -1).children[0] ?? -1); } }
        this.close({ kind: 'Try', block: body.id, handler: handler.id, handlerBinding: binding, fallthrough: fallthrough.id }); this.current = body; this.statement(node.children[0] ?? -1); this.jump(normal.id, 2);
        this.current = handler;
        if(catchId !== undefined) { const clause = this.parser.node(catchId); this.statement(clause.children[clause.children.length - 1] ?? -1); this.jump(normal.id, 2); }
        else if(finalBlock !== undefined) { this.jump(finalBlock.id, 2); } else { this.close({ kind: 'Unreachable' }); }
        if(finalBlock !== undefined) { this.current = finalBlock; this.statement(finallyId ?? -1); this.jump(fallthrough.id, 2); }
        this.current = fallthrough;
    }
    valueBranch(id: number, logical: string | undefined): PlaceInterface {
        const node = this.parser.node(id);
        const result = this.fn.temporary(this.byte(node.pos), this.byte(node.end));
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
        const shared = (this.fn.identifiers[result.identifier] ?? panic('missing result')).declaration;
        this.current = first;
        const firstNode = node.children[logical === undefined ? 2 : 0] ?? -1;
        const firstValue = logical === undefined ? this.expression(firstNode) : left;
        const firstPlace = logical === undefined ? result : this.fn.temporary(this.byte(this.parser.node(firstNode).pos), this.byte(this.parser.node(firstNode).end), shared);
        this.emit({ kind: 'LoadLocal', place: firstValue }, firstNode, firstPlace); this.jump(fallthrough.id, 0);
        this.current = second;
        const secondNode = node.children[logical === undefined ? 4 : 2] ?? -1;
        const secondValue = this.expression(secondNode);
        const secondPlace = logical === undefined ? result : this.fn.temporary(this.byte(this.parser.node(secondNode).pos), this.byte(this.parser.node(secondNode).end), shared);
        this.emit({ kind: 'LoadLocal', place: secondValue }, secondNode, secondPlace); this.jump(fallthrough.id, 0);
        this.current = fallthrough;
        return result;
    }
    byte(index: number): number { return utf8Length(this.source.slice(0, index)); }
    emit(value: ValueType, id: number, target: PlaceInterface | undefined): PlaceInterface {
        const node = this.parser.node(id);
        const start = this.byte(node.pos);
        const end = this.byte(node.end);
        const place = target ?? this.fn.temporary(start, end);
        const block = this.ensureBlock();
        block.instructions.push(this.fn.instructions.length);
        this.fn.instructions.push(new Instruction(this.fn.instructions.length, place, value, start, end));
        return place;
    }
    // Resolve direct imports and same-file const aliases from compiler symbol identities.
    // Barrel export graphs remain a separate construction seam.
    exportOrigin(id: number, active: Set<number>): ModuleExportOriginInterface {
        const empty: ModuleExportOriginInterface = { module: '', exported: '' };
        if(this.symbols === undefined) { return empty; }
        const node = this.parser.node(id);
        if(node.kind === 'ParenthesizedExpression') { return this.exportOrigin(node.children[0] ?? -1, active); }
        if(node.kind === 'PropertyAccessExpression' || node.kind === 'ElementAccessExpression') {
            const receiver = this.exportOrigin(node.children[0] ?? -1, active);
            const property = this.parser.node(node.children[1] ?? -1);
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
                    if(parent?.semantic === '2' && initializer !== undefined) { result = this.exportOrigin(initializer, active); if(result.module !== '') { break; } }
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
            const propertyId = callee.children[1] ?? -1;
            const property = callee.kind === 'PropertyAccessExpression' ? this.emit({ kind: 'Primitive', literal: `string:${written(this.parser.node(propertyId).text)}` }, propertyId, undefined) : this.expression(propertyId);
            return this.emit({ kind: 'MethodCall', receiver, property, args: this.arguments(id), optional, origin }, id, undefined);
        }
        const place = this.expression(calleeId);
        return this.emit({ kind: 'CallExpression', callee: place, args: this.arguments(id), optional, origin }, id, undefined);
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
    expression(id: number): PlaceInterface {
        const node = this.parser.node(id);
        if(node.kind === 'JsxElement' || node.kind === 'JsxSelfClosingElement' || node.kind === 'JsxFragment') { return this.jsx(id); }
        if(node.kind === 'JsxExpression') { const inner = node.children.find((child) => this.parser.node(child).kind !== 'DotDotDotToken'); return inner === undefined ? this.emit({ kind: 'Primitive', literal: 'nil' }, id, undefined) : this.expression(inner); }
        if(node.kind === 'FunctionExpression' || node.kind === 'ArrowFunction') { return this.nested(id); }
        const primitive = literal(node.kind, node.text);
        if(primitive !== undefined) { return this.emit({ kind: 'Primitive', literal: primitive }, id, undefined); }
        if(node.kind === 'CallExpression' || node.kind === 'NewExpression') { return this.call(id); }
        if(node.kind === 'ConditionalExpression') { return this.valueBranch(id, undefined); }
        if(node.kind === 'Identifier') { return this.loadIdentifier(id); }
        if(node.kind === 'PropertyAccessExpression' || node.kind === 'ElementAccessExpression') {
            const object = this.expression(node.children[0] ?? -1);
            if(node.kind === 'PropertyAccessExpression') { return this.emit({ kind: 'PropertyLoad', object, property: this.parser.node(node.children[1] ?? -1).text }, id, undefined); }
            return this.emit({ kind: 'ComputedLoad', object, property: this.expression(node.children[1] ?? -1) }, id, undefined);
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
    for(const id of node.children) {
        const child = parser.node(id);
        if(child.kind === 'Identifier' && node.kind !== 'ArrowFunction') { name = child.text; }
        else if(child.kind === 'Block') { bodyId = id; }
        else if(node.kind === 'ArrowFunction' && child.kind === 'EqualsGreaterThanToken') { continue; }
        else if(node.kind === 'ArrowFunction' && supportedExpression(parser, id)) { conciseId = id; }
        else if(child.kind === 'Parameter' && child.children.length === 1 && parser.node(child.children[0] ?? -1).kind === 'Identifier') { parameters.push(child.children[0] ?? -1); }
        else { return undefined; }
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
    const builder = new StraightLineBuilder(parser, source, fn, symbols, enclosing, arena, functionIndex);
    for(const id of parameters) {
        const parameter = parser.node(id);
        fn.params.push(builder.bind(id));
    }
    builder.findContext(bodyId >= 0 ? bodyId : conciseId);
    builder.statements(statements);
    if(conciseId >= 0) { builder.close({ kind: 'Return', value: builder.expression(conciseId) }); }
    else { builder.close({ kind: 'Return', value: fn.returns }); }
    return builder;
}
// Construct once; the file entry cache owns the parser and checker facts.
export function lowerParsedFunction(parser: Parser, root: number, source: string, symbols: SymbolSnapshot | undefined): ConstructedHIR | undefined {
    if(!['FunctionDeclaration', 'FunctionExpression', 'ArrowFunction'].includes(parser.node(root).kind) || !supportedFunction(parser, root)) { return undefined; }
    const arena = new HIRArena();
    const builder = lowerFunction(parser, root, source, symbols, undefined, arena);
    if(builder === undefined) { return undefined; }
    constructHIR(arena, builder.functionIndex);
    return new ConstructedHIR(arena, builder.functionIndex);
}
// Corpus positions are UTF-8 offsets; path selects a nested graph constructed with its parent.
export function lowerSourceAt(source: string, start: number, end: number, symbols: SymbolSnapshot | undefined = undefined, path: string = ''): ConstructedHIR | undefined {
    const parser = new Parser(source, '/test.tsx'); parser.file();
    let root = -1;
    for(let index = 0; index < parser.nodes.length; index++) {
        const candidate = parser.node(index);
        if(['FunctionDeclaration', 'FunctionExpression', 'ArrowFunction'].includes(candidate.kind) && (start < 0 || (utf8Length(source.slice(0, candidate.pos)) === start && utf8Length(source.slice(0, candidate.end)) === end))) { root = index; break; }
    }
    if(root < 0) { return undefined; }
    const graph = lowerParsedFunction(parser, root, source, symbols);
    if(graph === undefined) { return undefined; }
    const arena = graph.arena;
    let index = graph.root;
    if(path !== '') { for(const part of path.split(',')) { index = arena.read(index).functions[Number.parseInt(part, 10)] ?? panic('missing nested path'); } }
    return new ConstructedHIR(arena, index);
}
export function lowerSource(source: string): ConstructedHIR | undefined { return lowerSourceAt(source, -1, -1); }
