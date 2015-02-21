# WIP

## What is flat JSON?

Flat JSON is a subset of JSON where the only support types are objects containing
strings, numbers, booleans or null values. There can't be nested objects or
arrays. The root element must be an object.

## What's the use for that?

If you log in JSON, likely your logs respect this principle. Using a JSON
parser that supports only this subset should be faster than using a general
purpose one.
