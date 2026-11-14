Array.of<(a: string) => string>((...items: string[]): string => items.join(",")).forEach((f) => {
	const two: (a: string, b: number) => string = f;
	console.log(two("x", 2));
});
