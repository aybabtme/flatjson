# WIP

## What is flat JSON?

Flat JSON is a subset of JSON where the only supported types are objects containing
strings, numbers, booleans or null values. There can't be nested objects or
arrays. The root element must be an object.

## What's the use for that?

If you log in JSON, likely your logs respect this principle. Using a JSON
parser that supports only this subset should be faster than using a general
purpose one. So this is one use case, parsing logs that are in JSON.

## Implementation

This is a WIP implementation of a flatjson parser.

## Speed

Comparing this implementation with the standard library JSON decoder. Both have
their output ignored:

- flajson's name/value pairs are ignored.
- encoding/json is decoding into an empty struct.

The goal here is to see how fast only the decoding part is. This is not necessarly
a characteristic workload since a normal use case would allocate memory for the
strings of the name/value pairs.

```
BenchmarkFlatJSON         1000000         1970 ns/op     177.15 MB/s           0 B/op          0 allocs/op
BenchmarkEncodingJSON     100000         20962 ns/op      36.83 MB/s        2151 B/op        130 allocs/op
```

At this time, the API of flatjson is not nailed down, so I haven't benchmarked a real
use case. However, the benchmark above at least demonstrates that the potential
for greatly improved speed is there.
