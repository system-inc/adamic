# Queue JavaScript and source hunks

All ten emitted JavaScript files were compared. Only _tsc.js, typescript.js and run.js differ, entirely inside createQueue. Both complete core.ts transpilation modes also differ only at that owner. Every nonempty source or emitted JavaScript hunk follows.

## _tsc.js

```diff
--- /workspace/scratch/queue-box/lane-count/adapted-tree/built/local/_tsc.js	2026-11-07 03:47:05.613278500 +0000
+++ /workspace/scratch/queue-box/after/built/local/_tsc.js	2026-11-07 04:35:43.667370121 +0000
@@ -822,19 +822,21 @@
   }
 }
 function createQueue(items) {
-  const elements = (items == null ? void 0 : items.slice()) ?? [];
+  const elements = [];
+  if (items) for (const value of items) elements.push({ value });
   let headIndex = 0;
   function isEmpty() {
     return headIndex === elements.length;
   }
   function enqueue(...items2) {
-    elements.push(...items2);
+    for (const value of items2) elements.push({ value });
   }
   function dequeue() {
     if (isEmpty()) {
       throw new Error("Queue is empty");
     }
-    const result = elements[headIndex];
+    const box = elements[headIndex];
+    if (box === void 0) throw new Error("Queue invariant violated");
     elements[headIndex] = void 0;
     headIndex++;
     if (headIndex > 100 && headIndex > elements.length >> 1) {
@@ -848,7 +850,7 @@
       elements.splice(newLength);
       headIndex = 0;
     }
-    return result;
+    return box.value;
   }
   return {
     enqueue,
```

## run.js

```diff
--- /workspace/scratch/queue-box/lane-count/adapted-tree/built/local/run.js	2026-11-07 03:47:38.979839268 +0000
+++ /workspace/scratch/queue-box/after/built/local/run.js	2026-11-07 04:36:21.433667062 +0000
@@ -946,19 +946,21 @@
   }
 }
 function createQueue(items) {
-  const elements = (items == null ? void 0 : items.slice()) ?? [];
+  const elements = [];
+  if (items) for (const value of items) elements.push({ value });
   let headIndex = 0;
   function isEmpty() {
     return headIndex === elements.length;
   }
   function enqueue(...items2) {
-    elements.push(...items2);
+    for (const value of items2) elements.push({ value });
   }
   function dequeue() {
     if (isEmpty()) {
       throw new Error("Queue is empty");
     }
-    const result = elements[headIndex];
+    const box = elements[headIndex];
+    if (box === void 0) throw new Error("Queue invariant violated");
     elements[headIndex] = void 0;
     headIndex++;
     if (headIndex > 100 && headIndex > elements.length >> 1) {
@@ -972,7 +974,7 @@
       elements.splice(newLength);
       headIndex = 0;
     }
-    return result;
+    return box.value;
   }
   return {
     enqueue,
```

## source

```diff
--- /workspace/scratch/queue-box/lane-count/adapted-tree/src/compiler/core.ts	2026-11-07 03:46:39.534406098 +0000
+++ /workspace/scratch/queue-box/after/src/compiler/core.ts	2026-11-07 04:34:10.815393652 +0000
@@ -1577,23 +1577,23 @@
 
 /** @internal */
 export function createQueue<T>(items?: readonly T[]): Queue<T> {
-    const elements: (T | undefined)[] = items?.slice() ?? [];
+    const elements: ({ value: T } | undefined)[] = [];
+    if (items) for (const value of items) elements.push({ value });
     let headIndex = 0;
-
     function isEmpty() {
         return headIndex === elements.length;
     }
 
     function enqueue(...items: T[]) {
-        elements.push(...items);
+        for (const value of items) elements.push({ value });
     }
 
     function dequeue(): T {
         if (isEmpty()) {
             throw new Error("Queue is empty");
         }
-
-        const result = elements[headIndex] as T;
+        const box = elements[headIndex];
+        if (box === undefined) throw new Error("Queue invariant violated");
         elements[headIndex] = undefined; // Don't keep referencing dequeued item
         headIndex++;
 
@@ -1607,7 +1607,7 @@
             headIndex = 0;
         }
 
-        return result;
+        return box.value;
     }
 
     return {
```

## typescript.js

```diff
--- /workspace/scratch/queue-box/lane-count/adapted-tree/built/local/typescript.js	2026-11-07 03:47:06.109257075 +0000
+++ /workspace/scratch/queue-box/after/built/local/typescript.js	2026-11-07 04:35:43.935358316 +0000
@@ -3208,19 +3208,21 @@
   }
 }
 function createQueue(items) {
-  const elements = (items == null ? void 0 : items.slice()) ?? [];
+  const elements = [];
+  if (items) for (const value of items) elements.push({ value });
   let headIndex = 0;
   function isEmpty() {
     return headIndex === elements.length;
   }
   function enqueue(...items2) {
-    elements.push(...items2);
+    for (const value of items2) elements.push({ value });
   }
   function dequeue() {
     if (isEmpty()) {
       throw new Error("Queue is empty");
     }
-    const result = elements[headIndex];
+    const box = elements[headIndex];
+    if (box === void 0) throw new Error("Queue invariant violated");
     elements[headIndex] = void 0;
     headIndex++;
     if (headIndex > 100 && headIndex > elements.length >> 1) {
@@ -3234,7 +3236,7 @@
       elements.splice(newLength);
       headIndex = 0;
     }
-    return result;
+    return box.value;
   }
   return {
     enqueue,
```
