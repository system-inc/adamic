function collect(...items: string[]): string {
	return items.map((item) => item).join(";");
}
console.log(collect("a", "b"));
