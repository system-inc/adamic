// Go cache.go: one file/checker owns the cache; returned graph owners retain their arenas.
import { utf8Length } from 'adamic';
import { lowerParsedFunction } from './lower.ts';
import { Parser } from '../../typescript/parser/parser.ts';
import type { SymbolSnapshot } from './symbol.ts';
import type { ConstructedHIR } from './core.ts';
export class HIRFile {
    readonly source: string;
    readonly parser: Parser;
    readonly symbols: SymbolSnapshot | undefined;
    private readonly functions: Map<string, ConstructedHIR | undefined> = new Map<string, ConstructedHIR | undefined>();
    constructor(source: string, symbols: SymbolSnapshot | undefined, path: string = '/test.tsx') {
        this.source = source; this.symbols = symbols; this.parser = new Parser(source, path);
        this.parser.file();
    }
    forFunction(index: number): ConstructedHIR | undefined {
        if(this.symbols === undefined || index < 0) { return undefined; }
        const node = this.parser.node(index);
        const key = `${node.kind}:${node.pos}`;
        if(this.functions.has(key)) { return this.functions.get(key); }
        const graph = lowerParsedFunction(this.parser, index, this.source, this.symbols);
        this.functions.set(key, graph);
        return graph;
    }
    at(start: number, end: number): number {
        for(let index = 0; index < this.parser.nodes.length; index++) {
            const node = this.parser.node(index);
            if(['FunctionDeclaration', 'FunctionExpression', 'ArrowFunction'].includes(node.kind) && utf8Length(this.source.slice(0, node.pos)) === start && utf8Length(this.source.slice(0, node.end)) === end) { return index; }
        }
        return -1;
    }
}
export function ForFunction(file: HIRFile, index: number): ConstructedHIR | undefined { return file.forFunction(index); }
