interface TimerHost {
    setTimeout: (callback: (label?: unknown) => void, ms: number, label?: unknown) => number;
}
const host: TimerHost = {
    setTimeout: (callback: (label?: unknown) => void, ms: number, label?: unknown): number => {
        console.log(`${ms}`);
        callback(label);
        return 7;
    },
};
function forward(label: string, identity: number): void {
    console.log(label);
    console.log(`${identity}`);
}
const handle = (() => {
    const timerHost = host;
    const callback = forward;
    const timeoutDelay = 250;
    const t1 = "label";
    const t2 = 42;
    return timerHost.setTimeout(() => callback(t1, t2), timeoutDelay, t1);
})();
console.log(`${handle}`);
