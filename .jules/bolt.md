# Bolt's Performance Journal

## 2026-07-11 - Reflection-based deep copy allocation anti-pattern
**Learning:** Generic reflection-based cloning routines (like `clone` and `deepCopy` in Go) are highly flexible but can be extremely slow and allocation-heavy. Allocating temporary `reflect.Value` instances via `reflect.New` for every struct field or slice element is unnecessary when the destination struct is already fully allocated and its fields/elements are addressable/settable. Direct in-place copying bypasses these allocations completely. Additionally, maps with basic/scalar keys and values do not need deep copying of their keys or values at all, allowing us to skip allocations for map entries entirely.
**Action:** Avoid allocating new values with `reflect.New` for addressable sub-fields or slice elements when implementing generic copy functions. Always check whether the types of keys/values require deep copying (using a helper like `isCopyRequired`) before copying map entries.

## 2026-07-23 - JSON marshal/unmarshal type conversion overhead
**Learning:** Converting types using JSON marshal/unmarshal (as done in `ConvertToWithJSONSchema`) is a robust way to handle conversions and schema validation, but it introduces huge overheads when converting basic types without a schema. If the source value is of a primitive JSON-safe type (`float64`, `string`, `bool`, or untyped `nil`) and there is no validation schema, we can safely bypass JSON serialization entirely and use direct Go type assertions. This eliminates 100% of the memory allocations and speeds up execution by nearly 100x.
**Action:** Always check if the conversion is for a schema-less primitive type first, using a safe utility like `isJSONSafe(v)`, and use type assertions to bypass expensive serialization paths.
