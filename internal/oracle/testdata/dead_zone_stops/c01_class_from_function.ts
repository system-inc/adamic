function make(): string { return new Late().label; }
try { console.log(make()); } catch (error) { if (error instanceof Error) console.log(`${error.name}|${error.message}`); }
class Late { label = 'late'; }
console.log(make());
