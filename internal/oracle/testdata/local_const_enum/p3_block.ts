function pick(flag: boolean): string {
	if (flag) {
		const enum Side { Left = "l", Right = "r" }
		return Side.Left + Side.Right;
	}
	return "none";
}
console.log(`${pick(true)} ${pick(false)}`);
