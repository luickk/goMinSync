package util

import (
  "crypto/sha256"
	"encoding/hex"
  "path/filepath"
  "os"
  "fmt"
  "io"
)

// hashes content of file
// hashes path
// returns sha256 hex encoded string
func Sha256DirObj(path string) string {
  file, err := os.Open(path)
  if err != nil {
    fmt.Println(err)
  }
  defer file.Close()
  hasher := sha256.New()

  _, err = io.Copy(hasher, file)
  if err != nil {
    // file content is either not hashable or a directory
    fmt.Println(err)
    return ""
  }
	return hex.EncodeToString(hasher.Sum(nil))
}

// finds changes between new and old (key)path (val)hash map change
// returns map of (key) type of change (val) hash of changed file
func FindPathHashMapChange(new map[string]string, old map[string]string) map[string]string {
  diff := make(map[string]string, 0)
  for oldPath, oldHash := range old {
    // path is not new to dir
    // stayed the same
    if newHash, ok := new[oldPath]; ok {
      // file content did not change
      if newHash == oldHash {

      // file content did change
      } else {
        diff[oldPath] = "changed"

      }
    // path is removed from dir
    } else {
      diff[oldPath] = "removed"
    }
  }

  for newPath, _ := range new {
    // path is added to dir
    if _, ok := old[newPath]; !ok {
      diff[newPath] = "added"
    }
  }
  return diff
}

func CreatePathHasMap(dir string) (map[string]string, error) {
  var hash string
  pathHashMap := make(map[string]string, 0)
  err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
    fi, err := os.Stat(path)
    if err != nil {
      return err
    }
    if fi.Mode().IsRegular() {
      hash = Sha256DirObj(path)
      pathHashMap[path] = hash
    } else {
      pathHashMap[path] = ""
    }
    return nil
  })
  if err != nil {
    return nil, err
  }
  return pathHashMap, nil
}
