interface B {
	readonly n: number;
}
function show(value: B | null | void): void {
	console.log(String(value === null) + ' ' + String(value === undefined) + ' ' + typeof value);
}
show(null);
show(undefined);
show({ n: 1 });
