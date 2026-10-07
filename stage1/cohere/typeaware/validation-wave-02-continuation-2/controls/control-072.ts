Promise.race([new Promise((resolve,reject)=>{const timer=setTimeout(reject,ms);const obj={timer};})]);
export {};
