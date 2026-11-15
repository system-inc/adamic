async function run(): Promise<void> {
    const readers: (() => string)[] = [];
    try {
        await Promise.reject(readers);
    } catch (held) {
        if (held instanceof Error) {
            readers.push(() => held.message);
        }
    }
    console.log(String(readers.length));
}
await run();
