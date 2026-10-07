Promise.race([new Promise((resolve,reject)=>{return setTimeout(reject,ms);})]);
export {};
