namespace N {
	export function f(): void {
		console.log('orig');
	}
}
const swapped = (): void => {
	console.log('swapped');
};
(N.f satisfies () => void) = swapped;
N.f();
