// Gap 5: an empty array literal as a default, list ?? [], is refused as an array of never.
function size(list: readonly string[] | undefined): number {
	return (list ?? []).length;
}
console.log(`${size(['a'])} ${size(undefined)}`);
