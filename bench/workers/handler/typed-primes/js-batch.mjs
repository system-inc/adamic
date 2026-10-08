const {primes} = await import(process.argv[2]);
const limit = Number(process.argv[3]);
for(let i=0;i<2;i++) primes(limit);
let count=0;
const start=performance.now();
for(let i=0;i<20;i++) count+=primes(limit).count;
console.log(JSON.stringify({ms:(performance.now()-start)/20,rssKiB:process.resourceUsage().maxRSS,count}));
