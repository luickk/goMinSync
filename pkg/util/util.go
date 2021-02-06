package util

import (
  "crypto/sha256"
	"encoding/hex"
  "encoding/json"
  "net/http"
  "mime/multipart"
  "io/ioutil"
  "path/filepath"

  "bytes"
  "os"
  "fmt"


  "io"
)

type FileChange struct {
  Ctype string
  AbsPath string
  RelPath string
  FileHash string
  IsDir bool
  SyncGroup int
  OriginId int
}

func EncodeMsg(msg *FileChange) ([]byte, error) {
	serializedMsg, err := json.Marshal(&msg)
	if err != nil {
		return []byte{}, err
	}
	return serializedMsg, nil
}


func DecodeMsg(msg *FileChange, data []byte) error {
	err := json.Unmarshal(data, msg)
	if err != nil {
		return err
	}
	return nil
}

// by https://stackoverflow.com/questions/51234464/upload-a-file-with-post-request-golang
func PostUploadFile(url string, filePath string, filetype string) ([]byte, error) {
  file, err := os.Open(filePath)

  if err != nil {
    return nil, err
  }
  defer file.Close()


  body := &bytes.Buffer{}
  writer := multipart.NewWriter(body)
  part, err := writer.CreateFormFile(filetype, filepath.Base(file.Name()))

  if err != nil {
    return nil, err
  }

  io.Copy(part, file)
  writer.Close()
  request, err := http.NewRequest("POST", url, body)

  if err != nil {
    return nil, err
  }

  request.Header.Add("Content-Type", writer.FormDataContentType())
  client := &http.Client{}

  response, err := client.Do(request)

  if err != nil {
    return nil, err
  }
  defer response.Body.Close()

  content, err := ioutil.ReadAll(response.Body)

  if err != nil {
    return nil, err
  }

  return content, nil
}

// by https://stackoverflow.com/questions/11692860/how-can-i-efficiently-download-a-large-file-using-go
func DownloadFile(filepath string, url string) (err error) {
  // Create the file
  out, err := os.Create(filepath)
  if err != nil  {
    return err
  }
  defer out.Close()

  // Get the data
  resp, err := http.Get(url)
  if err != nil {
    return err
  }
  defer resp.Body.Close()

  // Check server response
  if resp.StatusCode != http.StatusOK {
    return fmt.Errorf("bad status: %s", resp.Status)
  }

  // Writer the body to file
  _, err = io.Copy(out, resp.Body)
  if err != nil  {
    return err
  }

  return nil
}

// hashes content of file
// hashes path
// returns sha256 hex encoded string
func Sha256DirObj(path string) (string, error) {
  file, err := os.Open(path)
  if err != nil {
    return "", err
  }
  defer file.Close()
  hasher := sha256.New()

  _, err = io.Copy(hasher, file)
  if err != nil {
    // file content is either not hashable or a directory
    return "", err
  }
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// finds changes between new and old (key)path (val)hash map change
// returns map of (key) type of change (val) hash of changed file
func FindPathHashMapChange(new map[string]string, old map[string]string, dir string) ([]FileChange, error) {
  changes := []FileChange{}

  for oldPath, oldHash := range old {
    // path is not new to dir
    // stayed the same
    if newHash, ok := new[oldPath]; ok {
      // file content did not change
      if newHash == oldHash {
      // file content did change
      } else {
        ok, err := IsDirectory(dir+oldPath)
        if err != nil {
          return nil, err
        }
        changes = append(changes, FileChange{ "changed", dir+oldPath, oldPath, oldHash, ok, 0, 0 })
      }
    // path is removed from dir
    } else {
      changes = append(changes, FileChange{ "removed", dir+oldPath, oldPath, oldHash, false, 0, 0 })
    }
  }

  for newPath, newHash := range new {
    // path is added to dir
    if _, ok := old[newPath]; !ok {
      ok, err := IsDirectory(dir+newPath)
      if err != nil {
        return nil, err
      }
      changes = append(changes, FileChange{ "added", dir+newPath, newPath, newHash, ok, 0, 0 })
    }
  }
  return changes, nil
}

func HashFile(filePath string) (string, error) {
  var hash string
  fi, err := os.Stat(filePath)
  if err != nil {
    return "", err
  }
  if fi.Mode().IsRegular() {
    hash, err = Sha256DirObj(filePath)
    if err != nil {
      return "", err
    }
  } else {
    return "", nil
  }
  if err != nil {
    return "", err
  }
  return hash, nil
}


func Hash(toHash []byte) (string, error) {
  hasher := sha256.New()
  _, err := io.Writer(hasher).Write(toHash)
  if err != nil {
    // file content is either not hashable or a directory
    return "", err
  }
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func CreateFileSig(path string) (string, error) {
  var (
    err error
    filePathHash string
    fileHash string
    fileSig string
  )
  filePathHash, err = Hash([]byte(path))
  if err != nil {
    return "", err
  }
  fileHash, err = HashFile(path)
  if err != nil {
    return "", err
  }
  fileSig, err = Hash([]byte(filePathHash+fileHash))
  if err != nil {
    return "", err
  }
  return fileSig, nil
}

func IsDirectory(path string) (bool, error) {
    fileInfo, err := os.Stat(path)
    if err != nil{
      return false, err
    }
    return fileInfo.IsDir(), err
}

func CreatePathHashMap(dir string) (map[string]string, error) {
  var (
    hash string
    relPath string
  )
  pathHashMap := make(map[string]string, 0)

  err := filepath.Walk(dir, func(absPath string, f os.FileInfo, err error) error {
    hash, err = HashFile(absPath)
    if err != nil {
      return err
    }
    relPath = absPath[len(dir):]
    if relPath != "" {
      pathHashMap[relPath] = hash
    }
    return nil
  })
  if err != nil {
    return nil, err
  }
  return pathHashMap, nil
}
