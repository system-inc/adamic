interface A { id: number; note?: string }
interface B { id: number; note: string; extra: boolean }
interface C { id: number; count: number }
interface Noted { id: number; note: string }
function isNoted(x: A | B | C): x is Noted {
	return x.id > 1;
}
const values: (A | B | C)[] = [{ id: 1 }, { id: 2, note: "two", extra: true }, { id: 3, note: "three" }];
let found = "";
for (const value of values) {
	if (isNoted(value)) {
		found += value.note + " ";
	}
}
console.log(found);
