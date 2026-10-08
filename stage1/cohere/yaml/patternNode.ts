// A small anchored schema-pattern tree, not a general RegExp API.
export class PatternNode {
    kind: string;
    characters = '';
    children: number[] = [];
    minimum = 0;
    maximum = 0;
    constructor(kind: string) {
        this.kind = kind;
    }
}
