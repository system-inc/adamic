export namespace Trace {
	export function push(): string { return "push"; }
	export function pop(): string { return "pop"; }
}
export const real: typeof Trace = Trace;
class Impl { pop(): string { return "impl pop"; } push(): string { return "impl push"; } }
const viewed: typeof Trace = new Impl();
console.log(viewed.push());
