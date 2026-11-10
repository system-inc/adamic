export namespace Trace {
	export function push(): string { return "push"; }
	export function pop(): string { return "pop"; }
}
export const real: typeof Trace = Trace;
function call(t: typeof Trace): string { return t.push(); }
console.log(call({ pop: (): string => "arg pop", push: (): string => "arg push" }));
console.log(call(Trace));
