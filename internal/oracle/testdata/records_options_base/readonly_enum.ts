// #6c4swvt piece 4, B2's road kept: an enumeration's object is the first value put into this readonly
// string-index type, so it stays held as the object it views (enumerationReadonlyType), and a for...in and a
// keyed read over it agree with Node.
enum Kind {
    A = 1,
    B = 2,
}
function rows(table: Readonly<Record<string, string | number>>): string {
    const shown: string[] = [];
    for (const name in table) {
        shown.push(name + "=" + String(table[name]));
    }
    return shown.join(",");
}
console.log(rows(Kind));
