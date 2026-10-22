interface Link { readonly next: Link | undefined; }
interface Other { readonly other: string; }
interface Receiver { readonly kind: 'R'; readonly value: Link | Other; }
let chain: Link = { next: undefined };
for (let i = 1; i < 2000; i++) chain = { next: chain };
const raw = { kind: 'R' as const, value: chain };
const hidden: { readonly kind: 'R' } = raw;
console.log(String((hidden as Receiver).value !== undefined));
