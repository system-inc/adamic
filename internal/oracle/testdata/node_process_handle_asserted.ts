// stdout's handle's setBlocking called off a let-held handle with a non-null assertion on the callee,
// handle.setBlocking!(true), which JavaScript runs as handle.setBlocking(true): the call keeps its
// object, so it blocks, as Node's does (#0bxc4jh). TypeScript, since .a refuses the assertion.
import 'node:process';
let handle = (process.stdout as any)._handle as { setBlocking?: (value: boolean) => void } | undefined;
if (handle === undefined) {
	console.log('no handle');
} else {
	handle.setBlocking!(true);
	console.log('blocked');
}
