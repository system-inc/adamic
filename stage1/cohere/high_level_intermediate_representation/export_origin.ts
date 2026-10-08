// Generic checker graph; every cyclic symbol/module edge is a branded numeric handle.
// Resolver ports cohere/internal/lint/rule/export_origin.go, without checker-side answers.
import { panic } from 'adamic';
import { Frames, header } from '../typeaware/frames.ts';
import { SymbolIndex, ModuleIndex, ExpressionIndex } from '../arena/arena_index.a';
interface DeclarationInterface { readonly kind: string; readonly name: string; readonly property: string; readonly module: ModuleIndex | undefined; readonly constant: boolean; readonly initializer: ExpressionIndex | undefined; readonly blocked: boolean; }
interface ExportInterface { readonly name: string; readonly symbol: SymbolIndex | undefined; }
interface SymbolInterface { readonly immediate: SymbolIndex | undefined; readonly resolved: SymbolIndex | undefined; readonly declarations: DeclarationInterface[]; readonly exports: ExportInterface[]; readonly stars: ModuleIndex[]; }
interface ModuleInterface { readonly name: string; readonly symbol: SymbolIndex | undefined; }
interface ExpressionInterface { readonly kind: string; readonly symbol: SymbolIndex | undefined; readonly inner: ExpressionIndex | undefined; readonly propertyKind: string; readonly property: string; }
export interface OriginInterface { readonly module: ModuleIndex | undefined; readonly name: string; }
export const emptyOrigin: OriginInterface = { module: undefined, name: '' };
export class SymbolGraph {
    private readonly symbols: SymbolInterface[] = [];
    private readonly modules: ModuleInterface[] = [];
    private readonly expressions: ExpressionInterface[] = [];
    private readonly symbolIndices: SymbolIndex[] = [];
    private readonly moduleIndices: ModuleIndex[] = [];
    private readonly expressionIndices: ExpressionIndex[] = [];
    constructor(wire: string) {
        const frames = new Frames(wire); header(frames, 'symbol-graph');
        // Wire ids are protocol numbers, decoded before arena edges are installed.
        const rawSymbols: { immediate: number; resolved: number; declarations: { kind: string; name: string; property: string; module: number; constant: boolean; initializer: number; blocked: boolean }[]; exports: { name: string; symbol: number }[]; stars: number[] }[] = [];
        const count = frames.natural();
        for(let index = 0; index < count; index++) {
            SymbolIndex.push(this.symbolIndices);
            const immediate = frames.natural(); const resolved = frames.natural();
            const declarations: { kind: string; name: string; property: string; module: number; constant: boolean; initializer: number; blocked: boolean }[] = [];
            const declarationCount = frames.natural();
            for(let offset = 0; offset < declarationCount; offset++) { declarations.push({ kind: frames.field(), name: frames.field(), property: frames.field(), module: frames.natural(), constant: frames.yes(), initializer: frames.natural(), blocked: frames.yes() }); }
            const exports: { name: string; symbol: number }[] = []; const exportCount = frames.natural();
            for(let offset = 0; offset < exportCount; offset++) { exports.push({ name: frames.field(), symbol: frames.natural() }); }
            const stars: number[] = []; const starCount = frames.natural(); for(let offset = 0; offset < starCount; offset++) { stars.push(frames.natural()); }
            rawSymbols.push({ immediate, resolved, declarations, exports, stars });
        }
        const rawModules: { name: string; symbol: number }[] = []; const moduleCount = frames.natural();
        for(let index = 0; index < moduleCount; index++) { ModuleIndex.push(this.moduleIndices); rawModules.push({ name: frames.field(), symbol: frames.natural() }); }
        const rawExpressions: { kind: string; symbol: number; inner: number; propertyKind: string; property: string }[] = []; const expressionCount = frames.natural();
        for(let index = 0; index < expressionCount; index++) { ExpressionIndex.push(this.expressionIndices); rawExpressions.push({ kind: frames.field(), symbol: frames.natural(), inner: frames.natural(), propertyKind: frames.field(), property: frames.field() }); }
        frames.end();
        for(const raw of rawSymbols) {
            const declarations: DeclarationInterface[] = [];
            for(const declaration of raw.declarations) { declarations.push({ kind: declaration.kind, name: declaration.name, property: declaration.property, module: this.moduleReference(declaration.module), constant: declaration.constant, initializer: this.expressionReference(declaration.initializer), blocked: declaration.blocked }); }
            const exports: ExportInterface[] = []; for(const entry of raw.exports) { exports.push({ name: entry.name, symbol: this.symbolReference(entry.symbol) }); }
            const stars: ModuleIndex[] = []; for(const star of raw.stars) { stars.push(this.moduleReference(star) ?? panic('absent star module')); }
            this.symbols.push({ immediate: this.symbolReference(raw.immediate), resolved: this.symbolReference(raw.resolved), declarations, exports, stars });
        }
        for(const raw of rawModules) { this.modules.push({ name: raw.name, symbol: this.symbolReference(raw.symbol) }); }
        for(const raw of rawExpressions) { this.expressions.push({ kind: raw.kind, symbol: this.symbolReference(raw.symbol), inner: this.expressionReference(raw.inner), propertyKind: raw.propertyKind, property: raw.property }); }
    }
    symbolReference(value: number): SymbolIndex | undefined { if(value === 0) { return undefined; } if(!Number.isInteger(value) || value < 1) { panic('invalid wire symbol id'); } const index = this.symbolIndices[value - 1] ?? panic('wire symbol id out of range'); SymbolIndex.read(this.symbolIndices, index); return index; }
    moduleReference(value: number): ModuleIndex | undefined { if(value === 0) { return undefined; } if(!Number.isInteger(value) || value < 1) { panic('invalid wire module id'); } const index = this.moduleIndices[value - 1] ?? panic('wire module id out of range'); ModuleIndex.read(this.moduleIndices, index); return index; }
    expressionReference(value: number): ExpressionIndex | undefined { if(value === 0) { return undefined; } if(!Number.isInteger(value) || value < 1) { panic('invalid wire expression id'); } const index = this.expressionIndices[value - 1] ?? panic('wire expression id out of range'); ExpressionIndex.read(this.expressionIndices, index); return index; }
    symbol(index: SymbolIndex): SymbolInterface { return this.symbols[SymbolIndex.read(this.symbolIndices, index)] ?? panic('missing symbol graph node'); }
    module(index: ModuleIndex): ModuleInterface { return this.modules[ModuleIndex.read(this.moduleIndices, index)] ?? panic('missing module graph node'); }
    expression(index: ExpressionIndex): ExpressionInterface { return this.expressions[ExpressionIndex.read(this.expressionIndices, index)] ?? panic('missing expression graph node'); }

}
export class ExportResolver {
    readonly graph: SymbolGraph;
    readonly target: string;
    private readonly activeSymbols: Set<SymbolIndex> = new Set<SymbolIndex>();
    // Preserves Go's (module node identity, export name) key without struct-key stringification.
    private readonly activeExports: Map<ModuleIndex, Set<string>> = new Map<ModuleIndex, Set<string>>();
    constructor(graph: SymbolGraph, target: string) { this.graph = graph; this.target = target; }
    symbol(index: SymbolIndex | undefined): OriginInterface {
        if(index === undefined || this.activeSymbols.has(index)) { return emptyOrigin; }
        this.activeSymbols.add(index); const symbol = this.graph.symbol(index); let result = emptyOrigin;
        for(const declaration of symbol.declarations) {
            if(declaration.blocked) { panic('export origin read a foreign function body outside shape facts'); }
            if(['ImportSpecifier', 'ImportClause', 'NamespaceImport', 'ExportSpecifier'].includes(declaration.kind) && declaration.module !== undefined) {
                if(declaration.kind === 'NamespaceImport') { result = { module: declaration.module, name: '*' }; break; }
                const name = declaration.kind === 'ImportClause' ? 'default' : declaration.property === '' ? declaration.name : declaration.property;
                result = this.export(declaration.module, name);
            }
            else if(declaration.kind === 'VariableDeclaration' && declaration.constant) { result = this.expression(declaration.initializer); }
            if(result.module !== undefined) { break; }
        }
        if(result.module === undefined && symbol.immediate !== undefined) { result = this.symbol(symbol.immediate); }
        this.activeSymbols.delete(index); return result;
    }
    expression(index: ExpressionIndex | undefined): OriginInterface {
        if(index === undefined) { return emptyOrigin; } const expression = this.graph.expression(index);
        if(expression.kind === 'Identifier') { return this.symbol(expression.symbol); }
        if(['ParenthesizedExpression', 'AsExpression', 'TypeAssertionExpression', 'NonNullExpression', 'SatisfiesExpression'].includes(expression.kind)) { return this.expression(expression.inner); }
        if(expression.kind === 'PropertyAccessExpression' || expression.kind === 'ElementAccessExpression') {
            const receiver = this.expression(expression.inner);
            if(receiver.module !== undefined && (receiver.name === '*' || receiver.name === 'default') && expression.property !== '' && (expression.kind === 'PropertyAccessExpression' || expression.propertyKind === 'StringLiteral' || expression.propertyKind === 'NoSubstitutionTemplateLiteral')) { return this.export(receiver.module, expression.property); }
        }
        return emptyOrigin;
    }
    exported(module: SymbolIndex | undefined, name: string): SymbolIndex | undefined { if(module !== undefined) { for(const entry of this.graph.symbol(module).exports) { if(entry.name === name) { return entry.symbol; } } } return undefined; }
    export(index: ModuleIndex, name: string): OriginInterface {
        const module = this.graph.module(index);
        if(module.name === this.target) { return { module: index, name }; }
        let active = this.activeExports.get(index); if(active === undefined) { active = new Set<string>(); this.activeExports.set(index, active); }
        if(active.has(name)) { return emptyOrigin; } active.add(name);
        const exported = this.exported(module.symbol, name); let result = emptyOrigin;
        if(exported !== undefined) {
            result = this.symbol(exported);
            if(result.module === undefined && module.symbol !== undefined) {
                for(const star of this.graph.symbol(module.symbol).stars) {
                    const candidate = this.exported(this.graph.module(star).symbol, name);
                    if(candidate !== undefined && this.graph.symbol(candidate).resolved === this.graph.symbol(exported).resolved) { result = this.export(star, name); if(result.module !== undefined) { break; } }
                }
            }
        }
        active.delete(name); return result;
    }
}
