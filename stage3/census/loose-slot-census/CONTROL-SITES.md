# Control sites

Distinct sites are exact source locations; these controls are separate from the 79-file count.

## arrays

9 events; 9 distinct sites. Forms: {'write': 4, 'literal init': 5}.

### arrays.a:48:1 — write

```typescript
itself.push(itself);
```

### arrays.a:52:1 — write

```typescript
after.push(after);
```

### arrays.a:57:1 — write

```typescript
outer.push(inner);
```

### arrays.a:63:1 — write

```typescript
holder.push(new Back());
```

### arrays.a:31:55 — literal init

```typescript
const mixed: Item[] = ['a', 1, null, undefined, true, new Named(), ['b', 'c'], new Error('e')];
```

### arrays.a:31:68 — literal init

```typescript
const mixed: Item[] = ['a', 1, null, undefined, true, new Named(), ['b', 'c'], new Error('e')];
```

### arrays.a:31:80 — literal init

```typescript
const mixed: Item[] = ['a', 1, null, undefined, true, new Named(), ['b', 'c'], new Error('e')];
```

### arrays.a:56:29 — literal init

```typescript
const inner: Item[] = ['i', outer];
```

### arrays.a:65:32 — literal init

```typescript
const throwing: Item[] = ['t', new Fails()];
```

## k17_captured_unknown

1 events; 1 distinct sites. Forms: {'captured': 1}.

### k17_captured_unknown.ts:3:11 — captured

```typescript
    const held: unknown = node;
```

## k18_entry_then_declared

1 events; 1 distinct sites. Forms: {'literal init': 1}.

### k18_entry_then_declared.ts:6:30 — literal init

```typescript
const a: Node = { name: "a", box: hide(holder) };
```
