function putField<T extends {}>(to: { v: T }, value: T): T {
	const wide: { v: {} } = to;
	wide.v = 1;
	return value;
}
const box: { v: string } = { v: "a" };
putField(box, "b");
console.log(typeof box.v);
