// A stop: an arrow written above the let it captures. Lowering lowers the arrow where it's written,
// before the let's declaration has made its local, so seen isn't a local yet when the arrow reads it
// (expression.go's "reading seen"), the stop getters_census_lazy_object.a has with an object. Only a
// function whose body has a nested function declaration has its lets' cells made first (hoisted.ts).
function early(): void {
	const show = (): string => `${seen}`;
	let seen: boolean | undefined;
	console.log(`early ${show()}`);
	seen = true;
	console.log(`early ${show()}`);
}

early();
