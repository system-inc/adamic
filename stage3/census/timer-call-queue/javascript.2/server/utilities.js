import { getBaseFileName } from "./_namespaces/ts.js";
import { LogLevel, } from "./_namespaces/ts.server.js";
/** @internal */
export class ThrottledOperations {
    host;
    pendingTimeouts = new Map();
    logger;
    constructor(host, logger) {
        this.host = host;
        this.logger = logger.hasLevel(LogLevel.verbose) ? logger : undefined;
    }
    /**
     * Wait `number` milliseconds and then invoke `cb`.  If, while waiting, schedule
     * is called again with the same `operationId`, cancel this operation in favor
     * of the new one.  (Note that the amount of time the canceled operation had been
     * waiting does not affect the amount of time that the new operation waits.)
     */
    schedule(operationId, delay, cb) {
        const pendingTimeout = this.pendingTimeouts.get(operationId);
        if (pendingTimeout) {
            // another operation was already scheduled for this id - cancel it
            this.host.clearTimeout(pendingTimeout);
        }
        // schedule new operation, pass arguments
        this.pendingTimeouts.set(operationId, (() => {
            const timerHost = this.host;
            const callback = ThrottledOperations.run;
            const timeoutDelay = delay;
            const t1 = operationId;
            const t2 = this;
            const t3 = cb;
            return timerHost.setTimeout(() => callback(t1, t2, t3), timeoutDelay, t1);
        })());
        if (this.logger) {
            this.logger.info(`Scheduled: ${operationId}${pendingTimeout ? ", Cancelled earlier one" : ""}`);
        }
    }
    cancel(operationId) {
        const pendingTimeout = this.pendingTimeouts.get(operationId);
        if (!pendingTimeout)
            return false;
        this.host.clearTimeout(pendingTimeout);
        return this.pendingTimeouts.delete(operationId);
    }
    static run(operationId, self, cb) {
        self.pendingTimeouts.delete(operationId);
        if (self.logger) {
            self.logger.info(`Running: ${operationId}`);
        }
        cb();
    }
}
/** @internal */
export class GcTimer {
    host;
    delay;
    logger;
    timerId;
    constructor(host, delay, logger) {
        this.host = host;
        this.delay = delay;
        this.logger = logger;
    }
    scheduleCollect() {
        if (!this.host.gc || this.timerId !== undefined) {
            // no global.gc or collection was already scheduled - skip this request
            return;
        }
        this.timerId = (() => {
            const timerHost = this.host;
            const callback = GcTimer.run;
            const timeoutDelay = this.delay;
            const t1 = this;
            return timerHost.setTimeout(() => callback(t1), timeoutDelay, t1);
        })();
    }
    static run(self) {
        self.timerId = undefined;
        const log = self.logger.hasLevel(LogLevel.requestTime);
        const before = log && self.host.getMemoryUsage(); // TODO: GH#18217
        self.host.gc(); // TODO: GH#18217
        if (log) {
            const after = self.host.getMemoryUsage(); // TODO: GH#18217
            self.logger.perftrc(`GC::before ${before}, after ${after}`);
        }
    }
}
/** @internal */
export function getBaseConfigFileName(configFilePath) {
    const base = getBaseFileName(configFilePath);
    return base === "tsconfig.json" || base === "jsconfig.json" ? base : undefined;
}
