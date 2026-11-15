import 'node:process';
const handle = (process.stdout as any)._handle as {};
const view = handle as { writeQueueSize?: number }; console.log(String(view.writeQueueSize));
