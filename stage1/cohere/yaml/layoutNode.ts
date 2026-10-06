export class LayoutNode {
    kind: string;
    text = '';
    parts: number[] = [];
    broken = false;
    soft = false;
    literal = false;
    group = -1;
    conditional = false;
    amount = 0;
    width = 0;
    constructor(kind: string) {
        this.kind = kind;
    }
}
