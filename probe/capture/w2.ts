import 'node:process';
Error.captureStackTrace = (target: object): void => { console.log("replaced"); };
const target = new Error("t");
Error.captureStackTrace(target);
