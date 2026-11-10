export namespace Trace {
	export function push(): string { return "push"; }
	export function pop(): string { return "pop"; }
}
export const real: typeof Trace = Trace;
const fake: typeof Trace = { pop: (): string => "fake pop", push: (): string => "fake push" };
let pick: typeof Trace | undefined = fake;
console.log(pick?.push() ?? "none");
console.log(real.push());
