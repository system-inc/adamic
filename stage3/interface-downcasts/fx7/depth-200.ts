interface Link { readonly next: Link | undefined; readonly value: number; }
interface Other { readonly other: string; }
interface Receiver { readonly kind: 'Receiver'; readonly value: Link | Other; }
let chain: Link = {next: undefined, value: 0};
for (let i = 1; i < 200; i++) chain = {next: chain, value: i};
const raw = {kind: 'Receiver' as const, value: chain};
const base: {readonly kind: 'Receiver'} = raw;
const view = base as Receiver;
console.log(String(view.value !== undefined));
