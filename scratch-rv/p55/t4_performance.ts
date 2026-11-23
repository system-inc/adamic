interface Clock {
    now(): number;
}
const clock: Clock = performance;
console.log(clock.now ? "has now" : "no now");
