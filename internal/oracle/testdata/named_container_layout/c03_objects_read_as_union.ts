interface Point { x: number }
const points: Point[] = [{ x: 1 }];
const view: readonly (Point | string)[] = points;
for (const p of view) { console.log(typeof p === "string" ? p : String(p.x)); }
