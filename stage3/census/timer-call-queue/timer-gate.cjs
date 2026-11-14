'use strict';
const fs = require('node:fs');
const Module = require('node:module');
const path = require('node:path');
const counts = {pid:process.pid, sites:{}, accessors:[], writes:[], transformed:[]};
const define = Object.defineProperty;
const active = new Map();
Object.defineProperty = function(target,key,descriptor) {
    const held=active.get(target);
    if(held && (key==='setTimeout'||key==='clearTimeout')) counts.writes.push({site:held.site,key,operation:'defineProperty'});
    return define(target,key,descriptor);
};
globalThis.__timerGateStart = function(target,site) {
    counts.sites[site]=(counts.sites[site]||0)+1;
    const saved=[];
    for(const key of ['setTimeout','clearTimeout']) {
        let holder=target, descriptor;
        while(holder && !(descriptor=Object.getOwnPropertyDescriptor(holder,key))) holder=Object.getPrototypeOf(holder);
        if(descriptor && ('get' in descriptor || 'set' in descriptor)) counts.accessors.push({site,key});
        if(descriptor && !('value' in descriptor)) continue;
        const own=Object.getOwnPropertyDescriptor(target,key);
        if(own && !own.configurable) throw Error('gate cannot instrument nonconfigurable '+site+':'+key);
        let value=descriptor?.value;
        define(target,key,{configurable:true,enumerable:own?.enumerable??false,get(){return value;},set(next){counts.writes.push({site,key,operation:'assignment'});value=next;}});
        saved.push({key,own,getValue:()=>value,initial:descriptor?.value});
    }
    active.set(target,{site});
    return () => {
        active.delete(target);
        for(const item of saved) {
            if(item.own) define(target,item.key,{...item.own,value:item.getValue()});
            else { delete target[item.key]; if(item.getValue()!==item.initial) define(target,item.key,{value:item.getValue(),writable:true,enumerable:true,configurable:true}); }
        }
    };
};
const compile=Module.prototype._compile;
Module.prototype._compile=function(text,filename) {
    if(!filename.startsWith("/tmp/adamic-timer-queue-baseline/built/local/")) return compile.call(this,text,filename);
    if(/\b(?:const|let|var) \{ setTimeout(?:: setTimeout\w*)? \} = timerHost;/.test(text) && text.includes('.call(timerHost,')) {
        let n=0;
        text=text.replace(/\b(?:const|let|var) \{ setTimeout(?:: setTimeout\w*)? \} = timerHost;/g,(original)=>`const __timerGateEnd = globalThis.__timerGateStart(timerHost, ${JSON.stringify(filename)} + ':' + ${++n});\n${original}`);
        let returns=0;
        text=text.replace(/return (setTimeout\w*)\.call\(timerHost,/g,(_,alias)=>{returns++;return '__timerGateEnd(); return '+alias+'.call(timerHost,';});
        if(n!==returns) throw Error('gate transform mismatch '+filename+':'+n+'/'+returns);
        counts.transformed.push({filename,sites:n});
        return compile.call(this,text,filename);
    }
    if(text.includes('function createDynamicPriorityPollingWatchFile(')) throw Error('gate: compiler bundle had no matching timer blocks '+filename);
    return compile.call(this,text,filename);
};
process.on('exit',()=>{
    if(!process.env.TIMER_GATE_OUTPUT) return;
    fs.mkdirSync(process.env.TIMER_GATE_OUTPUT,{recursive:true});
    fs.writeFileSync(path.join(process.env.TIMER_GATE_OUTPUT,process.pid+'.json'),JSON.stringify(counts,null,2));
});
