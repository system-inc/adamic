const narrow: ('left' | 'right')[] = ['left'];
function store(values: string[], value: string): void { values.push(value); }
store(narrow, "outside");
console.log(narrow[1] ?? '');
