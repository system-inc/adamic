export class Repair {
    readonly start: number;
    readonly end: number;
    readonly text: string;
    constructor(start: number, end: number, text: string) {
        this.start = start;
        this.end = end;
        this.text = text;
    }
}
