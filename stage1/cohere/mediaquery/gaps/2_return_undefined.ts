// Gap 2: `return undefined` from a function returning string | undefined lowers, and the C it becomes
// returns an adamic_object * from a function declared to return adamic_string *, which clang refuses.
function nothing(): string | undefined {
	return undefined;
}
console.log(nothing() ?? 'none');
