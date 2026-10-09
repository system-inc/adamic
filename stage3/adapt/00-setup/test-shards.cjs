"use strict";
// Stable independent-case partition, shared by the adaptation proof helpers.
const assert = require("node:assert/strict");
function select(cases) {
    const ordered = [...cases].sort();
    assert.equal(new Set(ordered).size, ordered.length, "duplicate proof case");
    const value = process.env.ADAMIC_TEST_SHARD || "0/1";
    assert(/^\d+\/\d+$/.test(value), "ADAMIC_TEST_SHARD must be i/n");
    const [index, count] = value.split("/").map(Number);
    assert(Number.isSafeInteger(index) && Number.isSafeInteger(count) && index < count,
        "ADAMIC_TEST_SHARD requires 0 <= i < n");
    if (process.env.ADAMIC_TEST_LIST_SHARDS === "1") {
        console.log(JSON.stringify(ordered.map((name, ordinal) => ({name, shard: `${ordinal % count}/${count}`}))));
        process.exit(0);
    }
    return ordered.filter((name, ordinal) => ordinal % count === index);
}
module.exports = { select };
