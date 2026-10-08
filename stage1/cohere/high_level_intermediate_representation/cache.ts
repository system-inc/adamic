// Go cache.go: one file/checker owns the cache; returned graph owners retain their arenas.
import { constructHIR } from './graph.ts';
import { utf8Length } from 'adamic';
import { lowerParsedFunction } from './lower.ts';
import { Parser } from '../../typescript/parser/parser.ts';
import type { SymbolSnapshot } from './symbol.ts';
import type { ConstructedHIR } from './core.ts';
// The lane-owned memo algorithm plugs into this owner-owned cache contract.
export interface ManualMemoizationPasses {
    readonly drop: (graph: ConstructedHIR) => void;
    readonly inlineIncludingMemoCallbacks: (graph: ConstructedHIR) => number;
}
export class HIRFile {
    readonly source: string;
    readonly cloneBeforeSSA: boolean;
    readonly parser: Parser;
    readonly symbols: SymbolSnapshot | undefined;
    private readonly functions: Map<string, ConstructedHIR | undefined> = new Map<string, ConstructedHIR | undefined>();
    private memoFunctions: Map<string, ConstructedHIR | undefined> | undefined = undefined;
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
    withoutManualMemoization(index: number,drop: (graph: ConstructedHIR) => void,inclusiveInline: (graph: ConstructedHIR) => number): ConstructedHIR | undefined {
        if(this.symbols === undefined || index < 0) { return undefined; }
        if(!MayNameManualMemoization(this,index)) { return this.forFunction(index); }
        const node = this.parser.node(index); const key = `${node.kind}:${node.pos}:no-manual-memo`;
        const cache = this.memoFunctions ?? new Map<string,ConstructedHIR | undefined>(); this.memoFunctions = cache;
        if(cache.has(key)) { return cache.get(key); }
        const graph = LowerFunction(this,index);
        if(graph !== undefined) {
            constructHIR(graph.arena,graph.root);
            drop(graph);
            if(inclusiveInline(graph) > 0) { constructHIR(graph.arena,graph.root); }
        }
        cache.set(key,graph); return graph;
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

export function LowerFunction(file: HIRFile,index: number): ConstructedHIR | undefined {
    if(index < 0) { return undefined; }
    return lowerParsedFunction(file.parser,index,file.source,file.symbols,false,false);
}
export function ForFunctionWithoutManualMemoization(file: HIRFile,index: number,drop: (graph: ConstructedHIR) => void,inclusiveInline: (graph: ConstructedHIR) => number): ConstructedHIR | undefined {
    return file.withoutManualMemoization(index,drop,inclusiveInline);
}
export function MayNameManualMemoization(file: HIRFile,index: number): boolean {
    const node = file.parser.node(index);
    if(node.pos < 0 || node.end > file.source.length || node.pos >= node.end) { return true; }
    const text = file.source.slice(node.pos,node.end);
    if(text.includes('useMemo') || text.includes('useCallback')) { return true; }
    if(!text.includes('\\')) { return false; }
    const pending = [index];
    while(pending.length > 0) {
        const child = file.parser.node(pending.pop() ?? -1);
        if((child.kind === 'Identifier' || child.kind === 'PrivateIdentifier' || child.kind === 'StringLiteral' || child.kind === 'NoSubstitutionTemplateLiteral') && (child.text === 'useMemo' || child.text === 'useCallback')) { return true; }
        for(const id of child.children) { pending.push(id); }
    }
    return false;
}
