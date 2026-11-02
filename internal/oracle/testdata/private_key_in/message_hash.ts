// "#message" in a caught Error made without a message: false on Node, where no string key reaches a
// private, but natively the Error keeps its inherited message in a #message slot (#a1b746w).
function has(e: unknown): boolean {
	return typeof e === 'object' && e !== null && '#message' in e;
}
try {
	throw new Error();
} catch (e) {
	console.log(`${has(e)}`);
}
