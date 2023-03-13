package workers

type IWorkerManager interface {
	Run(id string, fn func() error)
	Get(id string) (WorkStatusReport, error)
}
