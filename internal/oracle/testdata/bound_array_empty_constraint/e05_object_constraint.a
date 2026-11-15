function keep<T extends object>(value: T): T {
	return value;
}
const kept = keep([{ a: 1 }, { a: 2 }]);
let sum = 0;
for (const item of kept) {
	sum += item.a;
}
console.log(`${String(kept.length)} ${String(sum)}`);
