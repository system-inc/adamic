export interface Point {
	readonly x: number;
	readonly y: number;
}

export function distance(from: Point, to: Point): number {
	return Math.sqrt((to.x - from.x) ** 2 + (to.y - from.y) ** 2);
}

export function pathLength(points: readonly Point[]): number {
	let length = 0;
	let previous: Point | undefined = undefined;
	for (const point of points) {
		if (previous !== undefined) {
			length += distance(previous, point);
		}
		previous = point;
	}
	return length;
}
