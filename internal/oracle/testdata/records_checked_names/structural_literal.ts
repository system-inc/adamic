// Adamic's structural case: TypeScript lets an object literal stand where Settings is declared, so a
// Settings receiver is no proof its value is the class's. The proof is the type's alone, and the
// program refuses the literal anyway: a class type takes only what the class constructs.
interface Table { [key: string]: number | string; total: number }
const table: Table = { total: 2 };
table["mode"] = 1;
for (const key of Object.keys(table)) console.log(key);
interface Options { readonly total?: number; readonly mode?: 1 | 2 }
class Settings {
	total = 3;
	mode: 1 | 2 = 2;
}
function read<T extends Options>(t: T): string {
	return String(t.mode);
}
const literal: Settings = { total: 3, mode: 2 };
console.log(read(literal));
