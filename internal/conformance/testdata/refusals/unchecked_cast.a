// Spec: | `as` except `as const`, upcasts, and **checked downcasts** |
// Expect: refused a cast the runtime can't check
interface User {
	readonly name: string;
}

function load(): unknown {
	return 42;
}

const user = load() as User;
console.log(`${user.name.length}`);
