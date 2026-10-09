export class Binding {
    readonly name: string;
    readonly identifier: number;
    readonly declarations: readonly number[];
    readonly flags: number;
    constructor(name: string, identifier: number, declarations: readonly number[], flags: number) {
        this.name = name;
        this.identifier = identifier;
        this.declarations = declarations;
        this.flags = flags;
    }
}
