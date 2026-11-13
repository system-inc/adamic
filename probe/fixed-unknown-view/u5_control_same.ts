const show = (label?: unknown): string => typeof label + ' ' + String(label);
const view: (label?: unknown) => string = show;
console.log(view('x'));
