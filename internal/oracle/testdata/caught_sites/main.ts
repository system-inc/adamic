// D108, D128, D132: a message passed into a definite string use.
function diagnostic(message: string): void { console.log(`diagnostic ${message}`); }
function onError(message: string): void { console.log(`onError ${message}`); }
function report(value: unknown): void {
    try { throw value; } catch (e) { diagnostic(e.message); }
    try { throw value; } catch (e) { onError(e.message); }
    try { throw value; } catch (e) { onError(e.message); }
}
// D152: ordinary code comparison; a missing code remains undefined.
function watcherLimit(value: unknown): boolean {
    try { throw value; } catch (e) { return e.code === 'ENOSPC'; }
}
// D153: returned Error-like record does not change what was thrown.
interface ModuleImportResult { error?: { stack?: string; message?: string }; }
function requireResult(value: unknown): ModuleImportResult {
    try { throw value; } catch (e) { return { error: e }; }
}
function maybeMessage(value: unknown): string | undefined {
    try { throw value; } catch (e) { return e.message; }
}
report(new Error('e'));
report({ message: 'm' });
console.log(`${watcherLimit({ code: 'ENOSPC' })} ${watcherLimit('x')}`);
console.log(`${typeof requireResult('x').error} ${typeof requireResult(3).error}`);
console.log(`${maybeMessage('x') === undefined} ${maybeMessage(new Error('e')) === 'e'}`);

try {
    try { maybeMessage(undefined); } finally { console.log('cleanup finally'); }
} catch (e) { if (e instanceof Error) { console.log(`${e instanceof Error} ${e.name === 'TypeError'}`); } }
console.log('after');

try { throw { message: 'before' }; } catch (e) {
    const record: { message: string } = { message: e.message };
    record.message = 'after';
    console.log(record.message);
}
