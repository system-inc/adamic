import { panic } from 'adamic';
// Go high_level_intermediate_representation.go: owned tables and function-local IDs.
import type { BlockIdType, IdentifierIdType, DeclarationIdType, EvaluationOrderType, PhiInterface } from '../static_single_assignment/static_single_assignment.ts';
export enum FunctionIndex { First = 0 }
export enum BlockIndex { Entry = 1 }
export enum InstructionIndex { First = 0 }
export enum IdentifierIndex { First = 0 }
export enum DeclarationIndex { First = 1 }
export function blockIndex(value: number): BlockIndex { if(!Number.isInteger(value) || value < 1) { panic(`invalid block index ${value}`); } const index: BlockIndex = value; return index; }
export type EffectType = '<unknown>' | 'freeze' | 'read' | 'capture' | 'mutate-iterator?' | 'mutate?' | 'mutate' | 'store';
export interface PlaceInterface {
    readonly identifier: IdentifierIndex;
    readonly effect: EffectType;
    readonly reactive: boolean;
    readonly start: number;
    readonly end: number;
}
export interface IdentifierInterface {
    readonly id: IdentifierIndex;
    readonly declaration: DeclarationIndex;
    readonly name: string;
}
// The slice admits only these variants; unsupported syntax is declined before lowering.
export interface FunctionReferenceInterface { readonly index: FunctionIndex; readonly ordinal: number; }
export interface ArgumentInterface { place: PlaceInterface; readonly spread: boolean; }
export interface ModuleExportOriginInterface { readonly module: string; readonly exported: string; }
export type ValueType = { readonly kind: 'Primitive'; readonly literal: string } | { readonly kind: 'LoadLocal' | 'LoadContext'; place: PlaceInterface }
    | { readonly kind: 'UnaryExpression'; readonly operator: string; value: PlaceInterface }
    | { readonly kind: 'BinaryExpression'; left: PlaceInterface; readonly operator: string; right: PlaceInterface }
    | { readonly kind: 'LoadGlobal'; readonly name: string; readonly bindingKind: number; readonly source: string; readonly imported: string }
    | { readonly kind: 'StoreGlobal'; readonly name: string; value: PlaceInterface }
    | { readonly kind: 'DeclareLocal'; lvalue: PlaceInterface; readonly declarationKind: number }
    | { readonly kind: 'StoreLocal' | 'StoreContext'; lvalue: PlaceInterface; value: PlaceInterface; readonly declarationKind: number }
    | { readonly kind: 'PrefixUpdate' | 'PostfixUpdate'; lvalue: PlaceInterface; readonly operation: string; value: PlaceInterface }
    | { readonly kind: 'PropertyLoad'; object: PlaceInterface; readonly property: string }
    | { readonly kind: 'ComputedLoad'; object: PlaceInterface; property: PlaceInterface }
    | { readonly kind: 'PropertyStore'; object: PlaceInterface; readonly property: string; value: PlaceInterface }
    | { readonly kind: 'ComputedStore'; object: PlaceInterface; property: PlaceInterface; value: PlaceInterface }
    | { readonly kind: 'CallExpression'; callee: PlaceInterface; readonly args: ArgumentInterface[]; readonly optional: boolean; readonly origin: ModuleExportOriginInterface }
    | { readonly kind: 'MethodCall'; receiver: PlaceInterface; property: PlaceInterface; readonly args: ArgumentInterface[]; readonly optional: boolean; readonly origin: ModuleExportOriginInterface }
    | { readonly kind: 'NewExpression'; callee: PlaceInterface; readonly args: ArgumentInterface[] }
    | { readonly kind: 'GetIterator' | 'NextPropertyOf'; value: PlaceInterface }
    | { readonly kind: 'IteratorNext'; iterator: PlaceInterface; collection: PlaceInterface }
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
// A tagged arena record; accessors validate the fields required by each terminal kind.
export interface TerminalType {
    readonly kind: 'Return' | 'Throw' | 'Unreachable' | 'Unsupported' | 'Goto' | 'If' | 'Branch' | 'Logical' | 'Ternary' | 'While' | 'DoWhile' | 'For' | 'ForOf' | 'ForIn' | 'Switch' | 'Label' | 'Try';
    value?: PlaceInterface;
    readonly block?: BlockIndex;
    readonly variant?: number;
    testPlace?: PlaceInterface;
    readonly testBlock?: BlockIndex;
    readonly consequent?: BlockIndex;
    readonly alternate?: BlockIndex;
    readonly fallthrough?: BlockIndex;
    readonly operator?: string;
    readonly loop?: BlockIndex;
    readonly init?: BlockIndex;
    readonly update?: BlockIndex;
    readonly cases?: { test: PlaceInterface | undefined; readonly block: BlockIndex }[];
    readonly handler?: BlockIndex;
    handlerBinding?: PlaceInterface | undefined;
}
export function testPlace(terminal: TerminalType): PlaceInterface { return terminal.testPlace ?? panic('terminal requires a test place'); }
export function testBlock(terminal: TerminalType): BlockIndex { return terminal.testBlock ?? panic('terminal requires a test block'); }
export class BasicBlock {
    readonly id: BlockIndex;
    readonly instructions: InstructionIndex[] = [];
    predecessors: readonly BlockIndex[] = [];
    phis: readonly PhiInterface<PlaceInterface>[] = [];
    terminal: TerminalType;
    readonly kind: string;
    terminalOrder: EvaluationOrderType = 0;
    constructor(id: BlockIndex, terminal: TerminalType, kind: string = 'block') { this.id = id; this.terminal = terminal; this.kind = kind; }
}
export class HIRFunction {
    readonly name: string;
    readonly kind: string;
    readonly entry: BlockIndex = blockIndex(1);
    readonly blockTable: BasicBlock[] = [];
    blockOrder: BlockIndex[] = [];
    readonly retained: Set<number> = new Set<number>();
    nextBlock = 2;
    readonly instructions: Instruction[] = [];
    readonly identifiers: IdentifierInterface[] = [];
    readonly params: PlaceInterface[] = [];
    readonly context: PlaceInterface[] = [];
    readonly functions: FunctionIndex[] = [];
    readonly contextDeclarations: Set<number> = new Set<number>();
    returns: PlaceInterface;
    constructor(name: string) {
        this.name = name;
        const first = name.charCodeAt(0);
        const next = name.charCodeAt(3);
        this.kind = first >= 65 && first <= 90 ? 'component' : name.startsWith('use') && ((next >= 65 && next <= 90) || (next >= 48 && next <= 57)) ? 'hook' : 'other';
        this.returns = { identifier: 0, effect: '<unknown>', reactive: false, start: 0, end: 0 };
        this.blockTable.push(new BasicBlock(blockIndex(1), { kind: 'Return', value: this.returns }));
        this.blockOrder.push(blockIndex(1)); this.retained.add(1);
        this.identifiers.push({ id: 0, declaration: 1, name: '' });

    }
    named(name: string, start: number, end: number, declaration: number = 0): PlaceInterface {
        const id = this.identifiers.length;
        this.identifiers.push({ id, declaration: declaration === 0 ? id + 1 : declaration, name });
        return { identifier: id, effect: '<unknown>', reactive: false, start, end };
    }
    newBlock(kind: string): BasicBlock {
        const block = new BasicBlock(blockIndex(this.nextBlock), { kind: 'Unreachable' }, kind);
        this.nextBlock++;
        this.blockTable.push(block); this.blockOrder.push(block.id); this.retained.add(block.id);
        return block;
    }
    block(id: BlockIndex): BasicBlock {
        if(!Number.isInteger(id) || id < 1 || id > this.blockTable.length) { panic(`block index ${id} out of range`); }
        return this.blockTable[id - 1] ?? panic('missing arena block');
    }
    temporary(start: number, end: number, declaration: number = 0): PlaceInterface {
        const id = this.identifiers.length;
        this.identifiers.push({ id, declaration: declaration === 0 ? id + 1 : declaration, name: '' });
        return { identifier: id, effect: '<unknown>', reactive: false, start, end };
    }
}

// The graph owner is the only escaping result. Nodes contain indices, never arena links.
export class HIRArena {
    private readonly nodes: HIRFunction[] = [];
    create(name: string): FunctionIndex { const index: FunctionIndex = this.nodes.length; this.nodes.push(new HIRFunction(name)); return index; }
    read(index: FunctionIndex): HIRFunction {
        if(!Number.isInteger(index) || index < 0 || index >= this.nodes.length) { panic(`function index ${index} out of range`); }
        return this.nodes[index] ?? panic('missing arena function');
    }
}
export class ConstructedHIR {
    readonly arena: HIRArena;
    readonly root: FunctionIndex;
    constructor(arena: HIRArena, root: FunctionIndex) { this.arena = arena; this.root = root; }
}
