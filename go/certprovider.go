package main

import (

	"github.com/gopherjs/gopherjs/js"
)

// CertificateProvider is a wrapper for directly calling
// CertificateProvider event listener methods (currently, those
// provided by the Charismathics Smart Card Middleware extension) that
// handles making the async API synchronous
type CertificateProvider struct {
}

func (cp * CertificateProvider) ListClientCertificates() (matches [][]byte, err error) {
	// Uncaught exceptions in JS get translated into panics in Go
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	results := make(chan []*js.Object, 1)

	// find the event listener provided to Chrome via
	// chrome.certificateProvider.onCertificatesRequested.addListener(),
	// and then call that thing directly.
	// XXX or maybe we could actually just do
	// XXX chrome.certificateProvider.onCertificatesRequested.dispatch() ?
	backend := js.Global.Get("$jscomp").Get("scope").Get("certificateProviderBridgeBackend")
	backend.Call("boundCertificatesRequestListener_", func(matches []*js.Object) {
		go func() { results <- matches }()
	})

	objects := <-results
	for _, obj := range objects {
		cert := obj.Get("certificate")
		matches = append(matches, js.Global.Get("Uint8Array").New(cert).Interface().([]byte))
	}	
	return
}

func (cp *CertificateProvider) Sign(request js.M) (sig []byte, err error) {
	// Uncaught exceptions in JS get translated into panics in Go
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	result := make(chan []byte);

	// find the event listener provided to chrome via
	// chrome.certificateProvider.onSignDigestRequested.addListener(),
	// and then call that thing directly.
	backend := js.Global.Get("$jscomp").Get("scope").Get("certificateProviderBridgeBackend")
	backend.Call("boundSignDigestRequestListener_", request, func(sig *js.Object) {
		go func() { result <- js.Global.Get("Uint8Array").New(sig).Interface().([]byte) }()
	})
	sig = <-result 
	return
}
