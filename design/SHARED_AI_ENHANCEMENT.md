# Shared enhancement client

OpenAIClient delegates request construction, HTTP handling, error decoding,
model pricing, and retries to packages/ai. The circuit breaker still wraps one
logical enhancement call and the server returns the original prompt on an open
breaker or any terminal failure. Empty choices and empty content also retain
that fallback.

The existing request remains GPT-4, seed 1337, temperature 0.2, and one system
message containing the supplied prompt. The authorType argument remains unused.
There is no new token limit or model substitution. The shared registry now
contains GPT-4 with verified pricing; unsupported model configuration falls back
without sending a request.

The library owns the only AI retry loop. OpenAIRetryConfig still supplies three
total attempts and the existing exponential delay with jitter through RetryDelay.
RetryableHTTPOperation is removed; RetryWithBackoff remains for file operations.
The shared classifier avoids retrying permanent failures and exhausted quota.
Each attempt retains the 60-second deadline. The overall timeout accommodates
the configured attempts and maximum backoff; parent cancellation ends requests
and waits. EnhancePromptWithContext accepts that parent context, and the existing
EnhancePromptWithResilience entry point remains compatible.

OnAttempt updates the existing metrics once per attempted request, including
transport, provider, and malformed-response failures. Retry counts increase only
when another attempt actually happens, not for countdown notifications. The
caller request ID is sent on every retry and retained in metric logs; shared
API errors also expose the provider request ID and code through errors.As.
Timeout counters cover gateway timeouts and wrapped deadline/transport timeouts.
Empty successful responses now count as successful API attempts even though the
server falls back to the original prompt; previously those attempts were omitted.
The breaker records one outcome for the whole call.

Reported text usage is recorded once through ai.RecordCall with tool dalleserver,
including a completed empty-response fallback with usage. Missing usage and
failed or blocked requests do not fabricate ledger rows. Ledger failures are
logged without discarding successful enhancement.

Tests inject HTTP transports and isolated collectors/breakers and ledger paths.
They exercise retries, exact metrics, cancellation, fallback, and accounting
without paid API calls. The module declares its shared-ai dependency; build
targets and deployment configuration remain unchanged.
