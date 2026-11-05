interface Box { x: number | undefined; }
function fill(box: Box): boolean { box.x = 4; return true; }
const box: Box = { x: undefined };
if (box.x === undefined && fill(box)) console.log(String(box.x));
