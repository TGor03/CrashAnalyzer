package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"

	"github.com/sirupsen/logrus"
)

var (
	rePathFiles = regexp.MustCompile(`^/files/([^/]+)$`)

	errTokenMismatch = errors.New("token mismatched")
	errMissingToken  = errors.New("missing token")
)

// Server represents our upload server.
type Server struct {
	DocumentRoot string
	// MaxUploadSize limits the size of the uploaded content, specified with "byte".
	MaxUploadSize    int64
	SecureToken      string
	EnableCORS       bool
	ProtectedMethods []string
}

// NewServer creates a new upload server.
func NewServer(documentRoot string, maxUploadSize int64, token string, enableCORS bool, protectedMethods []string) Server {
	return Server{
		DocumentRoot:     documentRoot,
		MaxUploadSize:    maxUploadSize,
		SecureToken:      token,
		EnableCORS:       enableCORS,
		ProtectedMethods: protectedMethods,
	}
}

// Handle PUT requests
func (s Server) handlePut(w http.ResponseWriter, r *http.Request) {
	if s.EnableCORS {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}

	matches := rePathFiles.FindStringSubmatch(r.URL.Path)
	if matches == nil {
		logger.WithField("path", r.URL.Path).Info("invalid path")
		w.WriteHeader(http.StatusNotFound)
		writeError(w, fmt.Errorf("\"%s\" is not found", r.URL.Path))
		return
	}
	targetPath := path.Join(s.DocumentRoot, matches[1])

	// We have to create a new temporary file in the same device to avoid "invalid cross-device link" on renaming.
	// Here is the easiest solution: create it in the same directory.
	tempFile, err := os.CreateTemp(s.DocumentRoot, "upload_")
	if err != nil {
		logger.WithError(err).Error("failed to create a temporary file")
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	}
	defer r.Body.Close()
	srcFile, info, err := r.FormFile("file")
	if err != nil {
		logger.WithError(err).WithField("path", targetPath).Error("failed to acquire the uploaded content")
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	}
	defer srcFile.Close()
	// dump headers for the file
	logger.Debug(info.Header)

	size, err := getSize(srcFile)
	if err != nil {
		logger.WithError(err).WithField("path", targetPath).Error("failed to get the size of the uploaded content")
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	}
	if size > s.MaxUploadSize {
		logger.WithFields(logrus.Fields{
			"path": targetPath,
			"size": size,
		}).Info("file size exceeded")
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		writeError(w, errors.New("uploaded file size exceeds the limit"))
		return
	}

	n, err := io.Copy(tempFile, srcFile)
	if err != nil {
		logger.WithError(err).WithField("path", tempFile.Name()).Error("failed to write body to the file")
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	}
	// excplicitly close file to flush, then rename from temp name to actual name in atomic file
	tempFile.Close()
	targetPath = tempFile.Name() + ".dmp"
	if err := os.Rename(tempFile.Name(), targetPath); err != nil {
		os.Remove(tempFile.Name())
		logger.WithError(err).WithField("path", targetPath).Error("failed to rename temp file to final filename for upload")
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	}

	logger.WithFields(logrus.Fields{
		"path": r.URL.Path,
		"size": n,
	}).Info("file uploaded by PUT")
	w.Write([]byte(analyzedump(targetPath)))
}

func (s Server) handleOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, PUT, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.WriteHeader(http.StatusNoContent)
}

func (s Server) checkToken(r *http.Request) error {
	// first, try to get the token from the query strings
	token := r.URL.Query().Get("token")
	// if token is not found, check the form parameter.
	if token == "" {
		token = r.FormValue("token")
	}
	if token == "" {
		return errMissingToken
	}
	if token != s.SecureToken {
		return errTokenMismatch
	}
	return nil
}

func (s Server) isAuthenticationRequired(r *http.Request) bool {
	for _, m := range s.ProtectedMethods {
		if m == r.Method {
			return true
		}
	}
	return false
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := s.checkToken(r); s.isAuthenticationRequired(r) && err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		writeError(w, err)
		return
	}

	switch r.Method {
	case http.MethodPut:
		s.handlePut(w, r)
	case http.MethodOptions:
		s.handleOptions(w, r)
	default:
		w.Header().Add("Allow", "GET,HEAD,POST,PUT")
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeError(w, fmt.Errorf("method \"%s\" is not allowed", r.Method))
	}
}
