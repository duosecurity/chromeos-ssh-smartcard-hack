
package main

import (
	"log"

	"github.com/gopherjs/gopherjs/js"
	"golang.org/x/crypto/ssh/agent"
)


func main() {
	var backend Backend
	// if platformKeysSupported() {
	backend = NewCertificateProviderBackend()
	// } else {
	// 	var err error
	// 	backend, err = NewChromeStorageBackend()
	// 	if err != nil {
	// 		log.Printf("Failed to create ChromeStorageAgent: %v", err)
	// 		return
	// 	}
	// }
	launch(NewAgent(backend))
}

func launch(mga *Agent) {
	js.Global.Set("agent", js.MakeWrapper(mga))

	log.Printf("Starting agent")
	js.Global.Get("chrome").
		Get("runtime").
		Get("onConnectExternal").
		Call("addListener", func(port *js.Object) {
		p := NewAgentPort(port)
		go agent.ServeAgent(mga, p)
	})
}
