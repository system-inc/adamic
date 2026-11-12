// class.ts with the object arm an object literal type, not a program class.
type Session = { readonly name: string };

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
			active = { name: "first" };
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
	const enable = (): string => {
		if (active) {
			return active === "stopping" ? "enable while stopping" : `enable while ${active.name}`;
		}
		active = { name: "profiled" };
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
	if (active !== undefined && active !== "stopping") {
		write(1);
		console.log(`after a write the cell says ${typeof active}`);
	}
	console.log(see());
}
profiler();
