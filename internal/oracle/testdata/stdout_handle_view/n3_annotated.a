import 'node:process';
const pad = { fd: 7 }; console.log(String(pad.fd));
const handle: { setBlocking?: (value: boolean) => void; readonly fd?: number } | undefined = (process.stdout as { readonly _handle?: { setBlocking(blocking: boolean): void } | undefined })._handle;
console.log(String(handle?.fd));
