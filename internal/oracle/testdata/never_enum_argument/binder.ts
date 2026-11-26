// binder.ts:405: a const from a call, inside an arrow passed to forEachChild, read in the default of a
// switch over every member and passed to assertNever.
const enum ModuleInstanceState {
	NonInstantiated = 0,
	Instantiated = 1,
	ConstEnumOnly = 2,
}
function assertNever(value: never): never {
	throw new Error("Illegal value: " + value);
}
function stateOf(n: number): ModuleInstanceState {
	if (n === 0) return ModuleInstanceState.NonInstantiated;
	if (n === 1) return ModuleInstanceState.Instantiated;
	if (n === 2) return ModuleInstanceState.ConstEnumOnly;
	return n as ModuleInstanceState;
}
function forEachChild(children: number[], visit: (child: number) => boolean | undefined): boolean | undefined {
	for (const child of children) {
		const result = visit(child);
		if (result) return result;
	}
	return undefined;
}
function moduleState(children: number[]): ModuleInstanceState {
	let state = ModuleInstanceState.NonInstantiated;
	forEachChild(children, n => {
		const childState = stateOf(n);
		switch (childState) {
			case ModuleInstanceState.NonInstantiated:
				return;
			case ModuleInstanceState.ConstEnumOnly:
				state = ModuleInstanceState.ConstEnumOnly;
				return;
			case ModuleInstanceState.Instantiated:
				state = ModuleInstanceState.Instantiated;
				return true;
			default:
				assertNever(childState);
		}
	});
	return state;
}
console.log(moduleState([0, 2]) + " " + moduleState([0, 1, 2]) + " " + moduleState([0]));
