// by https://github.com/mayth/go-simple-upload-server
// repository is not cloned due to major changes!

package fileUploadServer

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"encoding/json"
	"path"
	"regexp"
	"strings"

	"goMinSync/pkg/util"
)

var (
	rePathUpload = regexp.MustCompile(`^/upload$`)
	rePathFiles  = regexp.MustCompile(`^/files/([^/]+)$`)

	errTokenMismatch = errors.New("token mismatched")
	errMissingToken  = errors.New("missing token")

	protectedMethods = []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut}
)

// Server represents a simple-upload server.
type Server struct {
	DocumentRoot string
	TokenEnabled bool
	SecureToken string
	EnableCORS bool
}

// NewServer creates a new simple-upload server.
func NewServer(documentRoot string, tokenEnabled bool, token string, enableCORS bool) Server {
	return Server{
		DocumentRoot: documentRoot,
		TokenEnabled: tokenEnabled,
		SecureToken: token,
		EnableCORS: enableCORS,
	}
}

func (s Server) handleGet(w http.ResponseWriter, r *http.Request) {
	if !rePathFiles.MatchString(r.URL.Path) {
		w.WriteHeader(http.StatusNotFound)
		writeError(w, fmt.Errorf("\"%s\" is not found", r.URL.Path))
		return
	}
	if s.EnableCORS {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	http.StripPrefix("/files/", http.FileServer(http.Dir(s.DocumentRoot))).ServeHTTP(w, r)
}

func (s Server) handlePost(w http.ResponseWriter, r *http.Request) {
	srcFile, info, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	}
	defer srcFile.Close()
	size, err := getSize(srcFile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	}

	body, err := ioutil.ReadAll(srcFile)

	newName := r.URL.Query().Get("name")
	var filename string
	if newName != "" {
		filename = newName
	} else {
		if info.Filename == "" {
			filename, err = util.Hash(body)
			if err != nil {
				return
			}
		}
		filename = info.Filename
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	}

	dstPath := path.Join(s.DocumentRoot, filename)
	dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	}
	defer dstFile.Close()
	if written, err := dstFile.Write(body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	} else if int64(written) != size {
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, fmt.Errorf("the size of uploaded content is %d, but %d bytes written", size, written))
	}
	uploadedURL := strings.TrimPrefix(dstPath, s.DocumentRoot)
	if !strings.HasPrefix(uploadedURL, "/") {
		uploadedURL = "/" + uploadedURL
	}
	uploadedURL = "/files" + uploadedURL
	if s.EnableCORS {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	w.WriteHeader(http.StatusOK)
	writeSuccess(w, uploadedURL)
}

func (s Server) handlePut(w http.ResponseWriter, r *http.Request) {
	matches := rePathFiles.FindStringSubmatch(r.URL.Path)
	if matches == nil {
		w.WriteHeader(http.StatusNotFound)
		writeError(w, fmt.Errorf("\"%s\" is not found", r.URL.Path))
		return
	}
	targetPath := path.Join(s.DocumentRoot, matches[1])

	// We have to create a new temporary file in the same device to avoid "invalid cross-device link" on renaming.
	// Here is the easiest solution: create it in the same directory.
	tempFile, err := ioutil.TempFile(s.DocumentRoot, "upload_")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	}
	defer r.Body.Close()
	srcFile, _, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	}
	defer srcFile.Close()

	_, err = io.Copy(tempFile, srcFile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	}
	// excplicitly close file to flush, then rename from temp name to actual name in atomic file
	// operation if on linux or other unix-like OS (windows hosts should look into https://github.com/natefinch/atomic
	// package for atomic file write operations)
	tempFile.Close()
	if err := os.Rename(tempFile.Name(), targetPath); err != nil {
		os.Remove(tempFile.Name())
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, err)
		return
	}

	if s.EnableCORS {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	w.WriteHeader(http.StatusOK)
	writeSuccess(w, r.URL.Path)
}

func (s Server) handleOptions(w http.ResponseWriter, r *http.Request) {
	var allowedMethods []string
	if rePathFiles.MatchString(r.URL.Path) {
		allowedMethods = []string{http.MethodPut, http.MethodGet, http.MethodHead}
	} else if rePathUpload.MatchString(r.URL.Path) {
		allowedMethods = []string{http.MethodPost}
	} else {
		w.WriteHeader(http.StatusNotFound)
		writeError(w, errors.New("not found"))
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ","))
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

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.TokenEnabled {
		if err := s.checkToken(r); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			writeError(w, err)
			return
		}
	}

	switch r.Method {
	case http.MethodGet, http.MethodHead:
		s.handleGet(w, r)
	case http.MethodPost:
		s.handlePost(w, r)
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

// UTIL
type response struct {
	OK bool `json:"ok"`
}

type uploadedResponse struct {
	response
	Path string `json:"path"`
}

func newUploadedResponse(path string) uploadedResponse {
	return uploadedResponse{response: response{OK: true}, Path: path}
}

type errorResponse struct {
	response
	Message string `json:"error"`
}

func newErrorResponse(err error) errorResponse {
	return errorResponse{response: response{OK: false}, Message: err.Error()}
}

func writeError(w http.ResponseWriter, err error) (int, error) {
	body := newErrorResponse(err)
	b, e := json.Marshal(body)
	// if an error is occured on marshaling, write empty value as response.
	if e != nil {
		return w.Write([]byte{})
	}
	return w.Write(b)
}

func writeSuccess(w http.ResponseWriter, path string) (int, error) {
	body := newUploadedResponse(path)
	b, e := json.Marshal(body)
	// if an error is occured on marshaling, write empty value as response.
	if e != nil {
		return w.Write([]byte{})
	}
	return w.Write(b)
}

func getSize(content io.Seeker) (int64, error) {
	size, err := content.Seek(0, os.SEEK_END)
	if err != nil {
		return 0, err
	}
	_, err = content.Seek(0, io.SeekStart)
	if err != nil {
		return 0, err
	}
	return size, nil
}
