interface Box { x: number | null | undefined; }
function fill(box: Box): void { box.x = 4; }
const box: Box = { x: 1 };
box.x = undefined;
fill(box);
if (box.x !== null) console.log(String(box.x));
