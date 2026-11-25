function never1(x: never): string {
	return "never1 " + typeof x;
}
console.log(never1("lie" as never));
