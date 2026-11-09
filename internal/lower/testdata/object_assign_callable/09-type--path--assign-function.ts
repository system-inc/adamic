// Logging-only adaptation for Adamic's console.log(message: string).
// Self-contained excerpts from microsoft/TypeScript tag v6.0.3.
// Only required Type/Symbol fields are retained. Dependency leaves are opaque:
// their contents do not affect rejection of a property-free function.
export {};
interface MapLike<T> { [index: string]: T; }
type SymbolId = number;
type __String = (string & { __escapedIdentifier: void; }) | (void & { __escapedIdentifier: void; }) | InternalSymbolName;
// Cleaned for #k8aak04: the ambient enum is used only as a type here, and an ambient enum is refused
// ahead of Object.assign, so it stands as its value type.
type InternalSymbolName = string;
// Cleaned for #k8aak04: the ambient enum is used only as a type here, and an ambient enum is refused
// ahead of Object.assign, so it stands as its value type.
type SymbolFlags = number;
interface Symbol {
 flags: SymbolFlags;
 escapedName: __String;
 id: SymbolId;
 mergeId: number;
 constEnumOnlyModule: boolean | undefined;
}
// Cleaned for #k8aak04: the ambient enum is used only as a type here, and an ambient enum is refused
// ahead of Object.assign, so it stands as its value type.
type TypeFlags = number;
type TypeId = number;
type TypeChecker = object;
interface Type {
 flags: TypeFlags;
 id: TypeId;
 checker: TypeChecker;
 symbol: Symbol;
}


// The value whose static receiver type is being tested is the record itself.
type Receiver = (path: string) => string;
type Holding = MapLike<Type>;
const h: Holding = {};
const source: Receiver = path => path;
const assigned = Object.assign(h, source);
const receiver: Receiver = assigned;
console.log(Object.is(receiver, h) ? "same record: true" : "same record: false");
console.log("typeof receiver: " + typeof receiver);
try {
 receiver("called");
 console.log("call returned");
} catch (error) {
 if (!(error instanceof TypeError)) throw error;
 console.log("call threw: TypeError");
}
