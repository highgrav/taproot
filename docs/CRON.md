# Cronjobs

Taproot includes a basic job scheduling capability. You can add functions matching the `cron.CronJob` signature using 
standard cron notation, and the server will run them. Note that cronjobs are not durable between restarts, so a server 
will not run any "missed" jobs if it is down during a scheduled job time.

Cronjobs are not suitable for high-precision operations, but should be considered reliable for anything down to once a 
minute execution.


### Example
~~~
// adds a cronjob to run every minute
err = server.AddCronJob("ticker", "* * * * *", func() error {
		logging.LogToDeck(context.Background(), "info", "TICK", "info","tick tock")
		return nil
	})
	if err != nil {
		panic(err)
	}
~~~