## 2024-05-18 - [Denial of Service via Unbuffered Channel]
**Vulnerability:** A non-positive `MaxConcurrentRuns` in `TriggerConfig` resulted in the creation of an unbuffered channel (`make(chan struct{}, 0)`) in `eventarc.go`. This causes all incoming requests to permanently block when attempting to send to the channel (`c.semaphore <- struct{}{}`), resulting in a Denial of Service (DoS) and goroutine leak.
**Learning:** `MaxConcurrentRuns <= 0` signifies unlimited concurrency in the application's configuration model and requires leaving the `semaphore` channel nil, as handled correctly in `pubsub.go`.
**Prevention:** Consistently apply the nil check or default logic for unlimited concurrency patterns across similar handlers (Eventarc vs Pub/Sub).
