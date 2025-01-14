package services

// IStatsService defines the contract for stats-related operations.
type IStatsService interface {
	// GetStats calculates the user's performance relative to others.
	GetStats(username string) (string, error)
}
