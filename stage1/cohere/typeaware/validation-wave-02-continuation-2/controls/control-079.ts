Promise.race([new Promise(function(resolve,reject){setTimeout(reject,ms);setTimeout(reject,ms);})]);
export {};
