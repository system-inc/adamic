interface User {
	readonly name: string;
}

function load(): unknown {
	return 42;
}

const user = load() as User;
console.log(`${user.name.length}`);
