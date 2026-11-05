interface Box { x: number | null | undefined; }
function clear(box: Box): void { box.x = null; }
const box: Box = { x: 1 };
box.x = undefined;
clear(box);
if (box.x == null) console.log(String(box.x));
