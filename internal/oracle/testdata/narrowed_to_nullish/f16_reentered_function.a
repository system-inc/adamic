interface Box { x: number | undefined; }
class Runner { again(box: Box): void { step(box, this, 1); } }
function step(box: Box, runner: Runner, depth: number): void {
	if (depth > 0) { box.x = 4; return; }
	box.x = undefined;
	runner.again(box);
	console.log(String(box.x));
}
step({ x: 1 }, new Runner(), 0);
