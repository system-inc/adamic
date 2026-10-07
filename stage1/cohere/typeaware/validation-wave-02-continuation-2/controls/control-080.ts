const obj={timer:undefined as NodeJS.Timeout|undefined};Promise.race([new Promise((resolve,reject)=>{obj.timer=setTimeout(reject,ms);})]);
export {};
