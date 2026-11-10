namespace N {
	export function f(): void {
		console.log('orig');
	}
}
const swapped = (): void => {
	console.log('swapped');
};
for (N.f of [swapped]) {}
N.f();
