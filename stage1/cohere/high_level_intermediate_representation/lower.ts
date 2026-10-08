// Construction slice of lower.go / lower_expression.go. Other paths are explicit declines.
import { panic, utf8Length } from 'adamic';
import { Parser } from '../../typescript/parser/parser.ts';
import { written } from '../../typescript/parser/nodes.ts';
import { HIRFunction, Instruction, BasicBlock } from './core.ts';
import type { PlaceInterface, ValueType, TerminalType } from './core.ts';
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
    if(node.kind === 'Identifier' || literal(node.kind, node.text) !== undefined) { return true; }
    if(['ParenthesizedExpression', 'TypeOfExpression', 'VoidExpression'].includes(node.kind)) {
        return node.children.length === 1 && supportedExpression(parser, node.children[0] ?? -1);
    }
    if(node.kind === 'PropertyAccessExpression' || node.kind === 'ElementAccessExpression') {
        return !node.optional && node.children.length === 2 && supportedExpression(parser, node.children[0] ?? -1) && (node.kind === 'PropertyAccessExpression' || supportedExpression(parser, node.children[1] ?? -1));
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
function supportedStatement(parser: Parser, id: number): boolean {
    const statement = parser.node(id);
    if(statement.kind === 'Block') { return statement.children.every((child) => supportedStatement(parser, child)); }
    if(statement.kind === 'IfStatement') {
        return supportedExpression(parser, statement.children[0] ?? -1) && supportedStatement(parser, statement.children[1] ?? -1) && (statement.children[2] === undefined || supportedStatement(parser, statement.children[2] ?? -1));
    }
    if(statement.kind === 'WhileStatement') { return supportedExpression(parser, statement.children[0] ?? -1) && supportedStatement(parser, statement.children[1] ?? -1); }
    if(statement.kind === 'BreakStatement' || statement.kind === 'ContinueStatement') { return statement.children.length === 0; }
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
class StraightLineBuilder {
    readonly parser: Parser;
    readonly source: string;
    readonly fn: HIRFunction;
    current: BasicBlock | undefined;
    readonly jumps: { readonly breakBlock: number; readonly continueBlock: number }[] = [];
    readonly symbols: SymbolSnapshot | undefined;
    readonly locals: Map<number, PlaceInterface> = new Map<number, PlaceInterface>();
    constructor(parser: Parser, source: string, fn: HIRFunction, symbols: SymbolSnapshot | undefined) { this.parser = parser; this.source = source; this.fn = fn; this.symbols = symbols; this.current = fn.blocks[0]; }
    identity(id: number): number {
        const node = this.parser.node(id);
        if(this.symbols === undefined) { return 0; }
        return this.symbols.read(this.byte(node.pos), this.byte(node.end)).identity;
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
            if(this.locals.has(this.identity(id))) { this.emit({ kind: 'StoreLocal', lvalue: this.bind(id), value, declarationKind: 2 }, id, undefined); }
            else { this.emit({ kind: 'StoreGlobal', name: node.text, value }, id, undefined); }
        }
        else {
            const object = this.expression(node.children[0] ?? -1);
            if(node.kind === 'PropertyAccessExpression') { this.emit({ kind: 'PropertyStore', object, property: this.parser.node(node.children[1] ?? -1).text, value }, id, undefined); }
            else { this.emit({ kind: 'ComputedStore', object, property: this.expression(node.children[1] ?? -1), value }, id, undefined); }
        }
        return value;
    }
    declarations(id: number): void {
        const list = this.parser.node(this.parser.node(id).children[0] ?? -1);
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
    jump(block: number, variant: number): void { this.close({ kind: 'Goto', block, variant }); }
    statements(ids: readonly number[]): void { for(const id of ids) { this.statement(id); } }
    statement(id: number): void {
        const node = this.parser.node(id);
        if(node.kind === 'Block') { this.statements(node.children); }
        else if(node.kind === 'VariableStatement') { this.declarations(id); }
        else if(node.kind === 'IfStatement') {
            const test = this.expression(node.children[0] ?? -1);
            const consequent = this.fn.newBlock('block'); const fallthrough = this.fn.newBlock('block');
            const alternate = node.children[2] === undefined ? fallthrough : this.fn.newBlock('block');
            this.close({ kind: 'If', test, consequent: consequent.id, alternate: alternate.id, fallthrough: fallthrough.id });
            this.current = consequent; this.statement(node.children[1] ?? -1); this.jump(fallthrough.id, 0);
            if(node.children[2] !== undefined) { this.current = alternate; this.statement(node.children[2] ?? -1); this.jump(fallthrough.id, 0); }
            this.current = fallthrough;
        }
        else if(node.kind === 'WhileStatement') {
            const test = this.fn.newBlock('block'); const loop = this.fn.newBlock('loop'); const fallthrough = this.fn.newBlock('block');
            this.close({ kind: 'While', test: test.id, loop: loop.id, fallthrough: fallthrough.id });
            this.current = test;
            const value = this.expression(node.children[0] ?? -1);
            this.close({ kind: 'Branch', test: value, consequent: loop.id, alternate: fallthrough.id, fallthrough: fallthrough.id });
            this.current = loop; this.jumps.push({ breakBlock: fallthrough.id, continueBlock: test.id });
            this.statement(node.children[1] ?? -1); this.jumps.pop(); this.jump(test.id, 1);
            this.current = fallthrough;
        }
        else if(node.kind === 'BreakStatement' || node.kind === 'ContinueStatement') {
            const target = this.jumps[this.jumps.length - 1];
            if(target === undefined) { this.close({ kind: 'Unsupported' }); }
            else { this.jump(node.kind === 'BreakStatement' ? target.breakBlock : target.continueBlock, node.kind === 'BreakStatement' ? 0 : 1); }
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
    valueBranch(id: number, logical: string | undefined): PlaceInterface {
        const node = this.parser.node(id);
        const result = this.fn.temporary(this.byte(node.pos), this.byte(node.end));
        const testBlock = this.fn.newBlock('value'); const fallthrough = this.fn.newBlock('block');
        if(logical === undefined) { this.close({ kind: 'Ternary', test: testBlock.id, fallthrough: fallthrough.id }); }
        else { this.close({ kind: 'Logical', operator: logical, test: testBlock.id, fallthrough: fallthrough.id }); }
        this.current = testBlock;
        const left = this.expression(node.children[0] ?? -1);
        const first = this.fn.newBlock('value'); const second = this.fn.newBlock('value');
        const swapped = logical === '||' || logical === '??';
        // A logical's first arm is the short-circuit arm; a ternary's is the true arm.
        const consequent = logical === undefined || swapped ? first : second;
        const alternate = logical === undefined || swapped ? second : first;
        this.close({ kind: 'Branch', test: left, consequent: consequent.id, alternate: alternate.id, fallthrough: fallthrough.id });
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
    expression(id: number): PlaceInterface {
        const node = this.parser.node(id);
        const primitive = literal(node.kind, node.text);
        if(primitive !== undefined) { return this.emit({ kind: 'Primitive', literal: primitive }, id, undefined); }
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
// Positions supplied by the corpus are byte offsets, as in Go. -1 chooses the first declaration.
export function lowerSourceAt(source: string, start: number, end: number, symbols: SymbolSnapshot | undefined = undefined): HIRFunction | undefined {
    const parser = new Parser(source, '/test.tsx');
    parser.file();
    let root = -1;
    for(let index = 0; index < parser.nodes.length; index++) {
        const candidate = parser.node(index);
        if(['FunctionDeclaration', 'FunctionExpression', 'ArrowFunction'].includes(candidate.kind) && (start < 0 || (utf8Length(source.slice(0, candidate.pos)) === start && utf8Length(source.slice(0, candidate.end)) === end))) { root = index; break; }
    }
    if(root < 0) { return undefined; }
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
    const fn = new HIRFunction(name);
    const builder = new StraightLineBuilder(parser, source, fn, symbols);
    for(const id of parameters) {
        const parameter = parser.node(id);
        fn.params.push(builder.bind(id));
    }
    builder.statements(statements);
    if(conciseId >= 0) { builder.close({ kind: 'Return', value: builder.expression(conciseId) }); }
    else { builder.close({ kind: 'Return', value: fn.returns }); }
    constructHIR(fn);
    return fn;
}
export function lowerSource(source: string): HIRFunction | undefined { return lowerSourceAt(source, -1, -1); }
