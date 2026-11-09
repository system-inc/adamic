import 'node:process';
const handle = (process.stdout as any)._handle;
console.log(String(handle === undefined));
