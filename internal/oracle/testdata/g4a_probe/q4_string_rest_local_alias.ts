const collect = (...items: string[]): string => {
	const xs = items;
	return xs.map((item) => item).join(";");
};
console.log(collect("a", "b"));
