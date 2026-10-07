const Promise={race(items:unknown[]){return items;}};Promise.race([work(),new globalThis.Promise((resolve,reject)=>{setTimeout(reject,ms);})]);
export {};
