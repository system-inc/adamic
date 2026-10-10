package estree

// Measurement checkpoint: base e77a4ae41f473c149aee910c51b73637686a804a.
// R2 restore succeeded; four-CPU quota; Go 1.27.1 cache warmed.
// Sequential anchored commands; products use -timeout 10m; units use -timeout 90s.
// No whole-package test run. Command wall seconds follow.
// {"exit": 0, "unit": "TestProduct_AcceptanceCatchLowered", "wall": 23.752}
// {"exit": 0, "unit": "TestProduct_AcceptanceCatchNative", "wall": 32.595}
// {"exit": 0, "unit": "TestProduct_AcceptanceClassLowered", "wall": 26.42}
// {"exit": 0, "unit": "TestProduct_AcceptanceClassNative", "wall": 26.847}
// {"exit": 0, "unit": "TestProduct_AcceptanceOracle", "wall": 12.262}
// {"exit": 0, "unit": "TestProduct_DeepMutantsFixture", "wall": 3.224}
// {"exit": 0, "unit": "TestProduct_DeepMutantsIR0", "wall": 3.296}
// {"exit": 0, "unit": "TestProduct_DeepMutantsIR1", "wall": 3.135}
// {"exit": 0, "unit": "TestProduct_DeepMutantsIR2", "wall": 3.204}
// {"exit": 0, "unit": "TestProduct_DeepMutantsLowered0", "wall": 19.36}
// {"exit": 0, "unit": "TestProduct_DeepMutantsLowered1", "wall": 18.788}
// {"exit": 0, "unit": "TestProduct_DeepMutantsLowered2", "wall": 19.191}
// {"exit": 0, "unit": "TestProduct_DeepMutantsNative0", "wall": 9.574}
// {"exit": 0, "unit": "TestProduct_DeepMutantsNative1", "wall": 9.854}
// {"exit": 0, "unit": "TestProduct_DeepMutantsNative2", "wall": 10.105}
// {"exit": 0, "unit": "TestProduct_LossyInputControlLowered", "wall": 23.277}
// {"exit": 0, "unit": "TestProduct_LossyInputControlNative", "wall": 8.293}
// {"exit": 0, "unit": "TestProduct_RecoveryMutantsAnswer", "wall": 4.314}
// {"exit": 0, "unit": "TestProduct_RecoveryMutantsLowered0", "wall": 22.865}
// {"exit": 0, "unit": "TestProduct_RecoveryMutantsLowered1", "wall": 24.237}
// {"exit": 0, "unit": "TestProduct_RecoveryMutantsLowered2", "wall": 22.919}
// {"exit": 0, "unit": "TestProduct_RecoveryMutantsNative0", "wall": 22.635}
// {"exit": 0, "unit": "TestProduct_RecoveryMutantsNative1", "wall": 24.293}
// {"exit": 0, "unit": "TestProduct_RecoveryMutantsNative2", "wall": 9.003}
// {"exit": 0, "unit": "TestProduct_SyntaxMutantIR_000", "wall": 4.087}
// {"exit": 0, "unit": "TestProduct_SyntaxMutantIR_001", "wall": 4.306}
// {"exit": 0, "unit": "TestProduct_SyntaxMutantIR_002", "wall": 3.538}
// {"exit": 0, "unit": "TestProduct_SyntaxMutantLowered_000", "wall": 22.8}
// {"exit": 0, "unit": "TestProduct_SyntaxMutantLowered_001", "wall": 22.153}
// {"exit": 0, "unit": "TestProduct_SyntaxMutantLowered_002", "wall": 24.116}
// {"exit": 0, "unit": "TestProduct_SyntaxMutantNative_000", "wall": 15.792}
// {"exit": 0, "unit": "TestProduct_SyntaxMutantNative_001", "wall": 5.423}
// {"exit": 0, "unit": "TestProduct_SyntaxMutantNative_002", "wall": 5.342}
// {"exit": 0, "unit": "TestProduct_SyntaxMutantsGoOracle", "wall": 3.512}
// {"exit": 0, "unit": "TestProduct_SyntaxMutantsSetup_000", "wall": 2.769}
// {"exit": 0, "unit": "TestProduct_SyntaxMutantsSetup_001", "wall": 3.273}
// {"exit": 0, "unit": "TestProduct_SyntaxMutantsSetup_002", "wall": 2.89}
// {"exit": 0, "unit": "TestProduct_ThreePortLowered0", "wall": 26.354}
// {"exit": 0, "unit": "TestProduct_ThreePortLowered1", "wall": 26.041}
// {"exit": 0, "unit": "TestProduct_ThreePortLowered2", "wall": 25.499}
// {"exit": 0, "unit": "TestProduct_ThreePortNative0", "wall": 8.602}
// {"exit": 0, "unit": "TestProduct_ThreePortNative1", "wall": 30.05}
// {"exit": 0, "unit": "TestProduct_ThreePortNative2", "wall": 8.597}
// {"exit": 0, "unit": "TestProduct_ThreePortOracle", "wall": 5.082}
// {"exit": 0, "unit": "TestProduct_UnattachedDecoratorIR", "wall": 3.983}
// {"exit": 0, "unit": "TestProduct_UnattachedDecoratorLowered", "wall": 22.969}
// {"exit": 0, "unit": "TestProduct_UnattachedDecoratorNative", "wall": 14.272}
// {"exit": 0, "unit": "TestProduct_scalar_edges_go_oracle", "wall": 4.017}
// {"exit": 0, "unit": "TestProduct_scalar_edges_lowered", "wall": 23.392}
// {"exit": 0, "unit": "TestProduct_scalar_edges_sanitized_native", "wall": 5.251}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestThreePortMutants_Setup", "wall": 10.87}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestThreePortMutants_Setup", "wall": 11.088}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestThreePortMutants_005", "wall": 11.735}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestThreePortMutants_005", "wall": 11.073}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestThreePortMutants_007", "wall": 11.026}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestThreePortMutants_007", "wall": 11.031}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestThreePortMutants_000", "wall": 11.911}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestThreePortMutants_000", "wall": 11.385}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestThreePortMutants_001", "wall": 11.271}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestThreePortMutants_001", "wall": 11.618}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestThreePortMutants_008", "wall": 11.338}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestThreePortMutants_008", "wall": 12.434}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestThreePortMutants_009", "wall": 11.429}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestThreePortMutants_009", "wall": 12.638}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestThreePortMutants_011", "wall": 12.501}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestThreePortMutants_011", "wall": 12.545}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestRecoveryMutants_Setup", "wall": 7.588}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestRecoveryMutants_Setup", "wall": 7.252}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestThreePortMutants_010", "wall": 11.094}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestThreePortMutants_010", "wall": 11.3}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestThreePortMutants_006", "wall": 11.582}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestThreePortMutants_006", "wall": 11.082}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestDeepMutants_001", "wall": 3.265}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestDeepMutants_001", "wall": 3.312}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestThreePortMutants_002", "wall": 11.943}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestThreePortMutants_002", "wall": 12.053}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestThreePortMutants_004", "wall": 11.54}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestThreePortMutants_004", "wall": 11.465}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestThreePortMutants_003", "wall": 11.252}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestThreePortMutants_003", "wall": 12.197}
// {"exit": 0, "run": 1, "timing": ["syntax_mutants_self_setup_test.go:285: setup: 0.285s", "syntax_mutants_self_setup_test.go:287: own work: 1.470s"], "unit": "TestSyntaxMutantsUnion", "wall": 3.402}
// {"exit": 0, "run": 2, "timing": ["syntax_mutants_self_setup_test.go:285: setup: 0.276s", "syntax_mutants_self_setup_test.go:287: own work: 1.435s"], "unit": "TestSyntaxMutantsUnion", "wall": 3.277}
// {"exit": 0, "run": 1, "timing": ["syntax_mutants_self_setup_test.go:228: setup: 0.291s"], "unit": "TestSyntaxMutants_002", "wall": 2.433}
// {"exit": 0, "run": 2, "timing": ["syntax_mutants_self_setup_test.go:228: setup: 0.272s"], "unit": "TestSyntaxMutants_002", "wall": 2.381}
// {"exit": 0, "run": 1, "timing": ["syntax_mutants_self_setup_test.go:228: setup: 0.311s"], "unit": "TestSyntaxMutants_001", "wall": 3.077}
// {"exit": 0, "run": 2, "timing": ["syntax_mutants_self_setup_test.go:228: setup: 0.270s"], "unit": "TestSyntaxMutants_001", "wall": 2.194}
// {"exit": 0, "run": 1, "timing": ["syntax_mutants_self_setup_test.go:228: setup: 0.268s"], "unit": "TestSyntaxMutants_000", "wall": 2.152}
// {"exit": 0, "run": 2, "timing": ["syntax_mutants_self_setup_test.go:228: setup: 0.275s"], "unit": "TestSyntaxMutants_000", "wall": 2.469}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestAcceptanceMutants_000", "wall": 5.698}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestAcceptanceMutants_000", "wall": 5.618}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestBoundedPortParser_003", "wall": 41.158}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestBoundedPortParser_003", "wall": 3.866}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestAcceptanceMutantsPlanted_001", "wall": 5.447}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestAcceptanceMutantsPlanted_001", "wall": 4.96}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestScalarEdges_000", "wall": 5.033}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestScalarEdges_000", "wall": 5.028}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestDeepMutants_002", "wall": 3.239}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestDeepMutants_002", "wall": 3.162}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestRecoveryLoweredRecipe", "wall": 41.136}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestRecoveryLoweredRecipe", "wall": 2.704}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestDeepMutants_000", "wall": 3.483}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestDeepMutants_000", "wall": 3.314}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestScalarEdges_001", "wall": 5.064}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestScalarEdges_001", "wall": 5.314}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestBoundedPortParser_052", "wall": 2.47}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestBoundedPortParser_052", "wall": 2.551}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestPortStallControl_000", "wall": 40.901}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestPortStallControl_000", "wall": 5.242}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestBoundedPortParser_057", "wall": 3.218}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestBoundedPortParser_057", "wall": 2.602}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestBoundedPortParser_058", "wall": 2.882}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestBoundedPortParser_058", "wall": 2.743}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestBoundedPortParser_041", "wall": 2.603}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestBoundedPortParser_041", "wall": 2.555}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestAcceptanceMutants_001", "wall": 6.456}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestAcceptanceMutants_001", "wall": 5.044}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestDecoratedExportMutant", "wall": 41.572}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestDecoratedExportMutant", "wall": 2.947}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestAcceptanceMutantsPlanted_000", "wall": 5.208}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestAcceptanceMutantsPlanted_000", "wall": 5.058}
// {"exit": 0, "run": 1, "timing": [], "unit": "TestScalarEdges_003", "wall": 5.163}
// {"exit": 0, "run": 2, "timing": [], "unit": "TestScalarEdges_003", "wall": 5.128}
