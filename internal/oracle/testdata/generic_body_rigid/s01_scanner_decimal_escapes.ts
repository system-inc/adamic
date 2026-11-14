function append<T extends {}>(to: T[], value: T | undefined): T[];
function append<T extends {}>(to: T[] | undefined, value: T): T[];
function append<T extends {}>(to: T[] | undefined, value: T | undefined): T[] | undefined;
function append<T extends {}>(to: T[] | undefined, value: T | undefined): T[] | undefined {
	if (value === undefined) return to;
	if (to === undefined) return [value];
	to.push(value);
	return to;
}
interface DecimalEscape {
	pos: number;
	end: number;
	value: number;
}
function scanDecimalEscapes(text: string): DecimalEscape[] | undefined {
	let decimalEscapes: DecimalEscape[] | undefined;
	for (let pos = 0; pos < text.length; pos++) {
		const end = pos + 1;
		const value = text.charCodeAt(pos) - 48;
		decimalEscapes = append(decimalEscapes, { pos, end, value });
	}
	return decimalEscapes;
}
const escapes = scanDecimalEscapes("172");
if (escapes !== undefined) {
	for (const escape of escapes) {
		console.log(`${String(escape.pos)} ${String(escape.end)} ${String(escape.value)}`);
	}
}
console.log(scanDecimalEscapes("") === undefined ? "none" : "some");
