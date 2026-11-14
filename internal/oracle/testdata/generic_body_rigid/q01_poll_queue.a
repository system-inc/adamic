interface Base {
	isClosed: boolean;
	name: string;
}
interface Sub extends Base {
	unchangedPolls: number;
}
function pollQ<T extends Base>(queue: (T | undefined)[], i: number): void {
	const item = queue[i];
	if (item === undefined || item.isClosed) {
		queue[i] = undefined;
		return;
	}
	queue[i] = undefined;
	queue.push(item);
}
const queue: (Sub | undefined)[] = [
	{ isClosed: false, name: "a", unchangedPolls: 1 },
	{ isClosed: true, name: "b", unchangedPolls: 2 },
];
pollQ(queue, 0);
pollQ(queue, 1);
for (const item of queue) {
	console.log(item === undefined ? "none" : item.name + " " + String(item.unchangedPolls));
}
