const hashTable = { "#": 1 };
const TypeFacts = { TypeofNEString: 0, TypeofNENumber: 1, TypeofNEBigInt: 2, TypeofNEBoolean: 3, TypeofNESymbol: 4, NEUndefined: 5, TypeofNEObject: 6, TypeofNEFunction: 7 };
const typeofNEFacts: ReadonlyMap<string, number> = new Map(Object.entries({
    string: TypeFacts.TypeofNEString,
    number: TypeFacts.TypeofNENumber,
    bigint: TypeFacts.TypeofNEBigInt,
    boolean: TypeFacts.TypeofNEBoolean,
    symbol: TypeFacts.TypeofNESymbol,
    undefined: TypeFacts.NEUndefined,
    object: TypeFacts.TypeofNEObject,
    function: TypeFacts.TypeofNEFunction,
}));
console.log(String(typeofNEFacts.get("number")));
