// The same 1,000 tasks as parallel.a (#sp4m5ex), written the way Kirk asked: an async task each,
// awaited together through Promise.all. On Node every task still runs on the one JavaScript thread,
// since async is about waiting, not cores. Adamic doesn't build Promise.all yet; demo.sh shows its stop.
function isPrime(n: number): boolean {
	if (n < 2) {
		return false;
	}
	if (n % 2 === 0) {
		return n === 2;
	}
	for (let divisor = 3; divisor * divisor <= n; divisor += 2) {
		if (n % divisor === 0) {
			return false;
		}
	}
	return true;
}

async function primesInTask(task: number): Promise<number> {
	const start = task * 10000;
	let count = 0;
	for (let n = start; n < start + 10000; n += 1) {
		if (isPrime(n)) {
			count += 1;
		}
	}
	return count;
}

const tasks: number[] = [];
for (let task = 0; task < 1000; task += 1) {
	tasks.push(task);
}
const counts = await Promise.all(tasks.map((task) => primesInTask(task)));
let total = 0;
let task = 0;
for (const count of counts) {
	console.log(`task ${task}: ${count} primes in [${task * 10000}, ${(task + 1) * 10000})`);
	total += count;
	task += 1;
}
console.log(`total: ${total} primes below 10,000,000, from ${task} tasks`);
