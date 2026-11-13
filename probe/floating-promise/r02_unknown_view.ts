async function fails(): Promise<void> { await Promise.resolve(); throw new Error("rejected"); }
const run: () => unknown = fails;
run();
console.log("after");
