// for (const { a, b } of items) reads each field once per step, in the pattern's order, as Node does (#rm22wzv).
class Item {
    readonly label: string;
    constructor(label: string) {
        this.label = label;
    }
    get a(): number {
        console.log("read a of " + this.label);
        return this.label.length;
    }
    get b(): string {
        console.log("read b of " + this.label);
        return this.label + "!";
    }
}
const items = [new Item("one"), new Item("three")];
for (const { a, b } of items) {
    console.log(String(a) + " " + b);
}
