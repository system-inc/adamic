// Rule-owned repair plans retain precise edit spans while shared serialization catches up.
export class Edit {
    readonly start: number;
    readonly end: number;
    readonly text: string;
    constructor(start: number, end: number, text: string) {
        this.start = start;
        this.end = end;
        this.text = text;
    }
}

export class Suggestion {
    readonly id: string;
    readonly message: string;
    readonly edits: Edit[];
    constructor(id: string, message: string, edits: Edit[]) {
        this.id = id;
        this.message = message;
        this.edits = edits;
    }
}
