Promise.race([new Promise((resolve,reject)=>{setTimeout(reject,ms).unref();})]);
export {};
