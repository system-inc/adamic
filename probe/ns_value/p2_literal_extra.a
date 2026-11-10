export namespace Trace {
	export function push(): string { return "push"; }
	export function pop(): string { return "pop"; }
}
export const real: typeof Trace = Trace;
const extra = { other: (): string => "other", pop: (): string => "extra pop", push: (): string => "extra push" };
const viewed: typeof Trace = extra;
console.log(viewed.push());
