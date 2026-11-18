# Findings changed outside the historical 40 sites

Pair key: kind, phase, reason, and text with `:line:column` substrings removed. Repeated keys are paired at identical unit/site first, then within unit, then by normalized key. All eight raw events are listed; no Boundary records were discarded.

## 1. gone — utilitiesPublic

base: `Boundary` / `lowering` at `/workspace/depthgate-adapted/src/compiler/utilitiesPublic.ts:625:9`; unit `/workspace/depthgate-adapted/src/compiler/utilitiesPublic.ts:623:1`.

Reason: failed statement; state rolled back; continued at next statement or declaration

````text
/workspace/depthgate-adapted/src/compiler/utilitiesPublic.ts:625:22: stage 0 can't lower a generic comparison deeper than 4 positions that can reach a Map yet
````

## 2. gone — utilitiesPublic

base: `Boundary` / `lowering` at `/workspace/depthgate-adapted/src/compiler/utilitiesPublic.ts:2638:5`; unit `/workspace/depthgate-adapted/src/compiler/utilitiesPublic.ts:2637:1`.

Reason: failed statement; state rolled back; continued at next statement or declaration

````text
/workspace/depthgate-adapted/src/compiler/utilitiesPublic.ts:2638:76: stage 0 can't lower a generic comparison deeper than 4 positions that can reach a Map yet
````

## 3. gone — utilitiesPublic

base: `Boundary` / `lowering` at `/workspace/depthgate-adapted/src/compiler/utilitiesPublic.ts:2639:5`; unit `/workspace/depthgate-adapted/src/compiler/utilitiesPublic.ts:2637:1`.

Reason: failed statement; state rolled back; continued at next statement or declaration

````text
/workspace/depthgate-adapted/src/compiler/utilitiesPublic.ts:2639:14: stage 0 can't lower reading last yet
````

## 4. gone — utilitiesPublic

base: `NotYet` / `lowering` at `/workspace/depthgate-adapted/src/compiler/utilitiesPublic.ts:2639:14`; unit `/workspace/depthgate-adapted/src/compiler/utilitiesPublic.ts:2637:1`.

Reason: reading last

````text
/workspace/depthgate-adapted/src/compiler/utilitiesPublic.ts:2639:14: stage 0 can't lower reading last yet
````

## 5. gone — utilities

base: `Boundary` / `lowering` at `/workspace/depthgate-adapted/src/compiler/utilities.ts:3095:5`; unit `/workspace/depthgate-adapted/src/compiler/utilities.ts:3113:1`.

Reason: failed statement; state rolled back; continued at next statement or declaration

````text
/workspace/depthgate-adapted/src/compiler/utilities.ts:3095:20: stage 0 can't lower a generic comparison deeper than 4 positions that can reach a Map yet
````

## 6. gone — utilities

base: `Boundary` / `lowering` at `/workspace/depthgate-adapted/src/compiler/utilities.ts:3095:5`; unit `/workspace/depthgate-adapted/src/compiler/utilities.ts:12296:1`.

Reason: failed statement; state rolled back; continued at next statement or declaration

````text
/workspace/depthgate-adapted/src/compiler/utilities.ts:3095:20: stage 0 can't lower a generic comparison deeper than 4 positions that can reach a Map yet
````

## 7. gone — utilities

base: `Boundary` / `lowering` at `/workspace/depthgate-adapted/src/compiler/utilities.ts:3095:5`; unit `/workspace/depthgate-adapted/src/compiler/utilities.ts:12304:1`.

Reason: failed statement; state rolled back; continued at next statement or declaration

````text
/workspace/depthgate-adapted/src/compiler/utilities.ts:3095:20: stage 0 can't lower a generic comparison deeper than 4 positions that can reach a Map yet
````

## 8. gone — transformers-utilities

base: `Boundary` / `lowering` at `/workspace/depthgate-adapted/src/compiler/transformers/utilities.ts:127:5`; unit `/workspace/depthgate-adapted/src/compiler/transformers/utilities.ts:124:1`.

Reason: failed statement; state rolled back; continued at next statement or declaration

````text
/workspace/depthgate-adapted/src/compiler/transformers/utilities.ts:127:17: stage 0 can't lower a generic comparison deeper than 4 positions that can reach a Map yet
````
