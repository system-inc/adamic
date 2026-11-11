// sys.ts:1481's shape: a let of a program class, a string literal and undefined, captured by the
// closures that write it and the ones that read it, as getNodeSystem's profiler closures do.
class Session {
	readonly name: string;
	constructor(name: string) {
		this.name = name;
	}
}

function profiler(): void {
	let active: Session | "stopping" | undefined;
	const see = (): string => {
		if (active === undefined) {
			return "undefined";
		}
		if (active === "stopping") {
			return "stopping";
		}
		return `session ${active.name}`;
	};
	const write = (step: number): void => {
		if (step === 0) {
			active = new Session("first");
		} else if (step === 1) {
			active = "stopping";
		} else {
			active = undefined;
		}
	};
	const truthy = (): boolean => (active ? true : false);
	console.log(`start ${see()} ${truthy()}`);
	for (let step = 0; step < 3; step++) {
		write(step);
		console.log(`step ${step} ${see()} ${truthy()} ${typeof active}`);
	}
	// enableCPUProfiler's test at sys.ts:1646 and disableCPUProfiler's at 1694.
	const enable = (): string => {
		if (active) {
			return active === "stopping" ? "enable while stopping" : `enable while ${active.name}`;
		}
		active = new Session("profiled");
		return "enabled";
	};
	const disable = (): string => {
		if (active && active !== "stopping") {
			const held = active;
			active = "stopping";
			const during = see();
			active = undefined;
			return `disabled ${held.name} during ${during}`;
		}
		return "nothing to disable";
	};
	console.log(enable());
	console.log(enable());
	console.log(disable());
	console.log(disable());
	write(1);
	console.log(enable());
	write(2);
	console.log(enable());
	// The declaring function reads the same cell the closures write.
	if (active !== undefined && active !== "stopping") {
		write(1);
		console.log(`after a write the cell says ${typeof active}`);
	}
	console.log(see());
}
profiler();
