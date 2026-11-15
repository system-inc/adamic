async function run(): Promise<void> {
    const readers: (() => string)[] = [];
    try {
        const error = new Error("carried");
        error.cause = readers;
        throw error;
    } catch (held) {
        if (held instanceof Error) {
            readers.push(() => held.message);
        }
    }
    await Promise.resolve();
    console.log(String(readers.length));
}
await run();
