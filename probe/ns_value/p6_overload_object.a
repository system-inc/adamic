export namespace Trace {
	export function push(): string { return "push"; }
	export function pop(): string { return "pop"; }
}
export const real: typeof Trace = Trace;
function keys(t: typeof Trace): string;
function keys(t: object): string { return Object.keys(t).join(","); }
console.log(keys(Trace));
