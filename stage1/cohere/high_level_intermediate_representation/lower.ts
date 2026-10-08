// Straight-line part of lower.go / lower_expression.go. Other paths are explicit declines.
import { panic, utf8Length } from 'adamic';
import { Parser } from '../../typescript/parser/parser.ts';
import { written } from '../../typescript/parser/nodes.ts';
import { HIRFunction, Instruction } from './core.ts';
import type { PlaceInterface, ValueType } from './core.ts';
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
function supportedExpression(parser: Parser, id: number): boolean {
    const node = parser.node(id);
    if(literal(node.kind, node.text) !== undefined) { return true; }
    if(['ParenthesizedExpression', 'TypeOfExpression', 'VoidExpression'].includes(node.kind)) {
        return node.children.length === 1 && supportedExpression(parser, node.children[0] ?? -1);
    }
    if(node.kind === 'PrefixUnaryExpression') {
        return ['PlusToken', 'MinusToken', 'ExclamationToken', 'TildeToken'].includes(node.operator) && supportedExpression(parser, node.children[0] ?? -1);
    }
    if(node.kind === 'BinaryExpression') {
        const operator = parser.node(node.children[1] ?? -1).kind;
        return (operator === 'CommaToken' || operators.has(operator)) && supportedExpression(parser, node.children[0] ?? -1) && supportedExpression(parser, node.children[2] ?? -1);
    }
    return false;
}
class StraightLineBuilder {
    readonly parser: Parser;
    readonly source: string;
    readonly fn: HIRFunction;
    constructor(parser: Parser, source: string, fn: HIRFunction) { this.parser = parser; this.source = source; this.fn = fn; }
    byte(index: number): number { return utf8Length(this.source.slice(0, index)); }
    emit(value: ValueType, id: number, target: PlaceInterface | undefined): PlaceInterface {
        const node = this.parser.node(id);
        const start = this.byte(node.pos);
        const end = this.byte(node.end);
        const place = target ?? this.fn.temporary(start, end);
        const block = this.fn.blocks[0] ?? panic('missing entry');
        block.instructions.push(this.fn.instructions.length);
        this.fn.instructions.push(new Instruction(this.fn.instructions.length, place, value, start, end));
        return place;
    }
    expression(id: number): PlaceInterface {
        const node = this.parser.node(id);
        const primitive = literal(node.kind, node.text);
        if(primitive !== undefined) { return this.emit({ kind: 'Primitive', literal: primitive }, id, undefined); }
        if(node.kind === 'ParenthesizedExpression') { return this.expression(node.children[0] ?? -1); }
        if(node.kind === 'BinaryExpression') {
            const left = this.expression(node.children[0] ?? -1);
            const operator = this.parser.node(node.children[1] ?? -1).kind;
            const right = this.expression(node.children[2] ?? -1);
            if(operator === 'CommaToken') { return right; }
            return this.emit({ kind: 'BinaryExpression', left, operator: operators.get(operator) ?? panic('unsupported binary'), right }, id, undefined);
        }
        const value = this.expression(node.children[0] ?? -1);
        const operator = node.kind === 'TypeOfExpression' ? 'typeof' : node.kind === 'VoidExpression' ? 'void' : operators.get(node.operator) ?? panic('unsupported unary');
        return this.emit({ kind: 'UnaryExpression', operator, value }, id, undefined);
    }
}
// Positions supplied by the corpus are byte offsets, as in Go. -1 chooses the first declaration.
export function lowerSourceAt(source: string, start: number, end: number): HIRFunction | undefined {
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
        if(child.kind === 'Identifier') { name = child.text; }
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
    for(const id of statements) {
        const statement = parser.node(id);
        if(!['ExpressionStatement', 'ReturnStatement', 'EmptyStatement'].includes(statement.kind)) { return undefined; }
        if(statement.children.length > 1) { return undefined; }
        for(const expression of statement.children) { if(!supportedExpression(parser, expression)) { return undefined; } }
    }
    const fn = new HIRFunction(name);
    const builder = new StraightLineBuilder(parser, source, fn);
    for(const id of parameters) {
        const parameter = parser.node(id);
        fn.params.push(fn.named(parameter.text, builder.byte(parameter.pos), builder.byte(parameter.end)));
    }
    const block = fn.blocks[0] ?? panic('missing entry');
    for(const id of statements) {
        const statement = parser.node(id);
        if(statement.kind === 'EmptyStatement') { continue; }
        let result: PlaceInterface | undefined;
        const expressionId = statement.children[0];
        if(expressionId !== undefined) { result = builder.expression(expressionId); }
        if(statement.kind === 'ReturnStatement') {
            const value: ValueType = result === undefined ? { kind: 'Primitive', literal: 'nil' } : { kind: 'LoadLocal', place: result };
            builder.emit(value, id, fn.returns);
            break;
        }
    }
    if(conciseId >= 0) { block.terminal = builder.expression(conciseId); }
    constructHIR(fn);
    return fn;
}
export function lowerSource(source: string): HIRFunction | undefined { return lowerSourceAt(source, -1, -1); }
