import type { ExtraEdit } from './extra_edit.ts';
export class RepairSuggestion {
    readonly id: string;
    readonly message: string;
    readonly edits: ExtraEdit[];
    constructor(id: string, message: string, edits: ExtraEdit[]) {
        this.id = id;
        this.message = message;
        this.edits = edits;
    }
}
