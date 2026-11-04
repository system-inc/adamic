class Base { readonly label: string = 'base'; }
function make(): string { return new Late().label; }
function base(callback: () => string): typeof Base { try { console.log(callback()); } catch (error) { if (error instanceof Error) console.log(`${error.name}|${error.message}`); } return Base; }
class Late extends base(make) {}
console.log(make());
