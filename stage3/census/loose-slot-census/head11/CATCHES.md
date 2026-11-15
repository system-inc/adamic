# Catch inventory at -11

24 AST catch clauses; 7 have bindings; 0 have bindings captured by closures; 0 captured catch bindings stop. There are 5 distinct calls (5 argument entries) passing a catch binding directly to an unknown-typed parameter. Parentheses are unwrapped; derived values such as String(error) are not counted as the caught binding.

Capture means a nested function reads the catch binding’s checker symbol; a catch declared within an existing callback remains local to that callback. Emitter helper templates and comments are source text, not catch AST nodes. An independent TypeScript 6.0.3 AST/checker walk agrees with the Go/checker inventory.

## Stopped captured catch bindings

None. There is no stopped catch-cell clause/closure/H-witness triple to quote in this program.

## Every bound catch clause

### commandLineParser.ts:2301:12 — e

Captured by a closure: **false**. Stops as a captured catch binding: **false**.

`commandLineParser.ts:2301:5`

```typescript
    catch (e) {
```

Capturing closures: none.

### program.ts:406:16 — e

Captured by a closure: **false**. Stops as a captured catch binding: **false**.

`program.ts:406:9`

```typescript
        catch (e) {
```

Capturing closures: none.

### program.ts:441:16 — e

Captured by a closure: **false**. Stops as a captured catch binding: **false**.

`program.ts:441:9`

```typescript
        catch (e) {
```

Capturing closures: none.

### program.ts:2849:16 — e

Captured by a closure: **false**. Stops as a captured catch binding: **false**.

`program.ts:2849:9`

```typescript
        catch (e) {
```

Capturing closures: none.

### sys.ts:1281:20 — e

Captured by a closure: **false**. Stops as a captured catch binding: **false**.

`sys.ts:1281:13`

```typescript
            catch (e) {
```

Capturing closures: none.

### sys.ts:1562:28 — e

Captured by a closure: **false**. Stops as a captured catch binding: **false**.

`sys.ts:1562:21`

```typescript
                    catch (e) {
```

Capturing closures: none.

### sys.ts:1642:24 — error

Captured by a closure: **false**. Stops as a captured catch binding: **false**.

`sys.ts:1642:17`

```typescript
                catch (error) {
```

Capturing closures: none.

## Every catch-binding call to an unknown parameter

### commandLineParser.ts:2302:98

Binding `commandLineParser.ts:2301:12`; argument type `unknown`; parameter `value: unknown` at `hostErrors.ts:12:28`. Independently makes H every shape: **true**.

`commandLineParser.ts:2302:98`

```typescript
        return createCompilerDiagnostic(Diagnostics.Cannot_read_file_0_Colon_1, fileName, String(propertyOf(e, "message")));
```

Parameter declaration:

`hostErrors.ts:12:28`

```typescript
export function propertyOf(value: unknown, key: string): unknown {
```

### program.ts:408:32

Binding `program.ts:406:16`; argument type `unknown`; parameter `value: unknown` at `hostErrors.ts:12:28`. Independently makes H every shape: **true**.

`program.ts:408:32`

```typescript
                onError(String(propertyOf(e, "message")));
```

Parameter declaration:

`hostErrors.ts:12:28`

```typescript
export function propertyOf(value: unknown, key: string): unknown {
```

### program.ts:443:32

Binding `program.ts:441:16`; argument type `unknown`; parameter `value: unknown` at `hostErrors.ts:12:28`. Independently makes H every shape: **true**.

`program.ts:443:32`

```typescript
                onError(String(propertyOf(e, "message")));
```

Parameter declaration:

`hostErrors.ts:12:28`

```typescript
export function propertyOf(value: unknown, key: string): unknown {
```

### sys.ts:1285:43

Binding `sys.ts:1281:20`; argument type `unknown`; parameter `value: unknown` at `hostErrors.ts:12:28`. Independently makes H every shape: **true**.

`sys.ts:1285:43`

```typescript
                hitSystemWatcherLimit ||= propertyOf(e, "code") === "ENOSPC";
```

Parameter declaration:

`hostErrors.ts:12:28`

```typescript
export function propertyOf(value: unknown, key: string): unknown {
```

### sys.ts:1563:29

Binding `sys.ts:1562:28`; argument type `unknown`; parameter `error: unknown` at `hostErrors.ts:2:27`. Independently makes H every shape: **true**.

`sys.ts:1563:29`

```typescript
                        if (errorCode(e) !== "EEXIST") {
```

Parameter declaration:

`hostErrors.ts:2:27`

```typescript
export function errorCode(error: unknown): string | undefined {
```
