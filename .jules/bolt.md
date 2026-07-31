# Bolt's Performance Journal

## 2026-07-11 - Reflection-based deep copy allocation anti-pattern
**Learning:** Generic reflection-based cloning routines (like `clone` and `deepCopy` in Go) are highly flexible but can be extremely slow and allocation-heavy. Allocating temporary `reflect.Value` instances via `reflect.New` for every struct field or slice element is unnecessary when the destination struct is already fully allocated and its fields/elements are addressable/settable. Direct in-place copying bypasses these allocations completely. Additionally, maps with basic/scalar keys and values do not need deep copying of their keys or values at all, allowing us to skip allocations for map entries entirely.
**Action:** Avoid allocating new values with `reflect.New` for addressable sub-fields or slice elements when implementing generic copy functions. Always check whether the types of keys/values require deep copying (using a helper like `isCopyRequired`) before copying map entries.

## 2026-07-12 - Repetitive JSON marshal/unmarshal cycle during schema validation
**Learning:** Performing `json.Marshal` followed by multiple `json.Unmarshal` calls to convert a Go value to its standard decoded JSON form (`any`) for schema validation is extremely expensive and allocation-heavy. When values are already standard JSON primitive types (like `float64`, `string`, `bool`, or untyped `nil`), they can bypass this serialization/deserialization cycle entirely without any safety risk or data races.
**Action:** Implement helper checks (e.g. `isJSONSafe`) to identify primitive JSON-safe types, and bypass standard marshal/unmarshal pipelines for validation and direct type assertion conversions wherever applicable.
