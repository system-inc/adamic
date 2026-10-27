// Reduced from TypeScript 6.0.3 utilities.ts:8040, assertions/13_flag_downcast.a.
type CheckFlags = number;
const SymbolFlags = { Transient: 33554432 } as const;
interface Symbol { readonly flags: number; }
interface TransientSymbol extends Symbol { readonly links: { readonly checkFlags: CheckFlags; }; }
function getCheckFlags(symbol: Symbol): CheckFlags {
    return symbol.flags & SymbolFlags.Transient ? (symbol as TransientSymbol).links.checkFlags : 0;
}
const raw = { flags: SymbolFlags.Transient, links: { checkFlags: true } };
const malformed: Symbol = raw;
console.log(String(getCheckFlags(malformed)));
