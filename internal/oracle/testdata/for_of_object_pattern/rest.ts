// A rest element keeps the declaration road's stop (#rm22wzv).
const rows = [{ a: 1, b: 2 }];
for (const { a, ...others } of rows) {
    console.log(String(a) + String(others.b));
}
