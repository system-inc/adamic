# Completed on abebb00b

Implementation: 90daeca3ba669cd4e1da154b984a201a83ad95e2, one commit over abebb00bf6312fb92798095515fbee86b8ed78e3.

All requested proofs pass. Full census: 933 units before, after and under the restoring mutant; exactly the Refused site at utilities.ts:10005:158 disappears, no new sites or phases appear, and the mutant matches baseline. All ten emitted JavaScript files and both public declarations are byte-identical. Stock tsc CLI and adamic types each report zero on both 79-root sets with unchanged strict options. Full apply, owner-drift contract, zero-edit replay, and generate_shards --check pass. Baseline apply restores a verified keyed product; candidate apply was forced with --build-product.

The entire stage3/lane Go package passes: all 398 shards ran and passed, with no skipped shards. Command: go test -json -count=1 -parallel=2 -timeout 27m ./stage3/lane. The Go child's exit is 0; wall time is 1554.729s. Every supervised process is limited to 1700s and reports its child's own exit code. Physical Go evidence files end in .go.txt.

The first complete lane run at parallelism four ended with 397 passing shards and one relationComplexityError.ts before-all hook failure. The same shard then passed all six cases alone in 14.6s under unchanged limits. The complete two-way rerun passed all 398, including that shard. Both whole-run logs and the isolated raw report are retained. No test timeout, checker option, inventory, expected count or source was changed to obtain the pass.

Main moved to e42d56d2b during the lane. The completed proofs remain explicitly on the started tip, abebb00b. Its later path-from-components adapter edits only path.ts and a different owner; later-adapter-composition.json verifies both adapters produce identical bytes in either order and replay with zero edits. No census or lane verdict is claimed for e42d56d2b.

summary.json is the authoritative structured result. run.py and the raw command status files retain child exits; log.gz files hold raw output, census JSONL streams have completion trailers, and provenance.json binds tool hashes and pins. The no-output census uses LATENT_FULL=1, LATENT_ONLY_FILE for utilities.ts and LATENT_ASSERT_NO_OUTPUT=1. The overlay generator changes only physical scratch filenames to .go.txt.
