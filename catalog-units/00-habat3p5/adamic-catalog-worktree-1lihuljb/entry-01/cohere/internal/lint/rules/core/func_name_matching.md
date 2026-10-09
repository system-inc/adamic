# `func-name-matching`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **178** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Require function names to match the name of the variable or property to which they are assigned

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

178 in the tree. Showing the first few.

**`app/(os-layout)/_components/select/TaskListMarquee.tsx:180`**

```
tickReference.current = function tick() {
```

> Function name `tick` should match property name `current`

**`app/(os-layout)/art/_components/ArtGallery.tsx:77`**

```
request: async function fetchArt(): Promise<ArtResponseInterface> {
```

> Function name `fetchArt` should match property name `request`

**`app/(os-layout)/art/_components/UniverseWeightsDialogBody.tsx:48`**

```
request: async function fetchUniverseWeights(): Promise<GetResponseInterface> {
```

> Function name `fetchUniverseWeights` should match property name `request`

**`app/(os-layout)/contacts/_hooks/useContact.ts:124`**

```
request: async function fetchContact(): Promise<{ contact: ContactDetailInterface }> {
```

> Function name `fetchContact` should match property name `request`

**`app/(os-layout)/contacts/_hooks/useContacts.ts:43`**

```
request: async function fetchContacts(): Promise<ContactsResponseInterface> {
```

> Function name `fetchContacts` should match property name `request`

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

