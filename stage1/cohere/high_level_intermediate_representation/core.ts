import { panic } from 'adamic';
// Go high_level_intermediate_representation.go: owned tables and function-local IDs.
import type { BlockIdType, IdentifierIdType, DeclarationIdType, EvaluationOrderType, PhiInterface } from '../static_single_assignment/static_single_assignment.ts';
import { FunctionIndex, BlockIndex, InstructionIndex, IdentifierIndex, DeclarationIndex, PatternIndex } from '../arena/arena_index.a';
export { FunctionIndex, BlockIndex, InstructionIndex, IdentifierIndex, DeclarationIndex, PatternIndex } from '../arena/arena_index.a';
export type EffectType = '<unknown>' | 'freeze' | 'read' | 'capture' | 'mutate-iterator?' | 'mutate?' | 'mutate' | 'store';
export interface PlaceInterface {
    readonly identifier: IdentifierIndex;
    readonly effect: EffectType;
    readonly reactive: boolean;
    readonly start: number;
    readonly end: number;
}
export interface IdentifierInterface {
    nodeIndex: number;
    readonly id: IdentifierIndex;
    readonly declaration: DeclarationIndex;
    readonly name: string;
}
// The slice admits only these variants; unsupported syntax is declined before lowering.
export interface FunctionReferenceInterface { readonly index: FunctionIndex; readonly ordinal: number; }
export interface ArgumentInterface { place: PlaceInterface; readonly spread: boolean; }
export interface ArrayElementInterface { place: PlaceInterface; readonly spread: boolean; readonly hole: boolean; }
export interface ObjectPropertyInterface { readonly key: string; computedKey: PlaceInterface | undefined; value: PlaceInterface; readonly spread: boolean; }
export interface JsxTagInterface { readonly name: string; place: PlaceInterface | undefined; }
export interface JsxAttributeInterface { readonly name: string; value: PlaceInterface; readonly spread: boolean; }
export interface ModuleExportOriginInterface { readonly module: string; readonly exported: string; }
export interface PatternPropertyInterface { readonly key: string; computedKey: PlaceInterface | undefined; defaultValue: PlaceInterface | undefined; readonly value: PatternIndex; }
export interface PatternElementInterface { readonly value: PatternIndex | undefined; defaultValue: PlaceInterface | undefined; }
export type PatternType = { readonly kind: 'Place'; place: PlaceInterface } | { readonly kind: 'Object'; readonly properties: PatternPropertyInterface[]; rest: PlaceInterface | undefined } | { readonly kind: 'Array'; readonly elements: PatternElementInterface[]; rest: PlaceInterface | undefined };
export type ValueType = { readonly kind: 'Destructure'; lvaluePattern: PatternIndex; readonly pattern: PatternIndex; value: PlaceInterface; readonly declarationKind: number }
    | { readonly kind: 'ObjectMethod'; readonly key: string; readonly functionReference: FunctionReferenceInterface }
    | { readonly kind: 'UnsupportedNode'; readonly nodeKind: string; readonly nodePos: number; readonly nodeEnd: number; readonly reason: string }
    | { readonly kind: 'Debugger' }
    | { readonly kind: 'MetaProperty'; readonly meta: string; readonly property: string }
    | { readonly kind: 'TypeCastExpression'; value: PlaceInterface; readonly nodeKind: string; readonly nodePos: number; readonly nodeEnd: number }
    | { readonly kind: 'Await'; value: PlaceInterface }
    | { readonly kind: 'PropertyDelete'; object: PlaceInterface; readonly property: string }
    | { readonly kind: 'ComputedDelete'; object: PlaceInterface; property: PlaceInterface }
    | { readonly kind: 'RegExpLiteral'; readonly pattern: string; readonly flags: string }
    | { readonly kind: 'TemplateLiteral'; readonly quasis: string[]; readonly subexprs: PlaceInterface[] }
    | { readonly kind: 'TaggedTemplateExpression'; tag: PlaceInterface; readonly quasis: string[]; readonly subexprs: PlaceInterface[] }
    | { readonly kind: 'Primitive'; readonly literal: string } | { readonly kind: 'LoadLocal' | 'LoadContext'; place: PlaceInterface }
    | { readonly kind: 'UnaryExpression'; readonly operator: string; value: PlaceInterface }
    | { readonly kind: 'BinaryExpression'; left: PlaceInterface; readonly operator: string; right: PlaceInterface }
    | { readonly kind: 'LoadGlobal'; readonly name: string; readonly bindingKind: number; readonly source: string; readonly imported: string }
    | { readonly kind: 'StoreGlobal'; readonly name: string; value: PlaceInterface }
    | { readonly kind: 'DeclareLocal'; lvalue: PlaceInterface; readonly declarationKind: number }
    | { readonly kind: 'StoreLocal' | 'StoreContext'; lvalue: PlaceInterface; value: PlaceInterface; readonly declarationKind: number }
    | { readonly kind: 'PrefixUpdate' | 'PostfixUpdate'; lvalue: PlaceInterface; readonly operation: string; value: PlaceInterface }
    | { readonly kind: 'PropertyLoad'; object: PlaceInterface; readonly property: string; readonly optional: boolean }
    | { readonly kind: 'ComputedLoad'; object: PlaceInterface; property: PlaceInterface; readonly optional: boolean }
    | { readonly kind: 'PropertyStore'; object: PlaceInterface; readonly property: string; value: PlaceInterface }
    | { readonly kind: 'ComputedStore'; object: PlaceInterface; property: PlaceInterface; value: PlaceInterface }
    | { readonly kind: 'CallExpression'; callee: PlaceInterface; readonly args: ArgumentInterface[]; readonly optional: boolean; readonly origin: ModuleExportOriginInterface }
    | { readonly kind: 'MethodCall'; receiver: PlaceInterface; property: PlaceInterface; readonly args: ArgumentInterface[]; readonly optional: boolean; readonly origin: ModuleExportOriginInterface }
    | { readonly kind: 'NewExpression'; callee: PlaceInterface; readonly args: ArgumentInterface[] }
    | { readonly kind: 'GetIterator' | 'NextPropertyOf'; value: PlaceInterface }
    | { readonly kind: 'IteratorNext'; iterator: PlaceInterface; collection: PlaceInterface }
    | { readonly kind: 'ArrayExpression'; readonly elements: ArrayElementInterface[] }
    | { readonly kind: 'ObjectExpression'; readonly properties: ObjectPropertyInterface[] }
    | { readonly kind: 'JsxExpression'; readonly tag: JsxTagInterface; readonly props: JsxAttributeInterface[]; readonly children: PlaceInterface[] }
    | { readonly kind: 'JsxFragment'; readonly children: PlaceInterface[] }
    | { readonly kind: 'JsxText'; readonly text: string }
    | { readonly kind: 'FunctionExpression'; readonly functionReference: FunctionReferenceInterface; readonly captures: PlaceInterface[] };
