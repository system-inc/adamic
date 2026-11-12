// heap_churn: a million small objects built, summed and dropped, ten rounds, then idle (#w7cg3be).
// held is module-level, so every object lives on the heap's size classes, never in a statement's
// region. Each round's shape, two, four or seven fields, and its labels' length change with the
// round, so several classes churn. After the last round nothing is held, and the program spins for
// 400 ms. It reads its own resident memory three times and writes it to stderr, so stdout stays the
// same under every cap: at its peak, just after the last round is built; dropped, just after it's let
// go; and idle, after the spin, before anything exits.

import { readFileSync } from 'node:fs';
import { performance } from 'node:perf_hooks';

interface Item {
	readonly value: number;
	readonly label: string;
}

interface Medium extends Item {
	readonly weight: number;
	readonly rank: number;
}

interface Wide extends Medium {
	readonly left: number;
	readonly right: number;
	readonly depth: number;
}

const count = 1000000;
const rounds = 10;
const prefixes = ['k', 'item', 'a-longer-label-'];

function made(round: number, index: number): Item {
	const label = `${prefixes[round % prefixes.length] ?? 'k'}${index % 997}`;
	const shape = round % 3;
	if (shape === 0) {
		const item: Item = { value: index, label };
		return item;
	}
	if (shape === 1) {
		const item: Medium = { value: index, label, weight: round, rank: index % 7 };
		return item;
	}
	const item: Wide = { value: index, label, weight: round, rank: index % 7, left: index - 1, right: index + 1, depth: round + 1 };
	return item;
}

// residentKilobytes is VmRSS from /proc/self/status, or -1 where there's none (macOS).
function residentKilobytes(): number {
	try {
		const status = readFileSync('/proc/self/status', 'utf8');
		for (const line of status.split('\n')) {
			if (line.startsWith('VmRSS:')) {
				const value = line.slice(6).trim();
				const space = value.indexOf(' ');
				return Number(space < 0 ? value : value.slice(0, space));
			}
		}
	} catch (error) {
		return -1;
	}
	return -1;
}

let held: Item[] = [];
let peak = 0;
let checksum = 0;
for (let round = 0; round < rounds; round += 1) {
	// The last round's objects are freed here, before this round's are made.
	held = [];
	for (let index = 0; index < count; index += 1) {
		held.push(made(round, index));
	}
	if (round === rounds - 1) {
		peak = residentKilobytes();
	}
	let total = 0;
	for (const item of held) {
		total = (total + item.value + item.label.length) % 1000000007;
	}
	checksum = (checksum * 31 + total) % 1000000007;
}
held = [];
const dropped = residentKilobytes();

const start = performance.now();
let spins = 0;
while (performance.now() - start < 400) {
	spins += 1;
}
const idle = residentKilobytes();
console.error(`rss peak ${peak} dropped ${dropped} idle ${idle}`);
console.log(`${rounds} rounds of ${count}, checksum ${checksum}, held ${held.length}, spun ${spins > 0}`);
