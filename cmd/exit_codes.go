package cmd

const (
	ExitSuccess      = 0 // Successful execution
	ExitError        = 1 // Generic error
	ExitUsage        = 2 // Invalid usage / flags
	ExitAuth         = 3 // Authentication error
	ExitAPIError     = 4 // API error (4xx)
	ExitNetworkError = 5 // Network / server error (5xx, timeout)
	ExitSafetyLimit  = 6 // Safety limit exceeded
)
