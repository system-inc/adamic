function never1(x: never): string {
	return "never1 " + typeof x;
}
interface Runner {
	run(x: string): string;
}
const r: Runner = { run: never1 };
console.log(r.run("hi"));
