package executor

type Limiter interface {
	LimitProcess(pid int) error
}
