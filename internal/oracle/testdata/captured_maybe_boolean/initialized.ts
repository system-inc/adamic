// uninitialized.ts's twin: the let's initializer is an explicit undefined, which a closure reads
// from the cell before anything writes it.
function run(): void {
	let seen: boolean | undefined = undefined;
	const show = (): string => `${seen}`;
	console.log(`initialized ${show()}`);
	seen = true;
	console.log(`initialized ${show()}`);
}

run();
