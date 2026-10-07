Promise.race([work(),new Promise((resolve,reject)=>{setTimeout(reject,ms);})]);
export {};
