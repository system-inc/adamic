import type { Types } from './types.ts';

export class Parameters {
    readonly facts: Types;
    readonly fixed: readonly number[];
    readonly rest: number;
    readonly restIndex: number;
    index = 0;
    consumed = false;
    constructor(facts: Types, fixed: readonly number[], rest: number, restIndex: number) {
        this.facts = facts;
        this.fixed = fixed;
        this.rest = rest;
        this.restIndex = restIndex;
    }
    next(): number {
        const index = this.index;
        this.index++;
        if(index < this.fixed.length && !this.consumed) {
            return this.fixed[index] ?? 0;
        }
        if(this.rest === 0) {
            return 0;
        }
        const rest = this.facts.type(this.rest);
        if(rest.tuple >= 0) {
            if(rest.arguments.length === 0) {
                return 0;
            }
            return this.consumed || index - this.restIndex < 0
                ? (rest.arguments[rest.arguments.length - 1] ?? 0)
                : (rest.arguments[index - this.restIndex] ?? rest.arguments[rest.arguments.length - 1] ?? 0);
        }
        return rest.element === 0 ? rest.id : rest.element;
    }
}
