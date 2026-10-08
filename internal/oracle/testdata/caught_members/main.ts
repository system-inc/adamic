interface ImportResult { error?: { stack?: string; message?: string }; }
function imported(value: unknown): ImportResult {
    try { throw value; } catch (e) { const error = e; return { error }; }
}
function text(value: string): string { return value; }
function inspect(value: unknown): void {
    try { throw value; } catch (e) {
        console.log(`${e instanceof Error} ${typeof e} ${e === value} ${e === null}`);
        try {
            console.log(`${typeof e.message} ${e.message === undefined} ${e.message === 'm'} ${e.message === 'e'} ${e.code === 'ENOSPC'}`);
        } catch (failure) {
            if (failure instanceof Error) { console.log(`${failure instanceof Error} ${text(failure.name)}: ${text(failure.message)}`); }
        }
    }
    const result = imported(value);
    console.log(`${typeof result.error} ${result.error === value} ${result.error instanceof Error}`);
    if (result.error !== undefined) {
        try { console.log(`${typeof result.error.message}`); }
        catch (failure) { console.log(`${failure instanceof Error}`); }
    }
}
inspect('x');
inspect({});
inspect({ message: 'm', code: 'ENOSPC' });
inspect(new Error('e'));
inspect(undefined);
inspect(null);
const nothing: null = null;
inspect(nothing);
inspect(7);
inspect({ message: 7 });
inspect({ message: true });
inspect({ message: null });
inspect({ message: nothing });
console.log('after');

class NullMessage {
    readonly message: null;
    constructor() { this.message = null; }
    keep(): string { return 'kept'; }
}
inspect(new NullMessage());
