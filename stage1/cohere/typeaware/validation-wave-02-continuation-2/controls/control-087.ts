Promise.race([new Promise((resolve,reject)=>{class Nested{static{setTimeout(reject,ms);}}})]);
export {};
