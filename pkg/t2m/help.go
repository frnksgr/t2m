package t2m

import (
	"fmt"
	"net/http"
)

// TODO: update

const helpText = `
/help			this text

/healthz
	Healthendpoint

common parameters:
/<any action>?<parameters>
	count:      positive integer, number of requests
				defaults to 1

	topology:   btree|chain|fan
				defines topoplogy of requests
				defaults to fan

	target:		all|leaves
				selected nodes to execute action
				defaults to leave

	duration:   milliseconds > 0
				duration of task execution
				defaults to 50
				
				leave actions:
	defaults to none
	
	/none
	Do nothing
	
	/fail
	Terminate TCP connection, do not send http response
	
	/crash
	Crash server process
	
	example:
		curl "http://<domain:port>/fail?topology=fan&count=1000"
		
`

func handleHelp(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, helpText)
}
