// Measurement-only preload. Does not inspect object properties or alter callback/arguments.
const fs = require('node:fs');
const rows = new Map();
const objects = new WeakMap();
let nextId = 0;
function value(v) {
    if (v !== null && (typeof v === 'object' || typeof v === 'function')) {
        if (!objects.has(v)) objects.set(v, ++nextId);
        return {type:typeof v, identity:objects.get(v)};
    }
    if (typeof v === 'number' && (!Number.isFinite(v) || Object.is(v,-0))) return {type:'number',value:String(v),negativeZero:Object.is(v,-0)};
    return {type:typeof v,value:v};
}
function wrap(fn,receiver,site) {
    return function(...args) {
        const tail=args.slice(2).map(value);
        const key=JSON.stringify([site,args.length,tail]);
        const row=rows.get(key);
        if(row)row.count++;
        else rows.set(key,{site,argumentCount:args.length,tail,count:1});
        return Reflect.apply(fn,receiver,args);
    };
}
globalThis.__timerSlotTap=(receiver,name,site)=>wrap(receiver[name],receiver,site);
globalThis.__timerSlotTapFunction=(fn,site)=>wrap(fn,undefined,site);
process.on('exit',()=> {
    if(process.env.TIMER_SLOT_OUTPUT)fs.writeFileSync(`${process.env.TIMER_SLOT_OUTPUT}/${process.pid}.json`,JSON.stringify({pid:process.pid,rows:[...rows.values()]}));
});
