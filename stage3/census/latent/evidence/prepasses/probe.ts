interface Box { value?: number; }
const box: Box = {};
export function first(): void { box.value = 1; }
export function second(): void { box.value = 2; }
