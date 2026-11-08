// A generic rest array is written through its constraint's push: T extends unknown[], so push takes unknown.
function relay<T extends unknown[]>(callback: (...args: T) => void, ...rest: T): void {
	rest.push(41);
	callback(...rest);
}
relay((first: string, second?: string) => { console.log(first.toUpperCase()); console.log(second === undefined ? "none" : second.toUpperCase()); }, "a");
