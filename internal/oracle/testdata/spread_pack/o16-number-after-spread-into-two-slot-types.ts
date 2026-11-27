type A = number | string;
function show(first?: number, ...rest: A[]): void { console.log(String(first) + ":" + String(rest.length)); }
show(...[], 7);
