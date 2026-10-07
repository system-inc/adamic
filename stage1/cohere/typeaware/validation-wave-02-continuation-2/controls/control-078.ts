Promise.race([new Promise((resolve,reject)=>{onEvent(()=>{setTimeout(reject,ms);});})]);
export {};
