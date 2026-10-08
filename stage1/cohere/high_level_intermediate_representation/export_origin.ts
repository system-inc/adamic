// Generic checker graph; every cyclic symbol/module edge is a branded numeric handle.
// Resolver ports cohere/internal/lint/rule/export_origin.go, without checker-side answers.
import { panic } from 'adamic';
import { Frames, header } from '../typeaware/frames.ts';
export enum SymbolIndex { Absent = 0 }
export enum ModuleIndex { Absent = 0 }
export enum ExpressionIndex { Absent = 0 }
interface DeclarationInterface { readonly kind: string; readonly name: string; readonly property: string; readonly module: ModuleIndex; readonly constant: boolean; readonly initializer: ExpressionIndex; readonly blocked: boolean; }
interface ExportInterface { readonly name: string; readonly symbol: SymbolIndex; }
interface SymbolInterface { readonly immediate: SymbolIndex; readonly resolved: SymbolIndex; readonly declarations: DeclarationInterface[]; readonly exports: ExportInterface[]; readonly stars: ModuleIndex[]; }
interface ModuleInterface { readonly name: string; readonly symbol: SymbolIndex; }
interface ExpressionInterface { readonly kind: string; readonly symbol: SymbolIndex; readonly inner: ExpressionIndex; readonly propertyKind: string; readonly property: string; }
export interface OriginInterface { readonly module: ModuleIndex; readonly name: string; }
export const emptyOrigin: OriginInterface = { module: 0, name: '' };
export class SymbolGraph {
    private readonly symbols: SymbolInterface[] = [];
    private readonly modules: ModuleInterface[] = [];
    private readonly expressions: ExpressionInterface[] = [];
    constructor(wire: string) {
        const frames = new Frames(wire); header(frames, 'symbol-graph'); const count = frames.natural();
        for(let index = 0; index < count; index++) {
            const immediate: SymbolIndex = frames.natural(); const resolved: SymbolIndex = frames.natural(); const declarations: DeclarationInterface[] = []; const declarationCount = frames.natural();
            for(let offset = 0; offset < declarationCount; offset++) { declarations.push({ kind: frames.field(), name: frames.field(), property: frames.field(), module: frames.natural(), constant: frames.yes(), initializer: frames.natural(), blocked: frames.yes() }); }
            const exports: ExportInterface[] = []; const exportCount = frames.natural(); for(let offset = 0; offset < exportCount; offset++) { exports.push({ name: frames.field(), symbol: frames.natural() }); }
            const stars: ModuleIndex[] = []; const starCount = frames.natural(); for(let offset = 0; offset < starCount; offset++) { const module: ModuleIndex = frames.natural(); stars.push(module); }
            this.symbols.push({ immediate, resolved, declarations, exports, stars });
        }
        const moduleCount = frames.natural(); for(let index = 0; index < moduleCount; index++) { this.modules.push({ name: frames.field(), symbol: frames.natural() }); }
        const expressionCount = frames.natural(); for(let index = 0; index < expressionCount; index++) { this.expressions.push({ kind: frames.field(), symbol: frames.natural(), inner: frames.natural(), propertyKind: frames.field(), property: frames.field() }); }
        frames.end();
    }
    symbol(index: SymbolIndex): SymbolInterface { if(!Number.isInteger(index) || index < 1 || index > this.symbols.length) { panic('symbol graph index out of range'); } return this.symbols[index - 1] ?? panic('missing symbol graph node'); }
    module(index: ModuleIndex): ModuleInterface { if(!Number.isInteger(index) || index < 1 || index > this.modules.length) { panic('module graph index out of range'); } return this.modules[index - 1] ?? panic('missing module graph node'); }
    expression(index: ExpressionIndex): ExpressionInterface { if(!Number.isInteger(index) || index < 1 || index > this.expressions.length) { panic('expression graph index out of range'); } return this.expressions[index - 1] ?? panic('missing expression graph node'); }
}
export class ExportResolver {
    readonly graph: SymbolGraph;
    readonly target: string;
    private readonly activeSymbols: Set<SymbolIndex> = new Set<SymbolIndex>();
    // Preserves Go's (module node identity, export name) key without struct-key stringification.
    private readonly activeExports: Map<ModuleIndex, Set<string>> = new Map<ModuleIndex, Set<string>>();
    constructor(graph: SymbolGraph, target: string) { this.graph = graph; this.target = target; }
    symbol(index: SymbolIndex): OriginInterface {
        if(index === 0 || this.activeSymbols.has(index)) { return emptyOrigin; }
        this.activeSymbols.add(index); const symbol = this.graph.symbol(index); let result = emptyOrigin;
        for(const declaration of symbol.declarations) {
            if(declaration.blocked) { panic('export origin read a foreign function body outside shape facts'); }
            if(['ImportSpecifier', 'ImportClause', 'NamespaceImport', 'ExportSpecifier'].includes(declaration.kind) && declaration.module !== 0) {
                if(declaration.kind === 'NamespaceImport') { result = { module: declaration.module, name: '*' }; break; }
                const name = declaration.kind === 'ImportClause' ? 'default' : declaration.property === '' ? declaration.name : declaration.property;
                result = this.export(declaration.module, name);
            }
            else if(declaration.kind === 'VariableDeclaration' && declaration.constant) { result = this.expression(declaration.initializer); }
            if(result.module !== 0) { break; }
        }
        if(result.module === 0 && symbol.immediate !== 0) { result = this.symbol(symbol.immediate); }
        this.activeSymbols.delete(index); return result;
    }
    expression(index: ExpressionIndex): OriginInterface {
        if(index === 0) { return emptyOrigin; } const expression = this.graph.expression(index);
        if(expression.kind === 'Identifier') { return this.symbol(expression.symbol); }
        if(['ParenthesizedExpression', 'AsExpression', 'TypeAssertionExpression', 'NonNullExpression', 'SatisfiesExpression'].includes(expression.kind)) { return this.expression(expression.inner); }
        if(expression.kind === 'PropertyAccessExpression' || expression.kind === 'ElementAccessExpression') {
            const receiver = this.expression(expression.inner);
            if(receiver.module !== 0 && (receiver.name === '*' || receiver.name === 'default') && expression.property !== '' && (expression.kind === 'PropertyAccessExpression' || expression.propertyKind === 'StringLiteral' || expression.propertyKind === 'NoSubstitutionTemplateLiteral')) { return this.export(receiver.module, expression.property); }
        }
        return emptyOrigin;
    }
    exported(module: SymbolIndex, name: string): SymbolIndex { if(module !== 0) { for(const entry of this.graph.symbol(module).exports) { if(entry.name === name) { return entry.symbol; } } } return 0; }
    export(index: ModuleIndex, name: string): OriginInterface {
        const module = this.graph.module(index);
        if(module.name === this.target) { return { module: index, name }; }
        let active = this.activeExports.get(index); if(active === undefined) { active = new Set<string>(); this.activeExports.set(index, active); }
        if(active.has(name)) { return emptyOrigin; } active.add(name);
        const exported = this.exported(module.symbol, name); let result = emptyOrigin;
        if(exported !== 0) {
            result = this.symbol(exported);
            if(result.module === 0 && module.symbol !== 0) {
                for(const star of this.graph.symbol(module.symbol).stars) {
                    const candidate = this.exported(this.graph.module(star).symbol, name);
                    if(candidate !== 0 && this.graph.symbol(candidate).resolved === this.graph.symbol(exported).resolved) { result = this.export(star, name); if(result.module !== 0) { break; } }
                }
            }
        }
        active.delete(name); return result;
    }
}
