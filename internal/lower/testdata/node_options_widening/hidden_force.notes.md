The structural options view declares only recursive. Its actual value also has
force, whose string value Node must validate rather than silently default away.
On Linux with Node v24.19.0 the source prints:

```text
The "options.force" property must be of type boolean. Received type string ('yes')
```

The host exemption must reject this argument: its annotation hides a runtime
subtype field, so the absence of force has not been proven.
