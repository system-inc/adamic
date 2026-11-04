class Other { readonly label: string; constructor() { this.label = new Late().label; } }
class Late { label = 'late'; static { try { console.log(new Other().label); } catch (error) { if (error instanceof Error) console.log(`${error.name}|${error.message}`); } } }
console.log(new Other().label);
