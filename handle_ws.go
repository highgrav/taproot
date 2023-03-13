package taproot

/*
func (srv *AppServer) HandleWS(brokerName string, autoTimeoutMinutes int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			deck.Error("HandleWS() could not upgrade ws request: " + err.Error())
			srv.ErrorResponse(w, r, 500, "could not upgrade ws request")
			return
		}
		broker, ok := srv.WSHubs[brokerName]
		user := srv.GetUserFromRequest(r)
		if !ok {
			// return error
			srv.ErrorResponse(w, r, 500, "ws message source not available")
			return
		}
		wsc := websock.WSConn{
			User:   user,
			Conn:   conn,
			Close:  make(chan bool),
			Reader: make(chan []byte),
			Writer: make(chan []byte),
		}

	}
}
*/
