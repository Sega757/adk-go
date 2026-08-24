# Bolt's Performance Journal

## 2026-07-11 - Reflection-based deep copy allocation anti-pattern
**Learning:** Generic reflection-based cloning routines (like `clone` and `deepCopy` in Go) are highly flexible but can be extremely slow and allocation-heavy. Allocating temporary `reflect.Value` instances via `reflect.New` for every struct field or slice element is unnecessary when the destination struct is already fully allocated and its fields/elements are addressable/settable. Direct in-place copying bypasses these allocations completely. Additionally, maps with basic/scalar keys and values do not need deep copying of their keys or values at all, allowing us to skip allocations for map entries entirely.
**Action:** Avoid allocating new values with `reflect.New` for addressable sub-fields or slice elements when implementing generic copy functions. Always check whether the types of keys/values require deep copying (using a helper like `isCopyRequired`) before copying map entries.

## 2026-08-02 - JSON schema conversion and validation fast-path
**Learning:** General-purpose type conversions and schema validations that marshal/unmarshal Go values into raw JSON strings can be extremely slow and allocation-heavy, especially for simple values like strings, float64s, and booleans. When the source and destination types are identical and safe to pass directly to a JSON Schema validation engine, we can bypass the entire JSON serialization pipeline, reducing execution time and allocation overhead significantly. However, care must be taken to only bypass checks on safe primitive types (`string`, `float64`, `bool`, and untyped `nil`) and avoid assertions on typed nil pointers (which would cause runtime type assertion errors or incorrect validation bypass).
**Action:** Always implement a selective fast-path for JSON-safe types during validation and conversion routines to avoid JSON marshal/unmarshal rounds, but ensure other numeric types and pointers continue to fall back to standard deserialization pathways for correctness.

## 2026-08-15 - Schema Canonicalization Direct Buffer and Fast-Path
**Learning:** Canonicalizing deeply nested JSON structures (like JSON Schemas) via intermediate `json.Marshal` and recursive type serialization is an allocation-heavy task. We can reduce allocations and latency significantly by:
1) Writing directly to a single shared `*bytes.Buffer` instead of generating intermediate `[]byte` values at every recursion level.
2) Bypassing recursive sorting logic for empty or single-property maps.
3) Allocating string key slices onto the stack (using a stack-allocated backing array) for maps containing up to 16 keys (common for schemas).
4) Implementing a safe ASCII fast-path for string keys and string properties that writes standard strings directly without `json.Marshal`, fallback-safely avoiding JSON/HTML escape vulnerabilities.
**Action:** Apply single shared buffers, stack-allocated slices for small maps, and ASCII string fast-paths in any serialization/canonicalization loops.
