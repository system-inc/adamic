function kept(values: number[], index: number): number { return values[index]; }
function blocked(values: number[], index: number): number { const value: number = values[index]; values.length = 0; return value; }
function scheduledJson(value: number): string { return JSON.stringify(value); }
