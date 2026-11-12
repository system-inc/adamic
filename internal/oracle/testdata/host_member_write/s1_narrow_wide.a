interface Narrow { read(): string; }
interface Wide { read(): string | undefined; }
const host: Narrow = { read(): string { return 'a'; } };
const wide: Wide = host;
wide.read = (): string | undefined => undefined;
console.log(host.read().toUpperCase());
