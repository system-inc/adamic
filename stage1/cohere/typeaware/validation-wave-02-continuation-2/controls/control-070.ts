let timer:NodeJS.Timeout|undefined;Promise.race([new Promise((resolve,reject)=>{timer=setTimeout(reject,ms);})]);
export {};
