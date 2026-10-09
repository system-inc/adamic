export interface Entry { key: string; value: number }
export type Index<T> = { [K in keyof T]: T[K] };
export function total(entries: Entry[]): number {
  let sum = 0;
  for (let i = 0; i < entries.length; i++) sum += entries[i].value;
  return sum;
}
export const sample: Index<Entry> = {key: "one", value: 1};
