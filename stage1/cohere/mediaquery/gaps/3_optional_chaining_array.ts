// Gap 3: ?. on an array (or a string) is refused; on an object it lowers.
function size(list: readonly string[] | undefined): number {
	return list?.length ?? 0;
}
console.log(`${size(['a'])} ${size(undefined)}`);
