import * as hooks from 'node:perf_hooks';
interface Clock {
    now(): number;
}
const clock: Clock = hooks.performance;
if (clock.now) {
    console.log(String(clock.now() >= 0));
} else {
    console.log("no now");
}
