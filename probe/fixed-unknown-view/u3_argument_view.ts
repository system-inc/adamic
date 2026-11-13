function run(f: (cb: () => void, ms: number, label: string) => string): string { return f(() => {}, 0, 'x'); }
const host = (cb: () => void, ms: number, label?: unknown): string => typeof label + ' ' + String(label) + ' ' + String(ms);
console.log(run(host));
