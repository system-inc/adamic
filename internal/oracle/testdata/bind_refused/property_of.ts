export function propertyOf(value: unknown, key: string): unknown {
  if (value === null || value === undefined) return undefined;
  return Reflect.get(Object(value), key, value);
}
