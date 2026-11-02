// A TypeError made and thrown, through a call, in the try itself, and from a const bound to one, and
// caught: its name and message are Node's, natively and on the JavaScript backend (#bbhwj30).
function fail(n: number): void {
	throw new TypeError('bad ' + String(n));
}
try {
	fail(3);
} catch (error) {
	if (error instanceof Error) {
		console.log(error.name);
		console.log(error.message);
		console.log(`${error.name}: ${error.message}`);
	}
}
const n = 4;
try {
	throw new TypeError('direct ' + String(n));
} catch (e) {
	if (e instanceof Error) {
		console.log(e.name);
		console.log(e.message);
	}
}
const failure = new TypeError('stored');
try {
	throw failure;
} catch (error) {
	if (error instanceof Error) {
		console.log(`${error.name} ${error.message}`);
	}
}
