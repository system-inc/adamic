function f(setTimeout:typeof globalThis.setTimeout){Promise.race([new Promise((resolve,reject)=>{setTimeout(reject,ms);})]);}
export {};
