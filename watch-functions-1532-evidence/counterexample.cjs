'use strict';
const assert = require('node:assert/strict');
const PollingInterval = { Low: 250, Medium: 500, High: 2000 };
const lowPollingIntervalQueue = { files: [] }, mediumPollingIntervalQueue = { files: [] }, highPollingIntervalQueue = { files: [] };
function scheduleNextPollIfNotAlreadyScheduled() {}
    function pollingIntervalQueue(pollingInterval) {
        switch (pollingInterval) {
            case PollingInterval.Low:
                return lowPollingIntervalQueue;
            case PollingInterval.Medium:
                return mediumPollingIntervalQueue;
            case PollingInterval.High:
                return highPollingIntervalQueue;
        }
    }

    function addToPollingIntervalQueue(file, pollingInterval) {
        pollingIntervalQueue(pollingInterval).files.push(file);
        scheduleNextPollIfNotAlreadyScheduled(pollingInterval);
    }


for (const value of [undefined, 123]) { assert.throws(() => addToPollingIntervalQueue({}, value), TypeError); console.log(String(value) + ': TypeError confirmed'); }
