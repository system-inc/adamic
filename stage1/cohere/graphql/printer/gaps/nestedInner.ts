export class Inner {
    readonly label: string;
    constructor(label: string) { this.label = label; }
    read(): string { return this.label; }
    show(): string { return this.read(); }
}
