function show(f: Function) { console.log(String(f.length)); }
function g(x: number, y: number, z = 1) { return x + y + z; }
show(g.bind(null, 1));
