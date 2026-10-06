import type { RepairEdit } from './repair_edit.ts';
export class RepairSuggestion {
    readonly id: string;
    readonly message: string;
    readonly edits: RepairEdit[];
    constructor(id: string, message: string, edits: RepairEdit[]) {
        this.id = id;
        this.message = message;
        this.edits = edits;
    }
}
