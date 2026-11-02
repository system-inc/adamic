// Object.entries of a plain object through the same view, in a program with no Error and no catch.
const o: { message: string } = { message: 'x' };
for (const [key, value] of Object.entries(o)) {
	console.log(`${key}:${value}`);
}
