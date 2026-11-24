async function one(): Promise<number> {
	return 1;
}
console.log(Object.prototype.toString.call(one()));
