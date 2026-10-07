# CFG read, write and loop hooks

One Go private method per .a file. The caller's numeric builder and node arena handles stand for original pointer identity; -1 represents a nil node. present represents whether that builder's corresponding Go hook is nonnil. The callback argument remains callable even when present is false, so tests can prove the presence guard rather than credit a crash. Callers must project presence from the current builder's actual hook registration.

read and write call their hook exactly once when present is true. They forward both handles unchanged, including a nil node. They do not filter on reachability. loop calls its hook exactly once only when present is true and the node is nonnil. None changes builder state outside whatever the callback itself does. The callback may mutate the original builder, including its current block, and those mutations remain visible to the caller.

The Go oracle overlays method-entry capture calls onto the original cfg.go, leaving every original body unchanged. It parses every distinct runtime source captured from the four consuming rule fixture sets and builds all IndexRoots with actual Go Build and empty hooks. Each helper invocation is counted. Since these methods never inspect node contents, replay deduplicates method/node-kind/reachability call shapes and retains one actual parser-node pointer for each kind. All captured call shapes are expanded across present/absent hooks, both builder identities and two callback modes. Controls additionally cover nil and every representative kind for all three operations and both reachabilities.

Replay invokes actual private Go read/write/loop methods. The test callback records the exact builder and node pointer identities, emits into the current block, and optionally moves the builder to the other block before emitting again. Whole state and ordered callback traces are compared to source Node, emitted JavaScript and ASan/UBSan native. Expected output is removed before Adamic reads inputs. Mutants must compile and finish without stderr on all backends before an independent Go mismatch counts.

The input capture runs ordinary Go Core and React tests through an overlay, without editing cohere. This is helper coverage over every consumer fixture and all roots, including roots beyond each rule's filtered graph selection. It does not instrument the consumers' exact filtered graph invocations, port the complete CFG builder or implement whole-rule findings. Numeric adapter validation, callback failure behavior and concurrent hook replacement are outside coverage.

From the repository root, source /workspace/adamic-tools/env.sh, then run:

```
python3 stage1/cohere/lint/helpers/slot02/batch10/testdata/regenerate.py > /tmp/slot02-batch10-regeneration.log 2>&1
go test ./stage1/cohere/lint/helpers -run '^TestSlot02Batch10$' -count=1 -v -timeout=20m > /tmp/slot02-batch10.log 2>&1
```

No shared harness/generator, compiler implementation, rule listener, Diagnostic model or regexp matcher changes. See REPORT.md for commands, all semantic mutants and limits.
