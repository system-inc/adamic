class Early { label = new Late().label; }
try { console.log(new Early().label); } catch (error) { if (error instanceof Error) console.log(`${error.name}|${error.message}`); }
class Late { label = 'late'; }
console.log(new Early().label);
