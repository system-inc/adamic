function initial(): string | undefined { return undefined; }
function read(value: string = initial()!): void { console.log(`${value}`); }
read();
