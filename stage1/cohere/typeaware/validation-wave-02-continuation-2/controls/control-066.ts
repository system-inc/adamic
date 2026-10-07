let timeout=new Promise((resolve,reject)=>{setTimeout(reject,ms);});Promise.race([work(),timeout]);
export {};
