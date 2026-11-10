namespace N {
	export function f(): void {
		console.log('orig');
	}
}
const swapped = (): void => {
	console.log('swapped');
};
({ f: N.f } = { f: swapped });
N.f();
