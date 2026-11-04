function make(): string { return new Late().label; }
class Late { label = 'late'; static { try { console.log(make()); } catch (error) { if (error instanceof Error) console.log(`${error.name}|${error.message}`); } } }
console.log(make());
