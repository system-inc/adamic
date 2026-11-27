// A stop: holder.e is narrowed to never across flip, which writes it, so the read isn't trusted and the
// call keeps its stop (enumNeverReadGap, #afs4xvb).
enum Mode {
	A,
	B,
}
interface Holder {
	mode: Mode;
}
function assertNever(value: never): never {
	throw new Error("Illegal value: " + value);
}
function flip(holder: Holder): void {
	holder.mode = Mode.A;
}
function modeName(holder: Holder): string {
	if (holder.mode === Mode.A) return "a";
	flip(holder);
	if (holder.mode === Mode.B) return "b";
	return assertNever(holder.mode);
}
console.log(modeName({ mode: Mode.A }));
