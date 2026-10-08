# Step 42 compiler selectors

Call spread is #aqh658t; JSON literal shorthand is #kqxkvg0. Both were filed by
the user in compiler's tree. These are the unchanged shortest proving programs
from stage1/cohere/command. Node must run them, and stage 0 must refuse them.
The selector checks in TestSettingsResolution and TestSettingsOutputInterface
fail when stage 0 starts lowering either program, requiring removal of the
corresponding workaround from stage 1.

Authorized recorded workarounds:

- #aqh658t: push one inherited ignore pattern at a time.
- #kqxkvg0: write the status property out as `status: status` in JSON literals.

Every deployed workaround site names its task. No compiler implementation is
changed. Reproduce both selectors with:

```sh
go test -v -count=1 -run '^TestSettings(Resolution|OutputInterface)$' ./stage1/cohere/command
```

The follow-up also retains `json_reference.a`: structural-object JSON.stringify
is refused because runtime shapes need complete value metadata. The original
whole-answer transport remains blocked; no workaround is deployed for this
third refusal. Independent scalar property probes test resolver values and
fixture array projections, without claiming complete JSON object serialization.
This gap has no assigned task ID in the supplied brief.
