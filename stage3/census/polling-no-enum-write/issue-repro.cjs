process.env.TSC_WATCHFILE = 'DynamicPriorityPolling';
process.env.TSC_WATCH_POLLINGINTERVAL_LOW = '17';
process.env.TSC_WATCH_POLLINGINTERVAL_MEDIUM = '41';
process.env.TSC_WATCH_POLLINGINTERVAL_HIGH = '89';
const scheduled = [];
global.setTimeout = (callback, interval) => {
    scheduled.push(interval);
    return { unref() {} };
};
global.clearTimeout = () => {};
const ts = require('typescript');
const watchers = [];
console.log(ts.version);
console.log(ts.PollingInterval.Low, ts.PollingInterval.Medium, ts.PollingInterval.High);
for (const interval of [250, 500, 2000, 17, 41, 89]) {
    try {
        watchers.push(ts.sys.watchFile(__filename, () => {}, interval));
        console.log(interval, 'ok');
    } catch (error) {
        console.log(interval, error.name + ': ' + error.message);
    }
}
console.log(scheduled);
watchers.forEach(watcher => watcher.close());
