// '#message' and 'message' in a caught Error made without a message and one made with 'x', then a plain
// object's '#ordinary' key. Node prints false, false, true, true, true: a string key never names a
// private, Error.prototype supplies message, and '#ordinary' is an ordinary string key. Natively the
// message-less Error keeps its inherited "" in a slot named #message (exceptions.c), and on main
// '#message' in it was true (#2emrsmd).
function hashMessage(e: unknown): boolean {
	return typeof e === 'object' && e !== null && '#message' in e;
}
function message(e: unknown): boolean {
	return typeof e === 'object' && e !== null && 'message' in e;
}
try {
	throw new Error();
} catch (e) {
	console.log(`${hashMessage(e)}`);
}
try {
	throw new Error('x');
} catch (e) {
	console.log(`${hashMessage(e)}`);
}
try {
	throw new Error();
} catch (e) {
	console.log(`${message(e)}`);
}
try {
	throw new Error('x');
} catch (e) {
	console.log(`${message(e)}`);
}
console.log(`${'#ordinary' in {'#ordinary': 1}}`);
