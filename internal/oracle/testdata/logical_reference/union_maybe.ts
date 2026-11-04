// An && chain whose left is a union the checker types boolean | undefined keeps its falsy value as
// that pair: false where the left was false, undefined where it was missing. Printed, not only
// tested, so the two can't be confused (#cme8paq, tsc's getTypeAnnotationFromAllAccessorDeclarations).
interface Accessor {
	readonly name: string;
}
interface Accessors {
	readonly first: Accessor;
	readonly second: Accessor | undefined;
}
function annotation(accessor: Accessor): string | undefined {
	return accessor.name === "typed" ? "num" + "ber" : undefined;
}
function chain(node: Accessor, accessors: Accessors): boolean | undefined {
	const found = annotation(node);
	return !found && accessors.second && node !== accessors.second;
}
function annotationOfAll(node: Accessor, accessors: Accessors): string | undefined {
	let found = annotation(node);
	if (!found && node !== accessors.first) {
		found = annotation(accessors.first);
	}
	if (!found && accessors.second && node !== accessors.second) {
		found = annotation(accessors.second);
	}
	return found;
}
const first: Accessor = { name: "un" + "typed" };
const second: Accessor = { name: "ty" + "ped" };
console.log(annotationOfAll(first, { first: first, second: second }) ?? "none");
console.log(annotationOfAll(first, { first: first, second: undefined }) ?? "none");
console.log(String(chain(first, { first: first, second: second })));
console.log(String(chain(first, { first: first, second: undefined })));
console.log(String(chain(second, { first: first, second: second })));
