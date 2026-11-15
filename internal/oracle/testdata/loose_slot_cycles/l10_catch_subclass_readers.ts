class Carrier extends Error {
    readers: (() => string)[] = [];
}
async function run(): Promise<void> {
    const readers: (() => string)[] = [];
    try {
        const carrier = new Carrier("carried");
        carrier.readers = readers;
        throw carrier;
    } catch (held) {
        if (held instanceof Error) {
            readers.push(() => held.message);
        }
    }
    await Promise.resolve();
    console.log(String(readers.length));
}
await run();
