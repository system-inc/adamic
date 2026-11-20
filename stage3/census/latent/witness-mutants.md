# Deferred witness test mutations

Run each named test with its mutation enabled. An ordinary run leaves all
mutation variables unset.

| Mutation | Environment variable | Test |
| --- | --- | --- |
| Remove the recorded body refusal from the fixture result | `WITNESS_TEST_MUTANT_DROP_STOP=1` | `test_latent_witness_hidden_body_forecasts` |
| Accept a type argument containing a type parameter | `WITNESS_TEST_MUTANT=LATENT_MUTANT_WITNESS_OPEN_TYPE` | `test_latent_witness_rejects_open_type_arguments` |
| Enable witness execution by default | `WITNESS_TEST_MUTANT=LATENT_MUTANT_WITNESS_DEFAULT` | `test_latent_witness_default_matches_previous_census` |
| Append a witness stop to its caller's census findings | `WITNESS_TEST_MUTANT=LATENT_MUTANT_WITNESS_FINDING` | `test_latent_witness_preserves_census_and_meter` |
| Append a witness unit to the timing stream | `WITNESS_TEST_MUTANT=LATENT_MUTANT_WITNESS_TIMING` | `test_latent_witness_preserves_census_and_meter` |
| Lower the witness on the shared state, leaving a failed instantiation cached for a later unit | `LATENT_MUTANT_WITNESS_SHARED_STATE=1` | `TestLatentWitnessSnapshotPreservesLaterCallResult` |
| Reset the caller at every function-like node | `WITNESS_TEST_MUTANT=LATENT_MUTANT_WITNESS_RESET_OWNER` | `test_latent_witness_uses_nearest_enclosing_census_unit` |
| Append a witness finding during a tail-on census | `WITNESS_TEST_MUTANT=LATENT_MUTANT_WITNESS_FINDING` | `test_latent_witness_preserves_scoped_tail_ledger` |
| Remove the exact generic argument selection anchor | Source reshape in the overlay anchor guard | `TestMakeOverlayTextEditsNameAReshape` |
