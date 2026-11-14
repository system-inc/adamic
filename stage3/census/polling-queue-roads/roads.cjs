'use strict';
const fs=require('node:fs'),assert=require('node:assert/strict');
const pins=require('../../adapt/99-correctness-timer-closures/owners.json');
const original=pins.find(p=>p.owner==='createDynamicPriorityPollingWatchFile').after;
const factory=`        const queue = [] as WatchedFileWithUnchangedPolls[] as PollingIntervalQueue;
        queue.pollingInterval = pollingInterval;
        queue.pollIndex = 0;
        queue.pollScheduled = false;
        return queue;`;
function replace(s,a,b){assert.equal(s.split(a).length,2,a);return s.replace(a,b)}
const roads={original};
let a=replace(original,'interface PollingIntervalQueue extends Array<WatchedFileWithUnchangedPolls> {','interface PollingIntervalQueue {\n        files: WatchedFileWithUnchangedPolls[];');
a=replace(a,factory,'        return { files: [] as WatchedFileWithUnchangedPolls[], pollingInterval, pollIndex: 0, pollScheduled: false };');
a=replace(a,'pollQueue(queue, queue.pollingInterval','pollQueue(queue.files, queue.pollingInterval');
a=replace(a,'if (queue.length)','if (queue.files.length)');
a=replace(a,'pollingIntervalQueue(pollingInterval).push(file)','pollingIntervalQueue(pollingInterval).files.push(file)');
roads.a=a;
roads.b=replace(original,factory,'        const queue: PollingIntervalQueue = Object.assign([] as WatchedFileWithUnchangedPolls[], { pollingInterval, pollIndex: 0, pollScheduled: false });\n        return queue;');
let c=replace(original,`    interface PollingIntervalQueue extends Array<WatchedFileWithUnchangedPolls> {
        pollingInterval: PollingInterval;
        pollIndex: number;
        pollScheduled: unknown;
    }`,`    class PollingIntervalQueue extends Array<WatchedFileWithUnchangedPolls> {
        pollingInterval: PollingInterval;
        pollIndex = 0;
        pollScheduled: unknown = false;
        constructor(pollingInterval: PollingInterval) {
            super();
            this.pollingInterval = pollingInterval;
        }
    }`);
roads.c=replace(c,factory,'        return new PollingIntervalQueue(pollingInterval);');
module.exports={roads,replace};
if(require.main===module){for(const [k,v]of Object.entries(roads))fs.writeFileSync(`${__dirname}/evidence/owner.${k}.ts`,v+'\n');}
