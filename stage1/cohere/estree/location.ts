// Byte ranges and Prettier's decorator, statement and semicolon overrides.
import type { Arena } from './arena.ts';
export function locStart(arena: Arena, id: number): number {
    const node = arena.node(id);
    const declaration = node.child('declaration');
    const decorators =
        declaration >= 0 && arena.node(declaration).has('decorators')
            ? arena.node(declaration).list('decorators')
            : node.list('decorators');
    const first = decorators[0] ?? -1;
    return first >= 0 ? Math.min(locStart(arena, first), node.start) : node.start;
}
export function shouldAddContentEnd(type: string): boolean {
    return [
        'ExpressionStatement',
        'Directive',
        'ImportDeclaration',
        'ExportDefaultDeclaration',
        'ExportNamedDeclaration',
        'ExportAllDeclaration',
        'ReturnStatement',
        'ThrowStatement',
        'DoWhileStatement',
    ].includes(type);
}
export function bodyOverride(type: string): boolean {
    return [
        'ForInStatement',
        'ForOfStatement',
        'ForStatement',
        'LabeledStatement',
        'WithStatement',
        'WhileStatement',
    ].includes(type);
}
export function locEnd(arena: Arena, id: number): number {
    const node = arena.node(id);
    if(node.type === 'IfStatement') {
        const alternate = node.child('alternate');
        return locEnd(arena, alternate >= 0 ? alternate : node.child('consequent'));
    }
    if(bodyOverride(node.type)) {
        return locEnd(arena, node.child('body'));
    }
    if(node.type === 'BreakStatement' || node.type === 'ContinueStatement') {
        const label = node.child('label');
        return label >= 0 ? locEnd(arena, label) : locStart(arena, id) + (node.type === 'BreakStatement' ? 5 : 8);
    }
    if(node.type === 'DebuggerStatement') {
        return locStart(arena, id) + 8;
    }
    if(node.type === 'VariableDeclaration') {
        const declarations = node.list('declarations');
        return locEnd(arena, declarations[declarations.length - 1] ?? -1);
    }
    return shouldAddContentEnd(node.type) && node.hasContentEnd ? node.contentEnd : node.end;
}
export function ignoredSemicolon(arena: Arena, id: number): boolean {
    const node = arena.node(id);
    if(shouldAddContentEnd(node.type) && node.hasContentEnd && node.contentEnd !== 0) {
        return true;
    }
    if(['BreakStatement', 'ContinueStatement', 'DebuggerStatement', 'VariableDeclaration'].includes(node.type)) {
        return true;
    }
    if(node.type === 'IfStatement') {
        const alternate = node.child('alternate');
        return ignoredSemicolon(arena, alternate >= 0 ? alternate : node.child('consequent'));
    }
    return bodyOverride(node.type) ? ignoredSemicolon(arena, node.child('body')) : false;
}
export function sameLocStart(arena: Arena, left: number, right: number): boolean {
    return locStart(arena, left) === locStart(arena, right);
}
export function sameLoc(arena: Arena, left: number, right: number): boolean {
    return sameLocStart(arena, left, right) && locEnd(arena, left) === locEnd(arena, right);
}
