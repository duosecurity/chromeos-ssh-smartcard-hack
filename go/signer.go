package main

import (
	"crypto"
	"crypto/elliptic"
	"crypto/x509"
	"errors"
	"io"

	"github.com/gopherjs/gopherjs/js"
)

var ErrUnsupportedHash = errors.New("unsupported hash")

type CPSigner struct {
	cp *CertificateProvider
	cert *x509.Certificate
}

func NewCPSigner(cp *CertificateProvider, cert *x509.Certificate) crypto.Signer {
	return &CPSigner{
		cp: cp,
		cert: cert,
	}
}

func (cps *CPSigner) Public() crypto.PublicKey {
	return cps.cert.PublicKey
}

// Limited to just the hashes supported by WebCrypto
var hashPrefixes = map[crypto.Hash][]byte{
	crypto.SHA1:    {0x30, 0x21, 0x30, 0x09, 0x06, 0x05, 0x2b, 0x0e, 0x03, 0x02, 0x1a, 0x05, 0x00, 0x04, 0x14},
	crypto.SHA256:  {0x30, 0x31, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x01, 0x05, 0x00, 0x04, 0x20},
	crypto.SHA384:  {0x30, 0x41, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x02, 0x05, 0x00, 0x04, 0x30},
	crypto.SHA512:  {0x30, 0x51, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x03, 0x05, 0x00, 0x04, 0x40},
	crypto.Hash(0): {}, // Special case in the golang interface to indicate that data is signed directly
}

var hashNames = map[crypto.Hash]string{
	crypto.SHA1:    "SHA1",
	crypto.SHA256:  "SHA256",
	crypto.SHA384:  "SHA384",
	crypto.SHA512:  "SHA512",
	crypto.Hash(0): "none",
}

var curveNames = map[elliptic.Curve]string{
	elliptic.P256(): "P-256",
	elliptic.P384(): "P-384",
	elliptic.P521(): "P-521",
}

func (cps *CPSigner) Sign(rand io.Reader, msg []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	hash := hashNames[opts.HashFunc()]

	signRequest := js.M{
		"digest": js.NewArrayBuffer(msg),
		"hash": hash,
		"certificate": js.NewArrayBuffer(cps.cert.Raw),
	}
	return cps.cp.Sign(signRequest)
}
