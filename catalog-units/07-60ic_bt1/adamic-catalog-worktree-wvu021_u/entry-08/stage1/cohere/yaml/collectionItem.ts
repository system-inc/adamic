// A CST collection item with explicit optional-field presence.
export class CollectionItem {
    start: number[] = [];
    key = -1;
    keyPresent = false;
    sep: number[] = [];
    sepPresent = false;
    value = -1;
    explicitKey = false;
    constructor(start: number[]) {
        this.start = start;
    }
}
