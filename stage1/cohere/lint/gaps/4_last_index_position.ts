class Search {
    readonly source: string;
    constructor(source: string) {
        this.source = source;
    }
    find(position: number): number {
        return this.source.lastIndexOf('(', position);
    }
}
console.log(`${new Search('a(a(').find(1)}`);
