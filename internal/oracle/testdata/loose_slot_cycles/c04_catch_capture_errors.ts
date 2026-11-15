async function run(): Promise<void> {
    const readers: (() => string)[] = [];
    let i = 0;
    while (i < 3) {
        try {
            throw new Error(`item${i}`);
        } catch (held) {
            if (held instanceof Error) {
                readers.push(() => held.message);
            }
        }
        await Promise.resolve();
        i++;
    }
    for (const read of readers) {
        console.log(read());
    }
}
await run();
