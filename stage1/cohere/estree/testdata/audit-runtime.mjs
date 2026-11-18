// Successful operations use the ordinary Node oracle. Audit refusals are catchable per file by this
// harness (corpus.mjs), never by the program's own catch: a panic here carries an identity that isPanic
// answers, so an emitted catch rethrows it as the Workers runtime's does (docs/rulings.md, a panic is
// never caught). These exports take precedence over the oracle's panic and isPanic.
export * from '../../../../oracle/adamic.mjs';
class AuditPanic extends Error {}
export function panic(message) { throw new AuditPanic(message); }
export function isPanic(value) { return value instanceof AuditPanic; }
