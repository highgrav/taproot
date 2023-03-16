# WORKERS
Taproot provides a facility for executing asynchronous tasks through its `workers.WorkQueue` feature. To use 
asynchronous tasks, you register functions for acting on specified task types and acting on the return for specified 
task types. (These functions need to satisfy the `workers.WorkHandler` and `workers.ResultHandler` type signatures, 
respectively.)

Once registered, you can send task requests to `AppServer.StartWork()`. *When starting a task, make sure you are sending 
over the expected `Data` type, and that your `WorkHandler` can deal with unexpected type issues.*

Each executed task has a unique ID that can be tracked as needed, for logging, business logic, or notifications.

~~~
server.AddWorkHandlers("email-lead", func(wk *workers.WorkRequest) workers.WorkStatusReport {
		deck.Info("Adding work")
		title := wk.Data.(string)
		fmt.Println("Saw work of type " + wk.Type + ", id " + wk.ID + ": " + title)
		return workers.WorkStatusReport{
			Type:      wk.Type,
			ID:        wk.ID,
			Msg:       *wk,
			Status:    "done",
			StartedOn: time.Time{},
			EndedOn:   time.Time{},
			Messages: []string{
				"work complete",
			},
			Result: "sent email with title '" + title + "'",
			Error:  nil,
		}
	}, func(res workers.WorkStatusReport) {
		titleResult := res.Result.(string)
		deck.Info("Saw result from " + res.Type + " id: " + res.ID + ": " + titleResult)
	})
~~~