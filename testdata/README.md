# Test data directory

This directory contains the generated test data obtained from running go test on the [_testsource](../_testsource) directory.

The `*.txt` files contain the raw output from the tests. These are used by the **tokenizer**, and the tests compare the tokenizer output with the `*.tokenizer.json` files.

The `*.tokenizer.json` files also double as input for the **parser**. The output of the parser is then compared with the `*.parser.json` files. 