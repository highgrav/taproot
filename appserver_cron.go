package taproot

import "github.com/highgrav/taproot/cron"

func (srv *AppServer) AddCronJob(name, schedule string, job cron.CronJob) error {
	if srv.CronHub == nil {
		srv.CronHub = cron.New()
	}
	return srv.CronHub.AddJob(name, schedule, job)
}

func (srv *AppServer) RemoveCronJob(name string) {
	if srv.CronHub == nil {
		return
	}
	srv.CronHub.RemoveJob(name)
}
