# Checked-cast fixture counts

Measured by `TestScannerCastCounts` on the complete delivery group; these 20 rows are also recorded in the canonical `internal/oracle/counts.md`. Positive programs finish cleanup; failing twins are counted where they stop 70, and cleanup after panic is not claimed. The required global refresh remains blocked by 39 inherited fixture failures.

| Fixture | Allocations | Frees | Retains | Releases | Peak live | In regions | Graph regions | Graph merges |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 01_enum_declaration.a | 4 | 4 | 3 | 6 | 4 | 0 | 0 | 0 |
| 01_enum_declaration_fails.a | 2 | 0 | 2 | 1 | 2 | 0 | 0 | 0 |
| 02_name_subunion.a | 2 | 2 | 4 | 4 | 2 | 0 | 0 | 0 |
| 02_name_subunion_fails.a | 1 | 0 | 2 | 0 | 1 | 0 | 0 | 0 |
| 03_chain_info.a | 3 | 3 | 2 | 5 | 3 | 0 | 0 | 0 |
| 03_chain_info_fails.a | 1 | 0 | 3 | 1 | 1 | 0 | 0 | 0 |
| 04_literal_payload.a | 3 | 3 | 7 | 10 | 2 | 0 | 0 | 0 |
| 04_literal_payload_fails.a | 3 | 1 | 5 | 5 | 2 | 0 | 0 | 0 |
| 05_extension.a | 2 | 2 | 5 | 5 | 2 | 0 | 0 | 0 |
| 05_extension_fails.a | 2 | 0 | 4 | 1 | 2 | 0 | 0 | 0 |
| 06_parser_keyword.a | 4 | 4 | 3 | 5 | 3 | 0 | 0 | 0 |
| 06_parser_keyword_fails.a | 3 | 0 | 3 | 1 | 3 | 0 | 0 | 0 |
| 07_generic_next.a | 8 | 8 | 5 | 13 | 4 | 0 | 0 | 0 |
| 07_generic_next_fails.a | 5 | 4 | 5 | 8 | 4 | 0 | 0 | 0 |
| 08_generic_node.a | 2 | 2 | 2 | 5 | 2 | 0 | 0 | 0 |
| 08_generic_node_fails.a | 1 | 0 | 1 | 2 | 1 | 0 | 0 | 0 |
| 09_primitive_string.a | 1 | 1 | 1 | 2 | 1 | 0 | 0 | 0 |
| 09_primitive_string_fails.a | 2 | 0 | 0 | 0 | 2 | 0 | 0 | 0 |
| 10_generator_label.a | 4 | 4 | 1 | 5 | 3 | 0 | 0 | 0 |
| 10_generator_label_fails.a | 2 | 0 | 2 | 0 | 2 | 0 | 0 | 0 |
