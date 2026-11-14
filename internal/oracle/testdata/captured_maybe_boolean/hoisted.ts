// A cell made before its declaration: show is a nested function declaration, so the cells of the
// lets it captures are made when hoisted runs, before seen's declaration (nested_functions.go). It's
// called only once that declaration has run, which has to leave the cell undefined.
function hoisted(): void {
	function show(): string {
		return `${seen}`;
	}
	let seen: boolean | undefined;
	console.log(`hoisted ${show()}`);
	seen = true;
	console.log(`hoisted ${show()}`);
	seen = undefined;
	console.log(`hoisted ${show()}`);
}

hoisted();
