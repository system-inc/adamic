// ?. on a local the checker proves present: the result is held as TypeScript types it, number.
const xs: number[] = [];
xs?.push(1);
const n = xs?.push(1, 2);
const typed: number = xs?.push(3);
console.log(`${xs.length} ${n} ${typed} ${xs?.indexOf(2)} ${xs?.includes(3)}`);
