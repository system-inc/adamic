import 'node:process';
/*
 * process.recordreplay, which tsc's sys.ts reads (#b7phwab): a member Node's process doesn't have, so it
 * reads as undefined on every target (Adamic on #b5w3ycg, Nov 7 2026). The prelude declares it optional
 * on NodeJS.Process, and lowering folds the read to undefined without reading a process slot. Node
 * prints false, then undefined.
 */
console.log(String(!!process.recordreplay));
console.log(typeof process.recordreplay);
