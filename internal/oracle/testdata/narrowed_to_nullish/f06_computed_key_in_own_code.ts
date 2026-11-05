const o: { x: number | undefined } = { x: 1 };
function set(k: 'x'): void {
    o.x = undefined;
    o[k] = 4;
    console.log(String(o.x));
}
set('x');
