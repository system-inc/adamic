function never1(x: never): string {
	return "never1 " + typeof x;
}
function receive(callback: (x: never) => string): string {
	return "received";
}
console.log(receive(never1));
