interface Point { x: number; y: number }
interface Loose { label: string; at: Point | undefined; size: number | string }
interface Placed { label: string; at: Point; size: number }
function isPlaced(item: Loose): item is Placed {
	return item.label.startsWith("p");
}
const items: Loose[] = [{ label: "p1", at: { x: 1, y: 2 }, size: 3 }, { label: "q", at: undefined, size: "big" }];
for (const item of items) {
	console.log(isPlaced(item) ? item.label + " " + String(item.at.x + item.at.y + item.size) : "loose " + item.label);
}
