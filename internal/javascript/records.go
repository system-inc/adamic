package javascript

// Matches runtime/record.c: own data keys only; an inherited Object.prototype
// member cannot silently become an absent entry.
const recordRuntime = `const adamicRecordGet = (record, key) => {
 if (Object.hasOwn(record, key)) return record[key];
 if (Object.hasOwn(Object.prototype, key)) panic("record member '" + key + "' is missing; records hold own keys only");
 return undefined;
};
`
