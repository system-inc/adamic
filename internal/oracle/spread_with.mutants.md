# Mutants for #k7g6f9h's SpreadWith literal

Each is a wrong version of the lowering or the emitter, planted on top of the head, with the rows it must turn red. The
mutant job plants each in turn, runs its rows uncached, and passes only when every named row ran and is red; the head's
rows are green first. W1 to W5 replace #xy74m5x's S1, whose stop is gone.

**W1, the width dropped.** objectLiteral copies the source by its static type's names into the literal's own fields and
spreads undefined, so the result holds only what the type declares. Red: `TestSpreadWithAgrees/spread_with/width_extra.ts`
(extra is gone from Object.keys, JSON.stringify, for-in and the later spread).

**W2, written by the source's static shape.** The emitter passes SpreadWith over and takes the copy arm, writing each
field into the slot the source's shape has, as #xy74m5x found. Red: `TestSpreadWithAgrees/spread_overwrite/number_present.ts`,
`optional_string.ts` and `spread_with/kind_change.ts`.

**W3, a uniform slot for an appended name.** uniformFieldOffsets records nothing for a SpreadWith literal, so k keeps the
plain literal's slot 0 and the result's k is read there. Red: `TestSpreadWithAgrees/spread_with/uniform.ts`.

**W4, no copy before a value that runs code.** The emitter always holds the source and builds from it after the values, so
a value that writes the source shows in the result. Red: `TestSpreadWithAgrees/spread_with/effect_order.ts`.

**W5, the accessor stop dropped.** The registry's stop (finishAccessors, spreadAccessorClass) is skipped, so an
object literal's own getter reaches adamic_object_spread_with, which passes over it where Node's spread runs it and
copies its value. Red: `TestSpreadWithStops/spread_with/accessor_literal.ts`.

**W6, an overwrite value counted twice.** Each reference value is retained once more on its way into the call. Red, on the
counted build: `TestSpreadWithAgrees/spread_with/resolver_279.ts` and `graph_fresh.ts` (fresh arrays and a fresh literal).

**W7, a graph value held after adoption too.** Today's copy arm's hold is added on top of the primitive's count, after the
adoption (Memory's mutant). On a member value it can't show: the adoption's merge takes the count off every edge it
brings inside one region, and adamic_graph_hold returns a member value without a retain (graph_regions.c), so
graph_member's and graph_fresh's counted rows stay balanced under it, and W6 keeps their count. On a value that isn't a
member, a fresh string, the hold retains it (graph_regions.c:672-678), one count too many. Red, on the counted build:
`TestSpreadWithAgrees/spread_with/graph_string.ts`.

**W8, the native storage check dropped.** Library's positive test in object_spread.c (`if (!plain)`) never stops, her
`storage-check-dropped`, planted on this stacked tree, so a tuple is copied as a plain object's slots. Red:
`TestSpreadWithSourceBackstop/backstop_tuple.ts`, on its native row.

**W9, the JavaScript mirror dropped.** The JavaScript backend spreads the source as it is, so a tuple's indices are
copied where native stops. Red: `TestSpreadWithSourceBackstop/backstop_tuple.ts`, on its javascript row.

**W10, the getter walk blind to a SpreadWith literal.** spreadCopyRead judges `{ ...inner }` by its own spread's type
alone, which holds inner's `extra` and so no Sized, and the read of weight lowers and loads a slot the copy lacks
(Memory's mutant). Red: `TestSpreadWithGetterReads/getter_copy_of_copy.ts`.

**W11, an object literal's holder judged by assignability.** spreadAccessorClass tests an object literal's own fresh
type against the spread's view, which the checker relates with the excess-property check, so a literal with a getter
reaching a narrower view by width is passed over (J-5e03qs3), and natively Library's backstop stops where Node runs.
Red: `TestSpreadWithStops/spread_with/accessor_literal.ts`.

Must stay green under each: `TestSpreadOverwriteNeighborsAgree` and `TestSpreadWithLeavesNeighbors`. A pin row the C's
own mutant turns red is Library's to name in object_spread_test.go; W8 is hers planted here to show the backend rows.
