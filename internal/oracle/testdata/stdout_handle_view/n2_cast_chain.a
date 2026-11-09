import 'node:process';
const pad = { fd: 7 }; console.log(String(pad.fd));
const handle = (process.stdout as any)?._handle as { setBlocking?: (value: boolean) => void; readonly fd?: number };
console.log(String(handle?.fd));
