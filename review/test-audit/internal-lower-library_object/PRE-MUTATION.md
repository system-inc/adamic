Starting commit: e2492670b06a4dce1deafe837158ce0366cf2bc4.
Code under test: Adamic lowering, object/refusal/type-view logic, console IR production, diagnostic formatting, and the direct regex capture-free proof.
Oracle: handwritten error classes and diagnostic strings, handwritten IR expectations, or rejection/non-rejection predicates, all self. No outside authority was checked.
Complete pre-mutation reached-function inventory: reached-functions.txt (447 coverage-observed functions in internal/lower). Anonymous visitors are included within their owning function coverage.
Fixed menu: menu.json, written before checking mutant failures. All M01-M20 are condition flips, changed constants, off-by-one bounds or early returns from production behavior. P01/P02 are empty entry probes, excluded from production kills.
All 11 names exist in the starting test list and remain in the four assigned files. Each is independent: shared lowerSource performs preparation, but their assertions and checks differ. No families, witnesses or helpers in the assigned set.
