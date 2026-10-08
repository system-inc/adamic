// First part of lower.go and lower_expression.go. Parse the source with the stage 1 parser.
import { panic } from 'adamic';
import { Parser } from '../../typescript/parser/parser.ts';
import { HIRFunction, Instruction } from './core.ts';
import type { PlaceInterface, ValueType } from './core.ts';
import { constructHIR } from './graph.ts';
function literal(kind: string, text: string): string | undefined {
    if(kind === 'NumericLiteral') { return `string:${text}`; }
    if(kind === 'TrueKeyword') { return 'bool:true'; }
    if(kind === 'FalseKeyword') { return 'bool:false'; }
    if(kind === 'NullKeyword') { return 'nil'; }
    return undefined;
}
export function lowerSource(source: string): HIRFunction | undefined {
    const parser = new Parser(source, '/test.tsx');
    parser.file();
    const node = parser.nodes.find((candidate) => candidate.kind === 'FunctionDeclaration');
    if(node === undefined) { return undefined; }
    let name = '';
    let bodyId = -1;
    for(const id of node.children) {
        const child = parser.node(id);
        if(child.kind === 'Identifier') { name = child.text; }
        else if(child.kind === 'Block') { bodyId = id; }
        else { return undefined; }
    }
    if(bodyId < 0) { return undefined; }
    const statements = parser.node(bodyId).children;
    for(const id of statements) {
        const statement = parser.node(id);
        if(statement.kind !== 'ExpressionStatement' && statement.kind !== 'ReturnStatement' && statement.kind !== 'EmptyStatement') { return undefined; }
        if(statement.children.length > 1) { return undefined; }
        for(const expression of statement.children) {
            const value = parser.node(expression);
            if(literal(value.kind, value.text) === undefined) { return undefined; }
        }
    }
    const fn = new HIRFunction(name);
    const block = fn.blocks[0] ?? panic('missing entry');
    for(const id of statements) {
        const statement = parser.node(id);
        if(statement.kind === 'EmptyStatement') { continue; }
        let result: PlaceInterface | undefined;
        const expressionId = statement.children[0];
        if(expressionId !== undefined) {
            const expression = parser.node(expressionId);
            result = fn.temporary(expression.pos, expression.end);
            const value: ValueType = { kind: 'Primitive', literal: literal(expression.kind, expression.text) ?? panic('unsupported literal') };
            block.instructions.push(fn.instructions.length);
            fn.instructions.push(new Instruction(fn.instructions.length, result, value, expression.pos, expression.end));
        }
        if(statement.kind === 'ReturnStatement') {
            const value: ValueType = result === undefined ? { kind: 'Primitive', literal: 'nil' } : { kind: 'LoadLocal', place: result };
            block.instructions.push(fn.instructions.length);
            fn.instructions.push(new Instruction(fn.instructions.length, fn.returns, value, statement.pos, statement.end));
            break;
        }
    }
    constructHIR(fn);
    return fn;
}
