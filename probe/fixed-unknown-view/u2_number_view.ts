const show = (label?: unknown): string => typeof label + ' ' + String(label);
const view: (label: number) => string = show;
console.log(view(7));
