Promise.race([new Promise((resolve,reject)=>{void setTimeout(reject,ms);})]);
export {};
