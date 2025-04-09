package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"

	"github.com/beevik/etree"
	goxmldsig "github.com/russellhaering/goxmldsig"
)

var (
	certFile  = flag.String("cert", "", "[required] Path to certificate file")
	//debugFlag = flag.Bool("debug", false, "Enable debug output")
	postURL = flag.String("postURL", "https://accounts.sap.com/saml2/idp/sso", "IdP's Post URL")
	wayflessURL = flag.String("wayflessURL", "https://dl.acm.org/action/ssostart?idp=https://accounts.sap.com", "Wayfless URL")
	samlData  []byte
)

const (
	dateFormat = "2006-01-02"
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

	allocatorContext, _ := chromedp.NewRemoteAllocator(context.Background(), "ws://127.0.0.1:9222")
	// also set up a custom logger
	ctx, cancel := chromedp.NewContext(allocatorContext, chromedp.WithLogf(log.Printf))
	defer cancel()

	// Set a timeout for our operations
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	samlReady := make(chan struct{})
	// Listen for network events
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *network.EventRequestWillBeSent:
			if e.Request.Method == "POST" && e.Request.URL == *postURL {
				fmt.Printf("POST request to: %s\n", e.Request.URL)

				// postData, _ := json.MarshalIndent(e.Request.PostDataEntries, "", "  ")

				encodedData := string(e.Request.PostDataEntries[0].Bytes)
				decodedData, _ := base64.StdEncoding.DecodeString(encodedData)
				postParams, _ := url.ParseQuery(string(decodedData))

				samlRequestEncode := postParams.Get("SAMLRequest")
				samlData, _ = base64.StdEncoding.DecodeString(samlRequestEncode)
				fmt.Printf("SAMLRequest: %s\n", samlData)
				samlReady <- struct{}{}
				// fmt.Printf("POST payload: %s\n", postData)
				// Print request headers
				// headers, _ := json.MarshalIndent(e.Request.Headers, "", "  ")
				// fmt.Printf("Headers: %s\n\n", headers)
			}
		}
	})
	// Enable network events
	if err := chromedp.Run(ctx, network.Enable()); err != nil {
		log.Fatal(err)
	}

	// Navigate to your target page and interact with it
	if err := chromedp.Run(ctx,
		chromedp.Navigate(*wayflessURL),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Wait for navigation or response
			select {
			case <-samlReady:
				close(samlReady)
			case <-time.After(10 * time.Second):
				fmt.Printf("Timeout waiting for navigation or response in ActionFunc\n")
			}
			return nil
		}),
	); err != nil {
		log.Fatal(err)
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

	startDate := Must(time.Parse(dateFormat, *dateStr))
	validationContext.Clock = goxmldsig.NewFakeClockAt(startDate)
	_, err = validationContext.Validate(doc.Root())
	if err != nil {
		// avoid fatal error
		log.Printf("Signature validation failed: %v", err)
		return
	}

	fmt.Println("Signature validation successful!")
}

func validateCertificate(certData []byte) *x509.Certificate {
	var certBlock *pem.Block
	certBlock, _ = pem.Decode(certData)
	if certBlock == nil {
		log.Fatal("Failed to parse certificate PEM data")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		log.Fatalf("Failed to parse certificate: %v", err)
	}

	// Extract public key
	_, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		log.Fatal("Certificate doesn't contain an RSA public key")
	}
	return cert
}

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}
