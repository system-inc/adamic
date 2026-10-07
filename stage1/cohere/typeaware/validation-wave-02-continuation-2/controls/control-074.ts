Promise.race([new Promise((resolve,reject)=>{let timer=setTimeout(reject,ms);timer=undefined;})]);
export {};
