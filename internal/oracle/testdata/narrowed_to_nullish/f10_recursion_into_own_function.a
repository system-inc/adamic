const o: { x: number | undefined } = { x: 1 };
function recur(n: number): void {
    if (n === 0) { o.x = 4; return; }
    o.x = undefined;
    recur(0);
    console.log(String(o.x));
}
recur(1);
