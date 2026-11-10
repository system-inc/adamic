namespace N {
	export function f(): void {
		console.log('orig');
	}
}
const swapped = (): void => {
	console.log('swapped');
};
(<() => void>N.f) = swapped;
N.f();
