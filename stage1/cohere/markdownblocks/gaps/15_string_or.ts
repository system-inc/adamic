function choose(value: string): string { return value || 'fallback'; }
console.log(choose(''));
console.log(choose('x'));
