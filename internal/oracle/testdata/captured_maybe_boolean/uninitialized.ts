// The first pin measures #fds729j's two inferences. fresh: a let of boolean | undefined declared
// without an initializer, captured, is read by a closure before anything writes it, so its cell has
// to start as undefined, not as false. early: the closure is made above the declaration, so the
// cell exists before the declaration runs, and it's called only once the declaration has run, so
// nothing reads the cell early.
function fresh(): void {
	let seen: boolean | undefined;
	const show = (): string => `${seen}`;
	console.log(`fresh ${show()}`);
	seen = false;
	console.log(`fresh ${show()}`);
}

function early(): void {
	const show = (): string => `${seen}`;
	let seen: boolean | undefined;
	console.log(`early ${show()}`);
	seen = true;
	console.log(`early ${show()}`);
}

fresh();
early();
