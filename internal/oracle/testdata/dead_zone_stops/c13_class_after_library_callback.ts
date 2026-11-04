function make(value: number): string { return new Late().label + String(value); }
let labels = '';
try { labels = [1].map(make).join(); } catch (error) { labels = error instanceof Error ? error.name : 'other'; }
class Late { label = 'late'; }
console.log(labels);
console.log([1].map(make).join());
