export namespace Trace {
	export function push(): string { return "push"; }
	export function pop(): string { return "pop"; }
}
export const real: typeof Trace = Trace;
function make(): typeof Trace { return { pop: (): string => "made pop", push: (): string => "made push" }; }
const viewed = make();
console.log(viewed.push());
