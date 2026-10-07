Promise.race([new Promise((resolve,reject)=>{globalThis.setTimeout(reject,ms);})]);
export {};
