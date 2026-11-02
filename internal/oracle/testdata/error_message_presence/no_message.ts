// An Error made without a message, caught as unknown and asked for its message: Node finds message on
// Error.prototype, as "", so 'message' in e is true. Natively, on main, it printed "none".
function msg(e: unknown): string {
	return typeof e === 'object' && e !== null && 'message' in e && typeof e.message === 'string' ? e.message : 'none';
}
try {
	throw new Error();
} catch (e) {
	console.log(JSON.stringify(msg(e)) ?? 'undefined');
}
