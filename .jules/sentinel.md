# Sentinel Security Ledger

## 2026-11-07 - Denial of Service via Zero-Capacity Concurrency Semaphore in Eventarc Trigger

- **Vulnerability**: In `server/adkrest/controllers/triggers/eventarc.go`, `NewEventarcControllerWithConfig` unconditionally initialized `semaphore` with `make(chan struct{}, cfg.TriggerConfig.MaxConcurrentRuns)`. When `MaxConcurrentRuns <= 0` (e.g. the default zero value intended to indicate unlimited concurrent runs), Go created a non-nil unbuffered channel (`make(chan, 0)`). Consequently, `EventarcTriggerHandler`'s nil check `if c.semaphore != nil` passed and attempted to send to the unbuffered channel (`c.semaphore <- struct{}{}`) without any concurrent receiver, permanently blocking every incoming HTTP request, leaking goroutines, and causing a complete service denial-of-service.
- **Learning**: Semaphore channels representing concurrency limits must be kept `nil` when concurrency is unlimited (`MaxConcurrentRuns <= 0`). Unbuffered channel instantiation `make(chan T, 0)` produces a non-nil channel that blocks sends until a receiver is ready, converting concurrency limit checks into permanent deadlocks when receivers do not exist.
- **Prevention**: Guard semaphore channel instantiation with `if cfg.TriggerConfig.MaxConcurrentRuns > 0` and verify with a negative regression test proving that `MaxConcurrentRuns = 0` runs immediately without blocking.
