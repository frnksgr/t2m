package t2m

import (
	"fmt"
	"net/http"
)

const helpText = `
/<query-paramters>
parameters:
	
	count:      positive integer, number of requests
				defaults to 1

	topology:   binary|linear|flat 
				defines topoplogy of requests
				defaults to binary

example:
	curl "http://<domain:port>/?topology=fan&count=1000"


/fail
	Terminate TCP connection, do not send http response


/crash
	Crash server process


/healthz
	Healthendpoint
`

func handleHelp(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, helpText)
}
