interface Animal {
	name: string;
}
interface Dog {
	name: string;
	bark: string;
}
function add<T>(to: T[], value: T): void {
	to.push(value);
}
const dogs: Dog[] = [{ name: "d", bark: "woof" }];
add(dogs, { name: "a" });
const last = dogs[1];
console.log(last === undefined ? "none" : String(last.bark));