export class Instruction {
    readonly id: InstructionIndex;
    order: EvaluationOrderType = 0;
    lvalue: PlaceInterface;
    readonly value: ValueType;
    readonly start: number;
    readonly end: number;
    constructor(id: InstructionIndex, lvalue: PlaceInterface, value: ValueType, start: number, end: number) {
        this.id = id; this.lvalue = lvalue; this.value = value; this.start = start; this.end = end;
    }
}
// gap 1 (GAPS.md, ruling (a)): presence and value are required booleans.
export interface OptionalFlagInterface { readonly present: boolean; readonly value: boolean; }
// A tagged arena record; accessors validate the fields required by each terminal kind.
export interface TerminalType {
    readonly kind: 'Optional' | 'Return' | 'Throw' | 'Unreachable' | 'Unsupported' | 'Goto' | 'If' | 'Branch' | 'Logical' | 'Ternary' | 'While' | 'DoWhile' | 'For' | 'ForOf' | 'ForIn' | 'Switch' | 'Label' | 'Try';
    value?: PlaceInterface | undefined;
    readonly block?: BlockIndex | undefined;
    readonly variant?: number | undefined;
    testPlace?: PlaceInterface | undefined;
    readonly testBlock?: BlockIndex | undefined;
    readonly consequent?: BlockIndex | undefined;
    readonly alternate?: BlockIndex | undefined;
    readonly fallthrough?: BlockIndex | undefined;
    // gap 1 (GAPS.md): the optional boolean is carried as a presence/value record.
    readonly optionalFlag?: OptionalFlagInterface | undefined;
    readonly operator?: string | undefined;
    readonly loop?: BlockIndex | undefined;
    readonly init?: BlockIndex | undefined;
    readonly update?: BlockIndex | undefined;
    readonly cases?: { test: PlaceInterface | undefined; readonly block: BlockIndex }[] | undefined;
    readonly handler?: BlockIndex | undefined;
    handlerBinding?: PlaceInterface | undefined;
}
export function testPlace(terminal: TerminalType): PlaceInterface { return terminal.testPlace ?? panic('terminal requires a test place'); }
export function testBlock(terminal: TerminalType): BlockIndex { return terminal.testBlock ?? panic('terminal requires a test block'); }
export class BasicBlock {
    readonly arenaIndices: readonly BlockIndex[];
    readonly id: BlockIndex;
    readonly instructions: InstructionIndex[] = [];
    predecessors: readonly BlockIndex[] = [];
    phis: readonly PhiInterface<PlaceInterface>[] = [];
    terminal: TerminalType;
    readonly kind: string;
    terminalOrder: EvaluationOrderType = 0;
    constructor(arenaIndices: readonly BlockIndex[], id: BlockIndex, terminal: TerminalType, kind: string = 'block') { this.arenaIndices = arenaIndices; this.id = id; this.terminal = terminal; this.kind = kind; }
}
export class HIRFunction {
    nodeIndex = -1;
    isAsync = false;
    isGenerator = false;
    readonly name: string;
    readonly kind: string;
    readonly blockIndices: BlockIndex[] = [];
    readonly patternIndices: PatternIndex[] = [];
    readonly patterns: PatternType[] = [];
    pattern(index: PatternIndex): PatternType { return this.patterns[PatternIndex.read(this.patternIndices, index)] ?? panic('missing pattern'); }
    addPattern(value: PatternType): PatternIndex { const index = PatternIndex.push(this.patternIndices); this.patterns.push(value); return index; }
    readonly instructionIndices: InstructionIndex[] = [];
    readonly identifierIndices: IdentifierIndex[] = [];
    readonly declarationIndices: DeclarationIndex[] = [];
    readonly entry: BlockIndex;
    readonly blockTable: BasicBlock[] = [];
    blockOrder: BlockIndex[] = [];
    readonly retained: Set<number> = new Set<number>();
    readonly instructions: Instruction[] = [];
    readonly identifiers: IdentifierInterface[] = [];
    get nextBlock(): number { return this.blockIndices.length + 1; }
    readonly params: PlaceInterface[] = [];
    readonly context: PlaceInterface[] = [];
    readonly functions: FunctionIndex[] = [];
    readonly contextDeclarations: Set<DeclarationIndex> = new Set<DeclarationIndex>();
    returns: PlaceInterface;
    constructor(name: string) {
        this.name = name;
        const first = name.charCodeAt(0);
        const next = name.charCodeAt(3);
        this.kind = first >= 65 && first <= 90 ? 'component' : name.startsWith('use') && ((next >= 65 && next <= 90) || (next >= 48 && next <= 57)) ? 'hook' : 'other';
        const identifier = IdentifierIndex.push(this.identifierIndices);
        const declaration = DeclarationIndex.push(this.declarationIndices);
        this.identifiers.push({ id: identifier, declaration, nodeIndex: -1, name: '' });
        this.returns = { identifier, effect: '<unknown>', reactive: false, start: 0, end: 0 };
        this.entry = BlockIndex.push(this.blockIndices);
        this.blockTable.push(new BasicBlock(this.blockIndices, this.entry, { kind: 'Return', value: this.returns }));
        this.blockOrder.push(this.entry); this.retained.add(1);
    }
    named(name: string, start: number, end: number, declaration: DeclarationIndex | undefined = undefined): PlaceInterface {
        const id = IdentifierIndex.push(this.identifierIndices);
        const fresh = DeclarationIndex.push(this.declarationIndices);
        const declared = declaration ?? fresh;
        DeclarationIndex.read(this.declarationIndices, declared);
        this.identifiers.push({ id, declaration: declared, name, nodeIndex: -1 });
        return { identifier: id, effect: '<unknown>', reactive: false, start, end };
    }
    newBlock(kind: string): BasicBlock {
        const id = BlockIndex.push(this.blockIndices);
        const block = new BasicBlock(this.blockIndices, id, { kind: 'Unreachable' }, kind);
        this.blockTable.push(block); this.blockOrder.push(id); this.retained.add(id.slot + 1); return block;
    }
    blockAt(value: number): BlockIndex {
        if(!Number.isInteger(value) || value < 1) { panic('invalid SSA block id'); }
        const index = this.blockIndices[value - 1] ?? panic('SSA block id out of range'); BlockIndex.read(this.blockIndices, index); return index;
    }
    identifierAt(value: number): IdentifierIndex {
        if(!Number.isInteger(value) || value < 0) { panic('invalid SSA identifier id'); }
        const index = this.identifierIndices[value] ?? panic('SSA identifier id out of range'); IdentifierIndex.read(this.identifierIndices, index); return index;
    }
    declarationAt(value: number): DeclarationIndex {
        if(!Number.isInteger(value) || value < 1) { panic('invalid SSA declaration id'); }
        const index = this.declarationIndices[value - 1] ?? panic('SSA declaration id out of range'); DeclarationIndex.read(this.declarationIndices, index); return index;
    }
    block(id: BlockIndex): BasicBlock { return this.blockTable[BlockIndex.read(this.blockIndices, id)] ?? panic('missing arena block'); }
    identifier(id: IdentifierIndex): IdentifierInterface { return this.identifiers[IdentifierIndex.read(this.identifierIndices, id)] ?? panic('missing arena identifier'); }
    instruction(id: InstructionIndex): Instruction { return this.instructions[InstructionIndex.read(this.instructionIndices, id)] ?? panic('missing arena instruction'); }
    temporary(start: number, end: number, declaration: DeclarationIndex | undefined = undefined): PlaceInterface { return this.named('', start, end, declaration); }
    mint(original: IdentifierIndex): IdentifierIndex {
        const old = this.identifier(original); const id = IdentifierIndex.push(this.identifierIndices);
        this.identifiers.push({ id, declaration: old.declaration, name: old.name, nodeIndex: old.nodeIndex }); return id;
    }
    emit(place: PlaceInterface, value: ValueType, start: number, end: number): InstructionIndex {
        const id = InstructionIndex.push(this.instructionIndices);
        this.instructions.push(new Instruction(id, place, value, start, end)); return id;
    }

}

// The graph owner is the only escaping result. Nodes contain indices, never arena links.
export class HIRArena {
    private readonly nodes: HIRFunction[] = [];
    private readonly indices: FunctionIndex[] = [];
    create(name: string): FunctionIndex { const index = FunctionIndex.push(this.indices); this.nodes.push(new HIRFunction(name)); return index; }
    read(index: FunctionIndex): HIRFunction { return this.nodes[FunctionIndex.read(this.indices, index)] ?? panic('missing arena function'); }

}
export class ConstructedHIR {
    readonly arena: HIRArena;
    readonly root: FunctionIndex;
    constructor(arena: HIRArena, root: FunctionIndex) { this.arena = arena; this.root = root; }
}
