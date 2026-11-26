type Value = boolean | null | undefined;
const entries = new Map<string, Value>([["null", null], ["undefined", undefined], ["yes", true]]);
for (const [key, value] of entries) {
    if (value === null) console.log(key + ":null");
    else if (value === undefined) console.log(key + ":undefined");
    else console.log(key + (value ? ":true" : ":false"));
}
