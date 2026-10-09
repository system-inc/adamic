component='^(shard-[0-9]+|dedication|internal|oracle|testdata|weak|dedication[.]a|numbers[.]a|weak_narrowed[.]a|freed[.]a|probe_chain[.]a|reuse[.]a|narrowed[.]a)$'
bounded='^(TestWASIAgreesWithNode|TestWASIEmission|TestWeakReadsUndefinedOnceFreed)$'+'/'+'/'.join([component]*7)
