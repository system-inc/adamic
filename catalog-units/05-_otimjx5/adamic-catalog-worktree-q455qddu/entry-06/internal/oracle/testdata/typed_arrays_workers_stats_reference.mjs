// Independent Node witness: original functions from platforms 900409b73d36ea34bb26ec4366c09952839ff897.
import { stripTypeScriptTypes } from 'node:module';
const original = "interface StatsInterface {\n    readonly count: number;\n    readonly sum: number;\n    readonly minimum: number;\n    readonly maximum: number;\n}\n\nfunction stats(values: readonly number[]): StatsInterface {\n    let sum = 0;\n    let minimum = Infinity;\n    let maximum = -Infinity;\n    for(const value of values) {\n        sum += value;\n        minimum = Math.min(minimum, value);\n        maximum = Math.max(maximum, value);\n    }\n    return { count: values.length, sum, minimum, maximum };\n}\n\ninterface StatsBody {\n\treadonly count: number;\n\treadonly mean: number;\n\treadonly median: number;\n\treadonly p95: number;\n\treadonly min: number;\n\treadonly max: number;\n\treadonly standardDeviation: number;\n}\nfunction numberAt(values: readonly number[], index: number): number {\n\treturn values[index] ?? panic('stats index out of bounds');\n}\nfunction summarize(values: readonly number[]): StatsBody {\n\tconst count = values.length;\n\tlet sum = 0;\n\tfor (let index = 0; index < count; index += 1) sum = sum + numberAt(values, index);\n\tconst mean = sum / count;\n\tconst sorted = values.slice().sort((left, right) => left - right);\n\tconst middle = Math.floor(count / 2);\n\tconst median = count % 2 === 1 ? numberAt(sorted, middle) : (numberAt(sorted, middle - 1) + numberAt(sorted, middle)) / 2;\n\tconst p95 = numberAt(sorted, Math.ceil(0.95 * count) - 1);\n\tlet squaredSum = 0;\n\tfor (let index = 0; index < count; index += 1) {\n\t\tconst difference = numberAt(values, index) - mean;\n\t\tsquaredSum = squaredSum + difference * difference;\n\t}\n\tconst standardDeviation = Math.sqrt(squaredSum / count);\n\treturn { count, mean, median, p95, min: numberAt(sorted, 0), max: numberAt(sorted, count - 1), standardDeviation };\n}\n";
const code = stripTypeScriptTypes(original, {mode: 'strip'});
const {stats, summarize} = await import('data:text/javascript,' + encodeURIComponent("const panic = (message) => { throw new Error(message); };\n" + code + '\nexport {stats, summarize};'));
for (const values of [[4, 1, 3, 2], [9.5, 1.25, 3.75], [-3.5], [4, 4, 4, 4], [0, -0], [-0, 0], [-0], [1e16, 1, -1e16], [-1e16, 1e16, 1], [1e308, 1e308], [1e308, -1e308], [5e-324, 1e-323], [19,18,17,16,15,14,13,12,11,10,9,8,7,6,5,4,3,2,1], [20,19,18,17,16,15,14,13,12,11,10,9,8,7,6,5,4,3,2,1], [21,20,19,18,17,16,15,14,13,12,11,10,9,8,7,6,5,4,3,2,1]]) {
    console.log(JSON.stringify(stats(values)));
    const result = summarize(values);
    console.log(JSON.stringify(result));
    console.log(`signs ${1 / result.min} ${1 / result.max}`);
    for (const value of values) { console.log(`${value} ${1 / value}`); }
}
const parent = [999, 10.5, -2.25, 5.75, 4.125, 999];
const values = parent.slice(1, -1);
    console.log(JSON.stringify(stats(values)));
    const result = summarize(values);
    console.log(JSON.stringify(result));
    console.log(`signs ${1 / result.min} ${1 / result.max}`);
    for (const value of values) { console.log(`${value} ${1 / value}`); }
console.log(`parent ${parent[1]} ${parent[4]}`);
console.log(JSON.stringify(stats([])));
