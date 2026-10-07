Promise.race<unknown>([new Promise<unknown>((resolve,reject)=>(setTimeout(reject,ms)))]);
export {};
