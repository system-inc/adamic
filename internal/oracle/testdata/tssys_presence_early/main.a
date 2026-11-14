/*
 * getNodeSystem's literal read early (#p2n17c4 T1, road 7): a function holding a nested declaration makes
 * its variables' cells before its body runs, and the reach takes that early cell for no source, so a
 * presence test through it is judged by the literal its declaration assigns. That rests on the read
 * before the assignment stopping. peek is hoisted, so getNodeSystem(true) calls it before nodeSystem's
 * declaration has run, and its presence test must throw Node's ReferenceError, never test a slot no
 * literal fills. getNodeSystem(false) calls it after, and the test reads the literal's method.
 * Node prints ReferenceError|Cannot access 'nodeSystem' before initialization, late: writeOutputIsTTY
 * present, done.
 */
interface System {
	write(s: string): void;
	writeOutputIsTTY?(): boolean;
}

function getNodeSystem(early: boolean): System {
	if (early) {
		peek('early');
	}
	const nodeSystem: System = {
		write(s: string): void {
			console.log(s);
		},
		writeOutputIsTTY() {
			return false;
		},
	};
	peek('late');
	return nodeSystem;

	function peek(when: string): void {
		console.log(`${when}: writeOutputIsTTY ${nodeSystem.writeOutputIsTTY ? 'present' : 'absent'}`);
	}
}

try {
	getNodeSystem(true);
} catch (error) {
	if (error instanceof Error) {
		console.log(`${error.name}|${error.message}`);
	}
}
getNodeSystem(false).write('done');
