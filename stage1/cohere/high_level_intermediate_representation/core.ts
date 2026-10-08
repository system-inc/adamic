// Go high_level_intermediate_representation.go: owned tables and function-local IDs.
import type { BlockIdType, IdentifierIdType, DeclarationIdType, EvaluationOrderType, PhiInterface } from '../static_single_assignment/static_single_assignment.ts';
export type EffectType = '<unknown>' | 'freeze' | 'read' | 'capture' | 'mutate-iterator?' | 'mutate?' | 'mutate' | 'store';
export interface PlaceInterface {
    readonly identifier: IdentifierIdType;
    readonly effect: EffectType;
    readonly reactive: boolean;
    readonly start: number;
    readonly end: number;
}
export interface IdentifierInterface {
    readonly id: IdentifierIdType;
    readonly declaration: DeclarationIdType;
    readonly name: string;
}
// The slice admits only these variants; unsupported syntax is declined before lowering.
export type ValueType = { readonly kind: 'Primitive'; readonly literal: string } | { readonly kind: 'LoadLocal'; place: PlaceInterface }
    | { readonly kind: 'UnaryExpression'; readonly operator: string; value: PlaceInterface }
    | { readonly kind: 'BinaryExpression'; left: PlaceInterface; readonly operator: string; right: PlaceInterface }
    | { readonly kind: 'LoadGlobal'; readonly name: string; readonly bindingKind: number; readonly source: string; readonly imported: string }
    | { readonly kind: 'StoreGlobal'; readonly name: string; value: PlaceInterface }
    | { readonly kind: 'DeclareLocal'; lvalue: PlaceInterface; readonly declarationKind: number }
    | { readonly kind: 'StoreLocal'; lvalue: PlaceInterface; value: PlaceInterface; readonly declarationKind: number }
    | { readonly kind: 'PrefixUpdate' | 'PostfixUpdate'; lvalue: PlaceInterface; readonly operation: string; value: PlaceInterface }
    | { readonly kind: 'PropertyLoad'; object: PlaceInterface; readonly property: string }
    | { readonly kind: 'ComputedLoad'; object: PlaceInterface; property: PlaceInterface }
    | { readonly kind: 'PropertyStore'; object: PlaceInterface; readonly property: string; value: PlaceInterface }
    | { readonly kind: 'ComputedStore'; object: PlaceInterface; property: PlaceInterface; value: PlaceInterface };
export class Instruction {
    readonly id: number;
    order: EvaluationOrderType = 0;
    lvalue: PlaceInterface;
    readonly value: ValueType;
    readonly start: number;
    readonly end: number;
    constructor(id: number, lvalue: PlaceInterface, value: ValueType, start: number, end: number) {
        this.id = id; this.lvalue = lvalue; this.value = value; this.start = start; this.end = end;
    }
}
export type TerminalType = { readonly kind: 'Return' | 'Throw'; value: PlaceInterface }
    | { readonly kind: 'Unreachable' | 'Unsupported' }
    | { readonly kind: 'Goto'; readonly block: number; readonly variant: number }
    | { readonly kind: 'If' | 'Branch'; test: PlaceInterface; readonly consequent: number; readonly alternate: number; readonly fallthrough: number }
    | { readonly kind: 'Logical'; readonly operator: string; readonly test: number; readonly fallthrough: number }
    | { readonly kind: 'Ternary'; readonly test: number; readonly fallthrough: number }
    | { readonly kind: 'While'; readonly test: number; readonly loop: number; readonly fallthrough: number };
export class BasicBlock {
    readonly id: BlockIdType;
    readonly instructions: number[] = [];
    predecessors: readonly BlockIdType[] = [];
    phis: readonly PhiInterface<PlaceInterface>[] = [];
    terminal: TerminalType;
    readonly kind: string;
    terminalOrder: EvaluationOrderType = 0;
    constructor(id: BlockIdType, terminal: TerminalType, kind: string = 'block') { this.id = id; this.terminal = terminal; this.kind = kind; }
}
export class HIRFunction {
    readonly name: string;
    readonly kind: string;
    readonly entry: BlockIdType = 1;
    blocks: readonly BasicBlock[];
    readonly byId: Map<number, BasicBlock> = new Map<number, BasicBlock>();
    nextBlock = 2;
    readonly instructions: Instruction[] = [];
    readonly identifiers: IdentifierInterface[] = [];
    readonly params: PlaceInterface[] = [];
    readonly context: PlaceInterface[] = [];
    readonly functions: HIRFunction[] = [];
    returns: PlaceInterface;
    constructor(name: string) {
        this.name = name;
        const first = name.charCodeAt(0);
        const next = name.charCodeAt(3);
        this.kind = first >= 65 && first <= 90 ? 'component' : name.startsWith('use') && ((next >= 65 && next <= 90) || (next >= 48 && next <= 57)) ? 'hook' : 'other';
        this.returns = { identifier: 0, effect: '<unknown>', reactive: false, start: 0, end: 0 };
        this.blocks = [new BasicBlock(1, { kind: 'Return', value: this.returns })];
        this.identifiers.push({ id: 0, declaration: 1, name: '' });
        this.byId.set(1, this.blocks[0] ?? new BasicBlock(1, { kind: 'Unreachable' }));
    }
    named(name: string, start: number, end: number, declaration: number = 0): PlaceInterface {
        const id = this.identifiers.length;
        this.identifiers.push({ id, declaration: declaration === 0 ? id + 1 : declaration, name });
        return { identifier: id, effect: '<unknown>', reactive: false, start, end };
    }
    newBlock(kind: string): BasicBlock {
        const block = new BasicBlock(this.nextBlock, { kind: 'Unreachable' }, kind);
        this.nextBlock++;
        this.blocks = [...this.blocks, block];
        this.byId.set(block.id, block);
        return block;
    }
    temporary(start: number, end: number, declaration: number = 0): PlaceInterface {
        const id = this.identifiers.length;
        this.identifiers.push({ id, declaration: declaration === 0 ? id + 1 : declaration, name: '' });
        return { identifier: id, effect: '<unknown>', reactive: false, start, end };
    }
}
