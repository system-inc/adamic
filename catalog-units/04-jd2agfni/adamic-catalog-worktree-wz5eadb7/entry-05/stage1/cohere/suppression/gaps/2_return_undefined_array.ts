// Gap 2: mediaquery's gap 2 for an array. `return undefined` from a function returning
// readonly string[] | undefined lowers, and the C it becomes returns an adamic_object * from a function
// declared to return adamic_array *, which clang refuses. An object | undefined lowers and runs.
function names(found: boolean): readonly string[] | undefined {
	if (found) {
		return ['a'];
	}
	return undefined;
}
const got = names(false);
console.log(`${got === undefined}`);
