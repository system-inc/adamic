import type { Repair } from './repair.ts';
export class Suggestion {
    readonly suggestionTag = true;
    readonly id: string;
    readonly message: string;
    readonly repairs: readonly Repair[];
    constructor(id: string, message: string, repairs: readonly Repair[]) {
        this.id = id;
        this.message = message;
        this.repairs = repairs;
    }
}
