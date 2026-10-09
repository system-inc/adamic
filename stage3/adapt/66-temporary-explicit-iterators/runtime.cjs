'use strict';
module.exports = `// Temporary generator fallback. Remove with adaptation 66.
function temporaryExplicitIterator<T>(
    pull: (value: unknown) => IteratorResult<T, undefined>,
    close: () => void = () => {},
    abrupt?: (completion: { kind: "return"; value: undefined } | { kind: "throw"; value: unknown }) => IteratorResult<T, undefined>,
    settled?: () => boolean,
): Generator<T, undefined, unknown> {
    let started = false;
    let finished = false;
    let running = false;
    return {
        next(value?: unknown): IteratorResult<T, undefined> {
            if (running) throw new TypeError("Generator is already running");
            if (finished) return { done: true, value: undefined };
            const input = started ? value : undefined;
            started = true;
            running = true;
            try {
                const result = pull(input);
                finished = settled ? settled() : !!result.done;
                return result;
            }
            catch (error) { finished = true; throw error; }
            finally { running = false; }
        },
        return(value: undefined): IteratorResult<T, undefined> {
            if (running) throw new TypeError("Generator is already running");
            if (!started || finished) { finished = true; return { done: true, value }; }
            running = true;
            try {
                if (abrupt) {
                    const result = abrupt({ kind: "return", value });
                    finished = settled ? settled() : !!result.done;
                    return result;
                }
                close();
                finished = true;
                return { done: true, value };
            }
            catch (error) { finished = true; throw error; }
            finally { running = false; }
        },
        throw(error: unknown): IteratorResult<T, undefined> {
            if (running) throw new TypeError("Generator is already running");
            if (!started || finished) { finished = true; throw error; }
            running = true;
            try {
                if (abrupt) {
                    const result = abrupt({ kind: "throw", value: error });
                    finished = settled ? settled() : !!result.done;
                    return result;
                }
                try { close(); } catch { /* The original throw wins over IteratorClose. */ }
                throw error;
            }
            catch (failure) { finished = true; throw failure; }
            finally { running = false; }
        },
        [Symbol.iterator]() { return this; },
        [Symbol.dispose]() { this.return(undefined); },
    };
}

function temporaryIteratorClose<T>(iterator: Iterator<T, unknown, unknown> | undefined): void {
    const close = iterator?.return;
    if (close != null) {
        const result = close.call(iterator);
        if (typeof result !== "object" || result === null) throw new TypeError("Iterator result is not an object");
    }
}

function temporaryIteratorResult<T, R>(result: IteratorResult<T, R>): IteratorResult<T, R> {
    if (typeof result !== "object" || result === null) throw new TypeError("Iterator result is not an object");
    return result;
}
`;
