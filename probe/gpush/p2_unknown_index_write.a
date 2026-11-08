// The same through an element write: T extends unknown[], so rest[1] takes unknown.
function relay<T extends unknown[]>(callback: (...args: T) => void, ...rest: T): void {
	rest[1] = 41;
	callback(...rest);
}
relay((first: string, second?: string) => { console.log(first.toUpperCase()); console.log(second === undefined ? "none" : second.toUpperCase()); }, "a", "b");
