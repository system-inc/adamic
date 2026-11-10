namespace Debug {
	export function fail(message?: string): never {
		throw new Error(message);
	}
}
function assertIsDefined<T>(value: T, message?: string): asserts value is NonNullable<T> {
	if (value === undefined || value === null) {
		Debug.fail(message);
	}
}
const H = {
	stop(): never {
		throw new Error('h');
	},
};
const G: { stop: () => void } = H;
G.stop = (): void => {};
const liar = (): never => H.stop();
const values: (string | undefined)[] = [undefined];
const value = values[0];
assertIsDefined(value, 'missing');
console.log(value.toUpperCase());
