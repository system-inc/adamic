async function fails(): Promise<void> { await Promise.resolve(); throw new Error("rejected"); }
[0].map(fails);
console.log("after");
