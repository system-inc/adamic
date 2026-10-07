import { distance, pathLength, type Point } from './geometry.ts';

const route: readonly Point[] = [
	{ x: 0, y: 0 },
	{ x: 3, y: 4 },
	{ x: 3, y: 10 },
	{ x: 0, y: 14 },
];
console.log(`legs: ${distance({ x: 0, y: 0 }, { x: 3, y: 4 })} first, ${pathLength(route)} total`);
console.log(`diagonal: ${distance({ x: 0, y: 0 }, { x: 1, y: 1 })}`);
