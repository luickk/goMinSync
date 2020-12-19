package util

import (
  "os"
  "crypto/sha256"
	"encoding/hex"
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
  io.Copy(hasher, file)
	return hex.EncodeToString(hasher.Sum(nil))
}

// finds changes between new and old (key)path (val)hash map change
// returns map of (key) type of change (val) hash of changed file
func FindPathHashMapChange(new map[string]string, old map[string]string) map[string]string {
  diff := make(map[string]string, 0)
  fmt.Println(new)
  fmt.Println(old)
  for oldPath, oldHash := range old {
    // path is not new to dir
    // stayed the same
    if newHash, ok := new[oldPath]; ok {
      // file content did not change
      if newHash == oldHash {

      // file content did not change
      } else {
        diff["change"] = oldPath

      }
    // path is removed from dir
    } else {
      diff["removed"] = oldPath
    }
  }

  for newPath, _ := range new {
    // path is added to dir
    if _, ok := old[newPath]; !ok {
      diff["added"] = newPath
    }
  }
  return diff
}
