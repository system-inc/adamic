# Worker environment shape boundary (#bxaz0p1, 3a environment)

The string-only bridge projects its declared fields, so observations and widening
are NotYet until a bridge can pass Workers' own env. The interface is identified
by its checker symbol across all entry dependencies, not by a parameter's name.
No KV type or call is admitted by this head.

`TestWorkerEnvironmentShapeMutantsAreCaught` uses buildcache.GoTestMutated through
workerTestProduct. Each child must fail its named pin; an unmutated child must
pass first. Top-level product tests only build their respective mutated binaries.

| Mutation | Named child pin | Catch site |
| --- | --- | --- |
| environment-shape-observed: replace the common observation NotYet with nil | TestWorkerEnvironmentShapeStops/{in,keys,values,entries,for_in,spread,json,has_own,reflect,object,unknown,record,structural,union,alias_keys,rest_destructure,generic_argument,generic_new,generic_rest,array_unknown}; each runs separately | environment_shape_test.go:59, `environment boundary` |
| environment-class-holder: drop the environment interface's slot restriction | TestWorkerEnvironmentShapeStops/{class_field,let,array,module_binding,return,optional_parameter,this_parameter,promise_return}; each runs separately | environment_shape_test.go:59, `environment boundary` |
| environment-parameter-write: drop the destructured-parameter write check | TestWorkerEnvironmentShapeStops/parameter_write | environment_shape_test.go:59, `environment boundary` |
| environment-local-destructure: refuse a const destructuring | TestWorkerEnvironmentFieldReadRoads/const_destructure | environment_shape_test.go:76, `environment field read build` |
| environment-parameter-destructure: refuse a parameter pattern | TestWorkerEnvironmentFieldReadRoads/parameter_destructure | environment_shape_test.go:76, `environment field read build` |
| environment-optional-read: refuse an optional declared-field read | TestWorkerEnvironmentFieldReadRoads/optional_read | environment_shape_test.go:76, `environment field read build` |
| environment-container-methods: walk Promise library methods instead of only its payload | TestWorkerEnvironmentFieldReadRoads/async_field_read | environment_shape_test.go:76, `environment field read build` |
| environment-temporary-slots: drop the implicit element and field slot rule | TestWorkerEnvironmentShapeStops/{array_temp,object_temp} | environment_shape_test.go:59, `environment boundary` |
| environment-parenthesized-argument: treat a parenthesis wrapper as a separate value use | TestWorkerEnvironmentFieldReadRoads/exact_argument | environment_shape_test.go:76, `environment field read build` |
| environment-discovery-missing: return nil instead of the environment type proved by the checked signature | TestWorkerEnvironmentShapeStops/class_field | environment_shape_test.go:59, `environment boundary class_field`; gets the compiler-bug stop instead of Refused |
| environment-intersection-observation: use exact symbol identity alone for observations | TestWorkerEnvironmentShapeStops/e01 | environment_shape_test.go:59, `environment boundary e01`; predicate-narrowed Object.keys bypasses the boundary |
| environment-empty-gate: gate the environment boundary on len(field names) != 0 | TestWorkerEnvironmentEmptyShapeStops | environment_shape_test.go:126, `empty environment observation`; empty interfaces must still reach the observation boundary |

The class-field child additionally logs ordinary Lower's result after the slot
mutant admits the program: this distinguishes removing the boundary from merely
reaching another compiler stop. The positive field-read tests execute the built
Worker bridge in Node. The independent workerd twin and recorded requests remain
under TestWorkerdEnvironmentHandler. The JavaScript-only declared-names pin is
TestEnvironmentBridgeProjectsDeclaredNames; .a observes no environment keys.

The e01 predicate and e02 assertion pins must stop in CheckWorkerEnvironment, naming Object.keys; e10 must stop there naming the widening to Marker. No later unproven-predicate refusal can satisfy these direct boundary pins.

The empty-interface pin calls Build, so bypassing CheckWorkerEnvironment cannot be hidden by a direct boundary call. It requires the observation NotYet, its Object.keys name and lifting condition, and no output directory.
