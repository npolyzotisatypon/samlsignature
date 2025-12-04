package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"flag"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"

	"github.com/beevik/etree"
	goxmldsig "github.com/russellhaering/goxmldsig"
)

var (
	certFile    = flag.String("cert", "", "[required] Path to certificate file")
	postURL     = flag.String("postURL", "https://accounts.sap.com/saml2/idp/sso", "IdP's Post URL")
	wayflessURL = flag.String("wayflessURL", "https://dl.acm.org/action/ssostart?idp=https://accounts.sap.com", "Wayfless URL")
	samlData    []byte
)

const (
	dateFormat             = "2006-01-02"
	browserTimeout         = 30 * time.Second
	navigationTimeout      = 10 * time.Second
	remoteDebuggingAddress = "ws://127.0.0.1:9222"
)

func main() {
	today := time.Now().Format(dateFormat)
	dateStr := flag.String("date", today, "Date to use for validation (YYYY-MM-DD)")
	flag.Parse()

	// Check required flags
	if *certFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Load certificate
	certData, err := os.ReadFile(*certFile)
	if err != nil {
		log.Fatalf("Error reading certificate file: %v", err)
	}

	// Parse certificate
	cert := validateCertificate(certData)

	// Create remote allocator context
	allocatorContext, cancelAllocator := chromedp.NewRemoteAllocator(context.Background(), remoteDebuggingAddress)
	defer cancelAllocator()

	// Create Chrome context with custom logger
	ctx, cancel := chromedp.NewContext(allocatorContext, chromedp.WithLogf(log.Printf))
	defer cancel()

	// Set a timeout for our operations
	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, browserTimeout)
	defer cancelTimeout()

	var samlOnce sync.Once
	samlReady := make(chan struct{})
	// Listen for network events
	chromedp.ListenTarget(timeoutCtx, func(ev interface{}) {
		switch e := ev.(type) {
		case *network.EventRequestWillBeSent:
			if e.Request.Method == "POST" && e.Request.URL == *postURL {
				log.Printf("POST request to: %s\n", e.Request.URL)

				if len(e.Request.PostDataEntries) == 0 {
					log.Printf("No POST data entries found")
					return
				}

				encodedData := string(e.Request.PostDataEntries[0].Bytes)
				decodedData, err := base64.StdEncoding.DecodeString(encodedData)
				if err != nil {
					log.Printf("Error decoding POST data: %v", err)
					return
				}

				postParams, err := url.ParseQuery(string(decodedData))
				if err != nil {
					log.Printf("Error parsing POST parameters: %v", err)
					return
				}

				samlRequestEncode := postParams.Get("SAMLRequest")
				var decodeErr error
				samlData, decodeErr = base64.StdEncoding.DecodeString(samlRequestEncode)
				if decodeErr != nil {
					log.Printf("Error decoding SAML request: %v", decodeErr)
					return
				}
				log.Printf("SAMLRequest: %s\n", samlData)

				// Signal completion only once, even if multiple POST requests occur
				samlOnce.Do(func() {
					close(samlReady)
				})
			}
		}
	})
	// Enable network events
	if err := chromedp.Run(timeoutCtx, network.Enable()); err != nil {
		log.Fatalf("Failed to enable network events: %v", err)
	}

	// Navigate to target page and wait for SAML request
	if err := chromedp.Run(timeoutCtx,
		chromedp.Navigate(*wayflessURL),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Wait for navigation or response
			select {
			case <-samlReady:
				// SAML request captured successfully
			case <-time.After(navigationTimeout):
				log.Printf("Timeout waiting for SAML request after %v\n", navigationTimeout)
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		}),
	); err != nil {
		log.Fatalf("Failed to navigate or capture SAML request: %v", err)
	}

	if len(samlData) == 0 {
		log.Fatal("SAMLRequest is empty or not captured")
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(samlData); err != nil {
		log.Fatalf("Error parsing XML: %v", err)
	}
	// Build validation context
	validationContext := &goxmldsig.ValidationContext{
		IdAttribute: "ID",
		CertificateStore: &goxmldsig.MemoryX509CertificateStore{
			Roots: []*x509.Certificate{cert},
		},
	}

	startDate, err := time.Parse(dateFormat, *dateStr)
	if err != nil {
		log.Fatalf("Error parsing date: %v", err)
	}

	validationContext.Clock = goxmldsig.NewFakeClockAt(startDate)
	_, err = validationContext.Validate(doc.Root())
	if err != nil {
		log.Fatalf("Signature validation failed: %v", err)
	}

	log.Println("Signature validation successful!")
}

func validateCertificate(certData []byte) *x509.Certificate {
	certBlock, rest := pem.Decode(certData)
	if certBlock == nil {
		log.Fatal("Failed to parse certificate PEM data")
	}
	if len(rest) > 0 {
		log.Printf("Warning: Extra data after PEM block (length: %d bytes)", len(rest))
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		log.Fatalf("Failed to parse certificate: %v", err)
	}

	// Verify certificate contains an RSA public key
	_, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		log.Fatal("Certificate doesn't contain an RSA public key")
	}
	return cert
}
