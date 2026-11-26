// A nested pattern keeps the declaration road's stop (#rm22wzv).
const rows = [{ inner: { value: 1 } }];
for (const { inner: { value } } of rows) {
    console.log(String(value));
}
