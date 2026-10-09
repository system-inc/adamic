import { panic } from 'adamic';
import { Scanner } from '../../typescript/scanner/scanner.ts';
import type { ParseNode } from '../../typescript/parser/nodes.ts';
import type { Arena } from './arena.ts';
import { listValue } from './values.ts';

export function typeList(
    arena: Arena,
    text: string,
    offsets: readonly number[],
    nodes: readonly ParseNode[],
    ids: readonly number[],
    parameters: boolean,
    convert: (id: number) => number,
): number {
    if(ids.length === 0) {
        return -1;
    }
    const scanner = new Scanner(text);
    const first = nodes[ids[0] ?? -1] ?? panic('missing first type');
    const type = parameters ? 'TSTypeParameterDeclaration' : 'TSTypeParameterInstantiation';
    if(ids.length === 1 && ['EmptyTypeArguments', 'EmptyTypeParameters'].includes(first.kind)) {
        scanner.pos = first.pos;
        scanner.scan();
        const result = arena.newNode(
            type,
            offsets[scanner.start] ?? panic('AST range outside source'),
            offsets[first.end] ?? panic('AST range outside source'),
        );
        arena.node(result).set('params', listValue([]));
        return result;
    }
    const last = nodes[ids[ids.length - 1] ?? -1] ?? panic('missing last type');
    const start = first.pos - 1;
    if(text.slice(start, start + 1) !== '<') {
        return panic('ESTree type list start is not represented');
    }
    scanner.pos = last.end;
    scanner.scan();
    if(scanner.kind === 'CommaToken') {
        scanner.scan();
    }
    if(scanner.kind !== 'GreaterThanToken') {
        return panic('ESTree type list end is not represented');
    }
    const result = arena.newNode(
        type,
        offsets[start] ?? panic('AST range outside source'),
        offsets[scanner.pos] ?? panic('AST range outside source'),
    );
    const params: number[] = [];
    for(const id of ids) {
        params.push(convert(id));
    }
    arena.node(result).set('params', listValue(params));
    return result;
}
