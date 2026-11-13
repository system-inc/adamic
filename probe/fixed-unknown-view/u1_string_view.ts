const show = (label?: unknown): string => typeof label + ' ' + String(label);
const view: (label: string) => string = show;
console.log(view('x'));
