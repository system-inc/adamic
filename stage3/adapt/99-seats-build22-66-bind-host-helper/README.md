# Builder writeFile binding

Adamic's November 22 21:49Z ruling on #bbssdcy approves this helper for family 66
in the shape of `docs/rulings.md:120`. The helper's receiver is one plain,
never-assigned parameter. It reads `writeFile` at bind time, captures that
receiver for the returned callback, and returns `undefined` when the property is
absent. Each fallback calls the helper only when its supplied `writeFile` is falsy.

| Adapted input site | Before | After |
| --- | --- | --- |
| `builder.ts:1750:30` | `writeFile \|\| host.writeFile?.bind(host)` | `writeFile \|\| bindHostWriteFile(host)` |
| `builder.ts:1838:34` | `writeFile \|\| host.writeFile?.bind(host)` | `writeFile \|\| bindHostWriteFile(host)` |
| `builder.ts:1937:78` | `writeFile \|\| host.writeFile?.bind(host)` | `writeFile \|\| bindHostWriteFile(host)` |

The helper is private to `builder.ts`. It follows the existing declarations so
their source lines stay stable. The diagnostic wording has its own Diagnostics
handoff under the same ruling.
