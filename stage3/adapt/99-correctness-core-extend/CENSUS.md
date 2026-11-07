# Complete core.ts site census

The exact guarded full-site workflow from core-sites-4186581e is used:
LATENT_FULL=1, LATENT_ONLY_FILE set to the complete core.ts absolute path,
LATENT_ASSERT_NO_OUTPUT=1, LATENT_TIMES set to a retained progress file. The
separate measurement binary is built with main's make_overlay.py. No production
compiler source, backend, refusal suppression or root fallback is changed.
Every retained Go source is named .go.txt.

All three streams have zero checker diagnostics. The meter's historical label
"measured on a checker-rejected program" remains unchanged in the raw data.
The complete file, including every unrelated unit and finding, is retained.

| Site | Before | After | Original-owner mutant |
| --- | --- | --- | --- |
| F6, 1506:29 | Refused unchecked cast | Absent | Restored |
| F7, 1509:14 | Refused unchecked cast | Absent | Restored |
| F8, 1509:36 | Refused unchecked cast | Absent | Restored |
| F9, 1515:14 | Refused unchecked cast | Absent | Restored |
| F10, 1515:36 | Refused unchecked cast | Absent | Restored |
| F11, 1526:14 | Refused unchecked cast | Absent | Restored |

There are 59 distinct (location, kind, reason, phase) tuples before and 53 after.
Exactly these six disappear. There is no new Refused or NotYet finding anywhere
in core.ts. Restoring both complete original owners returns exactly the entire
baseline finding set, not only the six selected sites.

Both owners remain deferred_generic_instantiation with instantiated=false on
both sides and in the mutant. Absence of a cast-scan finding does not establish
that either generic body has been lowered. The earlier census wall remains
getOwnValues, core.ts:1326:26; this row does not clear other ranked rows.
The selected type-only spelling clears all six requested findings. A
JavaScript-changing implementation is unnecessary and is not built.
