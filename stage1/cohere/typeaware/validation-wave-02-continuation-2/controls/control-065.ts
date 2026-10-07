const timeout=new Promise((resolve,reject)=>{const timer=setTimeout(reject,ms);});Promise.race([work(),timeout]);
export {};
