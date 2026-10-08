// Go cache.go: one file/checker owns the cache; returned graph owners retain their arenas.
import { utf8Length } from 'adamic';
import { lowerParsedFunction } from './lower.ts';
import { Parser } from '../../typescript/parser/parser.ts';
import type { SymbolSnapshot } from './symbol.ts';
import type { ConstructedHIR } from './core.ts';
export class HIRFile {
    readonly source: string;
    readonly cloneBeforeSSA: boolean;
    readonly parser: Parser;
    readonly symbols: SymbolSnapshot | undefined;
    private readonly functions: Map<string, ConstructedHIR | undefined> = new Map<string, ConstructedHIR | undefined>();
    constructor(source: string, symbols: SymbolSnapshot | undefined, path: string = '/test.tsx', cloneBeforeSSA: boolean = false, parser: Parser | undefined = undefined) {
        this.cloneBeforeSSA = cloneBeforeSSA; this.source = source; this.symbols = symbols; this.parser = parser ?? new Parser(source, path);
        if(parser === undefined) { this.parser.file(); }
    }
    forFunction(index: number): ConstructedHIR | undefined {
        if(this.symbols === undefined || index < 0) { return undefined; }
        const node = this.parser.node(index);
        const key = `${node.kind}:${node.pos}`;
        if(this.functions.has(key)) { return this.functions.get(key); }
        const graph = lowerParsedFunction(this.parser, index, this.source, this.symbols, this.cloneBeforeSSA);
        this.functions.set(key, graph);
        return graph;
    }
    at(start: number, end: number): number {
        for(let index = 0; index < this.parser.nodes.length; index++) {
            const node = this.parser.node(index);
            if(['FunctionDeclaration', 'FunctionExpression', 'ArrowFunction', 'MethodDeclaration', 'GetAccessor', 'SetAccessor', 'Constructor'].includes(node.kind) && utf8Length(this.source.slice(0, node.pos)) === start && utf8Length(this.source.slice(0, node.end)) === end) { return index; }
        }
        return -1;
    }
}
export function ForFunction(file: HIRFile, index: number): ConstructedHIR | undefined { return file.forFunction(index); }

export function AsCompilationUnit(file: HIRFile, graph: ConstructedHIR): ConstructedHIR {
    return ForFunction(file, graph.arena.read(graph.root).nodeIndex) ?? graph;
}
export function MayHoldComponentOrHook(parser: Parser, source: string): boolean {
    if((parser.path.endsWith('.tsx') || parser.path.endsWith('.jsx')) && source.includes('<')) { return true; }
    for(let start = source.indexOf('use'); start >= 0; start = source.indexOf('use',start + 3)) { const next = source.charCodeAt(start + 3); if((next >= 65 && next <= 90) || (next >= 48 && next <= 57)) { return true; } }
    if(!source.includes('\\')) { return false; }
    return parser.nodes.some((node) => { const next = node.text.charCodeAt(3); return node.kind === 'Identifier' && node.text.startsWith('use') && ((next >= 65 && next <= 90) || (next >= 48 && next <= 57)); });
}
