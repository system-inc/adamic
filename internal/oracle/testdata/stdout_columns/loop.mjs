// Node-side host-turn twin of loop.c. The async read really yields to the event
// loop; await Promise.resolve() alone must not be confused with a SIGWINCH turn.
import process from 'node:process';
import fs from 'node:fs';
console.error('initial', process.stdout.columns);
if (process.argv[2] === 'sync') {
    // Node makes its inherited pipe nonblocking. Open a fresh blocking
    // description for this deliberate synchronous barrier.
    const input = fs.openSync('/dev/stdin', 'r');
    fs.readSync(input, Buffer.alloc(1), 0, 1, null);
    fs.closeSync(input);
} else {
    await new Promise((resolve) => process.stdin.once('data', () => resolve()));
    process.stdin.pause();
}
console.error('later', process.stdout.columns);
