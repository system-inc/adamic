// Observation state never lives on the compiler's objects or replaces their identity.
module.exports=function slotRuntime(){
 const objects=new WeakMap(),stats=Object.create(null);
 function row(id){return stats[id]??=( {initializations:0,first_reads:0,before_write:0,after_write:0,nullish_first_reads:0,locations:Object.create(null),writes:0,opaque_calls:0});}
 function init(id){row(id).initializations++;return {id,written:false,read:false};}
 function read(state,value,where){if(state&&!state.read){state.read=true;const r=row(state.id);r.first_reads++;r[state.written?'after_write':'before_write']++;if(value==null)r.nullish_first_reads++;const key=where+'|'+(state.written?'after-write':'before-write')+'|'+(value===undefined?'undefined':value===null?'null':typeof value);r.locations[key]=(r.locations[key]||0)+1;}return value;}
 function written(state){if(state){state.written=true;row(state.id).writes++;}}
 function key(k){return typeof k==='symbol'?k:k!==null&&(typeof k==='object'||typeof k==='function')?Reflect.ownKeys({[k]:0})[0]:String(k);}
 function state(o,k){if(o==null)return undefined;const p=key(k);let owner=o;while(owner!==null&&(typeof owner==='object'||typeof owner==='function')){const found=objects.get(owner)?.get(p);if(found)return found;if(Object.hasOwn(owner,p))return undefined;owner=Object.getPrototypeOf(owner);}return undefined;}
 function mark(o,k,id){let map=objects.get(o);if(!map){map=new Map();objects.set(o,map);}map.set(key(k),init(id));}
 function get(o,k,where){const p=key(k),s=state(o,p);if(!s||s.read)return o[p];const snapshot={id:s.id,written:s.written,read:false};s.read=true;const value=o[p];return read(snapshot,value,where);}
 function put(o,k,v){'use strict';const p=key(k);o[p]=v;written(state(o,p));return v;}
 function initPut(o,k,v,id){'use strict';const p=key(k);o[p]=v;mark(o,p,id);return v;}
 function object(o,entries){for(const [k,id] of entries)mark(o,k,id);return o;}
 function variableInit(token,v,id){token.state=init(id);return v;}
 function variableRead(token,v,where){return read(token.state,v,where);}
 function variablePut(token,v){written(token.state);return v;}
 function call(o,k,where){const method=get(o,k,where);return (...args)=>{if(Array.isArray(o)&&objects.has(o))for(const s of objects.get(o).values())row(s.id).opaque_calls++;return Reflect.apply(method,o,args);};}
 function source(o,where){if(o==null||typeof o!=='object'&&typeof o!=='function')return o;return new Proxy(o,{get(t,k){return get(t,k,where);}});}
 function assign(target,sources,where){const proxy=new Proxy(target,{set(t,k,v){put(t,k,v);return true;}});Object.assign(proxy,...sources.map(s=>source(s,where)));return target;}
 function define(o,k,descriptor,where){const p=key(k);const result=Object.defineProperty(o,p,descriptor);written(state(o,p));return result;}
 function defines(o,descriptors,where){const result=Object.defineProperties(o,descriptors);for(const p of Reflect.ownKeys(descriptors))written(state(o,p));return result;}
 function descriptor(o,k,where){const p=key(k),d=Object.getOwnPropertyDescriptor(o,p);if(d&&Object.hasOwn(d,'value'))read(state(o,p),d.value,where);return d;}
 function descriptors(o,where){const ds=Object.getOwnPropertyDescriptors(o);for(const p of Reflect.ownKeys(ds))if(Object.hasOwn(ds[p],'value'))read(state(o,p),ds[p].value,where);return ds;}
 function json(o,replacer,space,where){if(Array.isArray(replacer))throw Error('unreviewed JSON replacer array in slot probe');return JSON.stringify(o,function(k,v){read(state(this,k),v,where);return typeof replacer==='function'?Reflect.apply(replacer,this,[k,v]):v;},space);}
 return {stats,get,put,initPut,object,variableInit,variableRead,variablePut,call,source,assign,define,defines,descriptor,descriptors,json};
};
