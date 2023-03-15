package workers

type WorkHandler func(msg *WorkRequest) WorkStatusReport

type ResultHandler func(res WorkStatusReport)
