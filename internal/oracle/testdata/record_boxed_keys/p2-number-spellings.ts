const record: Record<string, string> = {
    "0": "zero", "1e+21": "big", "1e-7": "small",
    "0.30000000000000004": "fraction", "NaN": "nan",
    "Infinity": "positive infinity", "-Infinity": "negative infinity",
};
function show(key: string | number): void {
    console.log(String(record[key]));
}
show(-0);
show(1e21);
show(1e-7);
show(0.1 + 0.2);
show(NaN);
show(Infinity);
show(-Infinity);
