// a-check: refused a void value at the initializer site through [] seen as unknown
const make: () => void = () => 42; const all: void[] = [1, 2].map(() => make());
const seen: readonly unknown[] = all; console.log(String(seen.length));
