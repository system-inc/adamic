const show = (label?: unknown): string => typeof label + ' ' + JSON.stringify(label);
const view: (label: { n: number }) => string = show;
console.log(view({ n: 1 }));
