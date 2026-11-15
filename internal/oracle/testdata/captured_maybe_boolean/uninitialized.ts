// A let of boolean | undefined declared without an initializer, captured, is read by a closure
// before anything writes it, so its cell has to start as undefined, not as false.
function fresh(): void {
	let seen: boolean | undefined;
	const show = (): string => `${seen}`;
	console.log(`fresh ${show()}`);
	seen = false;
	console.log(`fresh ${show()}`);
}

fresh();
