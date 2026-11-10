import 'node:process';
(Error as { captureStackTrace: (t: object) => void }).captureStackTrace = (t: object): void => { console.log("replaced"); };
Error.captureStackTrace(new Error("t"));
