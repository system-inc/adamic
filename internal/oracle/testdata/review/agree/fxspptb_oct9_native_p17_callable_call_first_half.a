type Target = ((value: number) => number) | ((value: string | number) => number);
interface Receiver { readonly kind: 'Receiver'; readonly value: Target; }
function make(seed: string): Target {
	const suffix = seed.repeat(3);
	const raw = { kind: 'Receiver' as const, value: (value: number): number => value + suffix.length };
	const base: { readonly kind: 'Receiver' } = raw;
	const view = base as Receiver;
	return view.value;
}
const callable = make('ab');
const again = make('xyz');
console.log(`${callable(10)}:${again(100)}`);
