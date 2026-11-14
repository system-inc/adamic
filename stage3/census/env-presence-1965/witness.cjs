"use strict";
const assert = require("node:assert/strict");
function run(sys, changed) {
    const calls=[];
    function setCustomPollingValues(system) { calls.push("setCustomPollingValues(sys)"); if (!system.getEnvironmentVariable) return; calls.push(system.getEnvironmentVariable("TSC_WATCH_POLLINGINTERVAL_LOW")); }
    try {
        if (changed ? sys : sys && sys.getEnvironmentVariable) {
            setCustomPollingValues(sys);
            calls.push(/^development$/i.test(sys.getEnvironmentVariable("NODE_ENV")) ? "AssertionLevel.Normal" : "AssertionLevel.None");
        }
        return {calls};
    } catch(e) { return {calls,error:e.name,message:e.message}; }
}
for (const [label,sys] of [["present",{getEnvironmentVariable(name){assert.equal(this,sys);return name==="NODE_ENV"?"development":"";}}],["undefined",undefined],["hypothetical missing member",{}]]) {
    const before=run(sys,false),after=run(sys,true);
    if(label!=="hypothetical missing member") assert.deepEqual(before,after);
    else {assert.equal(before.calls.length,0);assert.equal(after.error,"TypeError");}
    console.log(JSON.stringify({label,before,after}));
}
