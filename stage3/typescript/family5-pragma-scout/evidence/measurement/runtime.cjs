const assert = require("node:assert/strict");
const path = require("node:path");
const [before, after] = process.argv.slice(2).map(dir => require(path.resolve(dir)));
const inputs = [
    ["/* @jsx", "jsx"],
    ["/* @jsxfrag", "jsxfrag"],
    ["/* @jsximportsource", "jsximportsource"],
    ["/* @jsxruntime", "jsxruntime"],
    ["/// <amd-dependency path='' name='' />", "amd-dependency"],
    ["/// <amd-module name='' />", "amd-module"],
    ["/// <reference path='' />", "reference"],
];
function observe(compiler) {
    return inputs.map(([input, name]) => {
        const file = compiler.createSourceFile("pragma.ts", input, 99);
        const args = file.pragmas.get(name).arguments;
        return {
            input,
            keys: Object.keys(args),
            values: Object.entries(args).map(([key, value]) =>
                [key, value === undefined ? { undefined: true } : value]),
        };
    });
}
const baseline = observe(before);
const trial = observe(after);
assert.deepEqual(trial, baseline);
console.log(JSON.stringify({ baseline, trial }, null, 2));
