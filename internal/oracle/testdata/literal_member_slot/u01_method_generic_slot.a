interface Mapper { map<U>(f: (x: number) => U): U[] }
const mapper: Mapper = { map<U>(f: (x: number) => U): U[] { return [f(1)]; } };
console.log(String(mapper.map((x: number) => x + 1)[0]));
