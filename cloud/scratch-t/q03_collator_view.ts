const collator: { compare(a: string, b: string): number } = new Intl.Collator(undefined, { usage: 'sort', sensitivity: 'variant', numeric: true });
console.log(collator.toString());
