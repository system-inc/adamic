async function fails(): Promise<void> { await Promise.resolve(); throw new Error("rejected"); }
function call<T>(f: () => T): void { f(); }
call(fails);
console.log("after");
