import 'node:process';
const pad = { fd: 7 }; console.log(String(pad.fd));
const handle = (process.stdout as { readonly _handle?: { setBlocking(blocking: boolean): void; readonly fd?: number } | undefined })._handle;
console.log(String(handle?.fd));
