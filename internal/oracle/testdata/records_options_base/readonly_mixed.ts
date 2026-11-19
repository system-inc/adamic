// #6c4swvt piece 4, Runtime's must: one function compiled once reads its Readonly<Record<string, string |
// number>> parameter in one layout. The program first puts a record there, so it's held as a record, and the
// enumeration's object handed to the same parameter afterwards stops at its call, by name.
enum Kind {
    A = 1,
    B = 2,
}
function count(table: Readonly<Record<string, string | number>>): number {
    let length = 0;
    for (const name in table) {
        length += name.length;
    }
    return length;
}
const bag: Record<string, string | number> = {};
bag["one"] = 1;
console.log(String(count(bag)));
console.log(String(count(Kind)));
