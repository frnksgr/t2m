package t2m

import (
	"fmt"
	"net/http"
)

// TODO: update

const helpText = `
/help           this text

/healthz
    Healthendpoint

common parameters:
/<any action>?<parameters>
    size:       positive integer >= 1, number of requests
                defaults to 1

    topology:   tree|chain|fan
                defines topoplogy of request tree
                defaults to fan

    time:       task duration in milliseconds > 0
                defaults to 50
                
                leave actions:
    defaults to none
    
    /
    Do nothing

    /sleep	
    Sleep
    
    /fail
    Terminate TCP connection, do not send http response
    
    /crash
    Crash server process

    /cpu

    /ram
    
    example:
        curl "http://<domain:port>/fail?topology=fan&size=1000"
        
`

func handleHelp(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, helpText)
}
