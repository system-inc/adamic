type Target = ((value: number) => number) | ((value: string | number) => number);
interface Receiver { readonly kind: 'Receiver'; readonly value: Target; }
interface Box { readonly kind: 'Receiver'; value: Target; }
const box: Box = { kind: 'Receiver', value: (value: number): number => value * 2 };
const view = (box as { readonly kind: 'Receiver' }) as Receiver;
console.log(`${view.value(4)}`);
