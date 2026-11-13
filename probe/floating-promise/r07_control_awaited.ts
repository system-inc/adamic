async function fails(): Promise<void> { await Promise.resolve(); throw new Error("rejected"); }
async function main(): Promise<void> {
	try { await fails(); } catch (error) { console.log("caught"); }
}
await main();
console.log("after");
