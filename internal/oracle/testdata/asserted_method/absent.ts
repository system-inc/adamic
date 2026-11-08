// A call through an optional field with a non-null assertion on its callee, box.run!(), after a call
// emptied the field. Since #0bxc4jh the asserted call keeps its object, as an unasserted one does, and
// its check is still the assertion's own: the same words and exit as when it called the field's value
// with no object (asserted_method_test.go).
interface Box {
	run?: (() => void) | undefined;
}
function hello(): void {
	console.log('hello');
}
const box: Box = { run: hello };
function clear(): void {
	box.run = undefined;
}
box.run!();
clear();
box.run!();
console.log('after');
