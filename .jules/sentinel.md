## 2025-02-18 - Path Traversal Vulnerability in NewAgentCardProvider

**Vulnerability:**
The `NewAgentCardProvider` function inside `agent/remoteagent/v2/a2a_agent.go` parses a string `source` to identify if it is a local file or an HTTP URL. If it's classified as a local file, it passes the user-provided `source` directly to `os.ReadFile(source)` without validating whether the string contains path traversal characters such as `..`. An attacker could supply an `AgentCardProvider` string that leads to arbitrary file reads (e.g. `../../../../etc/passwd`).

**Learning:**
Any function that accepts a file path (or a file source string) from potentially untrusted configurations must be verified to prevent escaping the intended directory boundaries, especially before passing the string to functions like `os.ReadFile` or `os.Open`. Even if a provider is intended to be used primarily from code or trusted config, it is critical to implement defense-in-depth and enforce strict path validation.

**Prevention:**
Validate file paths using `strings.Contains(source, "..")` and/or `strings.HasPrefix(filepath.Clean(source), "..")` before performing any filesystem operations. If a file is intended to be loaded from an arbitrary path on the disk by the local application, ensure that there is no possibility of untrusted user input directly dictating that path.
