interface Reader { readonly extra: string; }
class Sneaky {
    readonly label: string;
    constructor(label: string) { this.label = label; }
    get extra(): string { kept.push(this); return "-" + this.label; }
}
const kept: Sneaky[] = [];
function make(n: number): string { const sneaky = new Sneaky("made" + String(n)); const reader: Reader = sneaky; return reader.extra; }
console.log(make(1));
console.log(make(2));
for (const sneaky of kept) { console.log(sneaky.label); }
