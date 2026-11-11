interface B {
	readonly n: number;
}
const make: () => void = () => 42;
const widened: () => B | null | void = make;
const got = widened();
console.log(String(got === undefined));
