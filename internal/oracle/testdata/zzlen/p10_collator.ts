function arity(f: Function): number { return f.length; }
const collator = new Intl.Collator("en");
console.log(String(arity(collator.compare)));
