package services

type IStatsService interface {
	GetStats(username string) (string, error)
}
