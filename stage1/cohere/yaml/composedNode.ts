// Composer ownership is an arena; all recursive node and CST links are numeric.
export class ComposedNode {
    kind: string;
    className: string;
    range: number[] = [];
    sourceToken = -1;
    sourceItemCollection = -1;
    sourceItem = -1;
    anchor = '';
    anchorPresent = false;
    tag = '';
    comment = '';
    commentBefore = '';
    spaceBefore = false;
    flow = false;
    items: number[] = [];
    key = -1;
    value = -1;
    source = '';
    sourcePresent = false;
    type = '';
    format = '';
    minFractionDigits = 0;
    valueKind = 'null';
    numberValue = 0;
    stringValue = '';
    booleanValue = false;
    binaryValue: number[] = [];
    constructor(kind: string, className: string) {
        this.kind = kind;
        this.className = className;
    }
}
