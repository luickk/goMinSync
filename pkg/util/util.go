package util

import (
  "os"
  "crypto/sha256"
	"encoding/hex"
  "fmt"
  "io"
)

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

func FindSliceDifference(slice1 []string, slice2 []string) map[string]string {
    var diff map[string]string

    for i := 0; i < 2; i++ {
        for _, s1 := range slice1 {
            found := false
            for _, s2 := range slice2 {
                if s1 == s2 {
                    found = true
                    break
                }
            }
            if !found {
                diff["added"] = s1
            }
        }
        if i == 0 {
            slice1, slice2 = slice2, slice1
        }
    }

    return diff
}
