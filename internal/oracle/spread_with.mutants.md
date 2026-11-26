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

**W5, the accessor stops dropped.** The registry's stop (finishAccessors) and the type's (objectLiteral) are skipped, so a
class instance with an accessor reaches adamic_object_spread_with, which refuses it at run time. Red:
`TestSpreadWithStops/spread_with/accessor_getter.ts`, `accessor_setter.ts`, `accessor_literal.ts` and `accessor_type.ts`.

**W5b, the type's stop alone dropped.** The registry still stops the class, in its own words, so the row tells them apart.
Red: `TestSpreadWithStops/spread_with/accessor_type.ts`.

**W6, an overwrite value counted twice.** Each reference value is retained once more on its way into the call. Red, on the
counted build: `TestSpreadWithAgrees/spread_with/resolver_279.ts` and `graph_fresh.ts` (fresh arrays and a fresh literal).

**W7, a graph value held after adoption too.** Today's copy arm's hold is added on top of the primitive's count, after the
adoption (Memory's mutant). Red, on the counted build: `TestSpreadWithAgrees/spread_with/graph_member.ts` and `graph_fresh.ts`.

**W8, the JavaScript backstop dropped.** The JavaScript backend spreads the source as it is, so a tuple's indices are
copied where native stops. Red: `TestSpreadWithSourceBackstop` (its javascript row exits 0). Native's check is the
primitive's, in Library's C.

Must stay green under each: `TestSpreadOverwriteNeighborsAgree` and `TestSpreadWithLeavesNeighbors`.
