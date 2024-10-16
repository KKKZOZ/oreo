package timesource

type TimeSourcer interface {
	GetTime(mode string) (int64, error)
}
