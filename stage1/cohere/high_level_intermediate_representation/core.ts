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
export type ValueType = { readonly kind: 'Primitive'; readonly literal: string } | { readonly kind: 'LoadLocal'; place: PlaceInterface };
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
export class BasicBlock {
    readonly id: BlockIdType;
    readonly instructions: number[] = [];
    predecessors: readonly BlockIdType[] = [];
    phis: readonly PhiInterface<PlaceInterface>[] = [];
    terminal: PlaceInterface;
    terminalOrder: EvaluationOrderType = 0;
    constructor(id: BlockIdType, terminal: PlaceInterface) { this.id = id; this.terminal = terminal; }
}
export class HIRFunction {
    readonly name: string;
    readonly kind: string;
    readonly entry: BlockIdType = 1;
    blocks: readonly BasicBlock[];
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
        this.blocks = [new BasicBlock(1, this.returns)];
        this.identifiers.push({ id: 0, declaration: 1, name: '' });
    }
    temporary(start: number, end: number): PlaceInterface {
        const id = this.identifiers.length;
        this.identifiers.push({ id, declaration: id + 1, name: '' });
        return { identifier: id, effect: '<unknown>', reactive: false, start, end };
    }
}
