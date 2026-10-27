// From TypeScript 6.0.3, src/compiler/utilities.ts:8040
// Census reason: type assertion sites
// Support types and driver are reduced scaffolding; the upstream function is unchanged.

type CheckFlags = number;
const SymbolFlags = { Transient: 33554432 } as const;
interface Symbol { readonly flags: number; }
interface TransientSymbol extends Symbol { readonly links: { readonly checkFlags: CheckFlags; }; }

export function getCheckFlags(symbol: Symbol): CheckFlags {
    return symbol.flags & SymbolFlags.Transient ? (symbol as TransientSymbol).links.checkFlags : 0;
}

const transient: TransientSymbol = { flags: SymbolFlags.Transient, links: { checkFlags: 8 } };
console.log(String(getCheckFlags(transient)));
console.log(String(getCheckFlags({ flags: 0 })));
