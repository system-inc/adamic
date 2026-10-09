const ages = new Map<string, number>([['Kirk', 42]]);
const age = ages.get('Ahra')!;
console.log(`next year Ahra is ${age + 1}`);
