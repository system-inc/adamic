# Mutants for the one iterator result builder (#ak92fcp)

Each mutant is one edit to the runtime, run on a Linux box (Cloud, job 01M6R38JK9W0B2KVE81GQ1RTNP, at
48a4d4fb) against internal/oracle's TestIteratorResultBuilderAgreesWithNode and
TestIteratorResultKeysAgreeWithNode with `ADAMIC_GATE_UNCACHED=1`, and internal/native's
TestCollectionIteratorStepsCountedElement and TestCollectionElementOfAnotherRepresentationStops, then put
back from a saved copy and checked with `cmp`.

| Mutant | Edit | Red |
|---|---|---|
| M5, the slots swapped | runtime/map_set.c, adamic_iterator_result: value written to slot 1 and done to slot 0, their representation bytes with them | TestIteratorResultKeysAgreeWithNode (collections.a), TestIteratorResultBuilderAgreesWithNode and TestCollectionIteratorStepsCountedElement: native exits differently, and the sanitized run stops on a misaligned read of the address 1 |
| M6, main's done bytes for a boolean | runtime/map_set.c: a boolean kind at done holds the absent maybe-number word with the MaybeNumber byte, as collection_next did before this change | TestIteratorResultBuilderAgreesWithNode: steps.a's `set values done` prints `true` where Node prints undefined. This is the silent wrong answer the change fixes |
| M7, the representation byte dropped | runtime/map_set.c: the builder's `adamic_object_field_types(result)[0] = element_tag;` made `(void)element_tag;` | TestIteratorResultBuilderAgreesWithNode: the read by a run-time key panics "dynamic read of a host scalar without type metadata"; TestCollectionIteratorStepsCountedElement goes red too |
| M9, no case 9 | runtime/union.c, dynamic_slot: both `case 9` arms removed | TestIteratorResultBuilderAgreesWithNode: the run-time key read of a finished boolean result panics "dynamic read of a host scalar without type metadata" |
| MB, collection_reference's old body | runtime/map_set.c: collection_reference returns true for anything but Number, Boolean and MaybeNumber | TestCollectionElementOfAnotherRepresentationStops: a MaybeBoolean element is retained as a pointer, with no panic, and the harness gets no stop message |

string_next now writes representation bytes, value String and done Boolean, where it left both zero
(no evidence), matching a string collection's result; no program reaches a string iterator's result
yet (#7nw4kgj).
