const row = { get computed(): string { return "got"; }, plain: "p" } as const;
console.log(Object.entries(row).map(([k, v]) => k + "=" + v).join(","));
