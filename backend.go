package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"crypto/rand"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func run(args []string) int {
	bindAddress := flag.String("ip", "0.0.0.0", "IP address to bind")                                                          //Default to localhost
	listenPort := flag.Int("port", 25478, "port number to listen on")                                                          //Default to 25478
	tlsListenPort := flag.Int("tlsport", 25443, "port number to listen on with TLS")                                           //Default to 25443
	maxUploadSize := flag.Int64("upload_limit", 5242880*3, "max size of uploaded file (byte)")                                 // Default max of 15 mb
	tokenFlag := flag.String("token", "f9403fc5f537b4ab332d", "specify the security token")                                    //Default to a random token
	protectedMethodFlag := flag.String("protected_method", "", "specify methods intended to be protect by the security token") //Default to none
	logLevelFlag := flag.String("loglevel", "info", "logging level")                                                           //Default to info
	certFile := flag.String("cert", "", "path to certificate file")                                                            //Default to no TLS
	keyFile := flag.String("key", "", "path to key file")                                                                      //Default to no TLS
	corsEnabled := flag.Bool("cors", true, "if true, add ACAO header to support CORS")                                         //Default to add CORS header
	serverRoot := flag.String("", "./dumps", "Where to save the dumps")

	if logLevel, err := logrus.ParseLevel(*logLevelFlag); err != nil {
		logrus.WithError(err).Error("failed to parse logging level, so set to default")
	} else {
		logger.Level = logLevel
	}
	token := *tokenFlag
	if token == "" {
		count := 10
		b := make([]byte, count)
		if _, err := rand.Read(b); err != nil {
			logger.WithError(err).Fatal("could not generate token")
			return 1
		}
		token = fmt.Sprintf("%x", b)
		logger.WithField("token", token).Warn("token generated")
	}
	protectedMethods := []string{}
	for _, method := range strings.Split((*protectedMethodFlag), ",") {
		if strings.EqualFold("PUT", method) {
			protectedMethods = append(protectedMethods, http.MethodPut)
		} else if strings.EqualFold("OPTIONS", method) {
			protectedMethods = append(protectedMethods, http.MethodOptions)
		}
	}
	tlsEnabled := *certFile != "" && *keyFile != ""
	server := NewServer(*serverRoot, *maxUploadSize, token, *corsEnabled, protectedMethods)
	http.Handle("/files/", server) //This is going to call with a goroutine so we can handle multiple requests at once

	errors := make(chan error)

	go func() {
		logger.WithFields(logrus.Fields{
			"ip":               *bindAddress,
			"port":             *listenPort,
			"token":            token,
			"protected_method": protectedMethods,
			"upload_limit":     *maxUploadSize,
			"root":             serverRoot,
			"cors":             *corsEnabled,
		}).Info("start listening")

		if err := http.ListenAndServe(fmt.Sprintf("%s:%d", *bindAddress, *listenPort), nil); err != nil {
			errors <- err
		}
	}()

	if tlsEnabled {
		go func() {
			logger.WithFields(logrus.Fields{
				"cert": *certFile,
				"key":  *keyFile,
				"port": *tlsListenPort,
			}).Info("start listening TLS")

			if err := http.ListenAndServeTLS(fmt.Sprintf("%s:%d", *bindAddress, *tlsListenPort), *certFile, *keyFile, nil); err != nil {
				errors <- err
			}
		}()
	}

	err := <-errors
	logger.WithError(err).Info("closing server")

	return 0
}

// EntryPoint
func main() {
	logger = logrus.New()
	logger.Info("starting up analyzer server")
	//Make sure we have our dump folder
	if _, err := os.Stat("./dumps"); os.IsNotExist(err) {
		os.Mkdir("./dumps", 0777)
	}
	//Clear out any old files in dump folder
	os.RemoveAll("./dumps/*")

	result := run(os.Args)
	os.Exit(result)
}
