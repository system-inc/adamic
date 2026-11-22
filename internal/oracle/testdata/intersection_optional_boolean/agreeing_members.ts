// tsc's program.ts shape (#tq6yzqy): both members of the intersection declare isInvalidated?: boolean,
// the same way. Read present and absent. Node prints: module true, module false, module undefined.
interface ResolvedModule {
	isInvalidated?: boolean;
	module: string;
}
interface ResolvedDirective {
	isInvalidated?: boolean;
	directive: number;
}
type Resolved = ResolvedModule & ResolvedDirective;
function probe(value: Resolved): void {
	console.log(`${value.module} ${String(value.isInvalidated)}`);
}
probe({ module: "module", directive: 1, isInvalidated: true });
probe({ module: "module", directive: 2, isInvalidated: false });
probe({ module: "module", directive: 3 });
