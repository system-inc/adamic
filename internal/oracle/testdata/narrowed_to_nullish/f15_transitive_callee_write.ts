interface Box { x: number | undefined; }
function inner(box: Box): void { box.x = 4; }
function outer(box: Box): void { inner(box); }
const box: Box = { x: 1 };
box.x = undefined;
outer(box);
console.log(String(box.x));
