package upyun

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"
)

// ----- concurrent upload with checkpoint  -----
const uploadCpMagic = "287E871C-90AA-4D71-91DD-9D68C08BDC8E"

type uploadCheckpoint struct {
	Magic     string         `json:"magic"`      // Magic
	MD5       string         `json:"md5"`        // Checkpoint file content's MD5
	FilePath  string         `json:"file_path"`  // Local file path
	FileStat  uploadFileStat `json:"file_stat"`  // File state
	ObjectKey string         `json:"object_key"` // Key
	UploadID  string         `json:"upload_id"`  // Upload ID
	Parts     []*uploadPart  `json:"parts"`      // All parts of the local file
	cpPath    string
}

type uploadPart struct {
	PartID    int   `json:"part_id"`
	PartSize  int64 `json:"part_size"`
	Offset    int64 `json:"offset"`
	Completed bool  `json:"completed"`
}
type uploadFileStat struct {
	Size         int64     `json:"size"`          // File size
	LastModified time.Time `json:"last_modified"` // File's last modified time
}

// isValid checks if the uploaded data is valid---it's valid when the file is not updated and the checkpoint data is valid.
func newCheckpoint(filePath, key, cpDir string) (*uploadCheckpoint, error) {
	cp := &uploadCheckpoint{
		FilePath:  filePath,
		ObjectKey: key,
		Magic:     uploadCpMagic,
	}
	if cpDir != "" {
		cpPath := path.Join(cpDir, getCpFileName(filePath, key))
		cp.cpPath = cpPath
		err := cp.load(cpPath)
		if err != nil {
			cp.remove()
		} else {
			valid, err := cp.isValid()
			if err != nil {
				return nil, fmt.Errorf("check checkpoint file error: %v", err)
			}
			if valid {
				return cp, nil
			}
			cp.remove()
		}
	}
	return cp, nil
}
func (cp *uploadCheckpoint) init(sliceSize int64) error {
	// Local file
	fd, err := os.Open(cp.FilePath)
	if err != nil {
		return err
	}
	defer fd.Close()

	st, err := fd.Stat()
	if err != nil {
		return err
	}
	fileSize := st.Size()
	cp.FileStat.Size = fileSize
	cp.FileStat.LastModified = st.ModTime()
	if err != nil {
		return err
	}

	chunkCount := (fileSize + sliceSize - 1) / sliceSize

	cp.Parts = make([]*uploadPart, chunkCount)
	for i := 0; i < int(chunkCount); i++ {
		cp.Parts[i] = &uploadPart{
			PartID:   i,
			Offset:   sliceSize * int64(i),
			PartSize: sliceSize,
		}
	}
	cp.Parts[chunkCount-1].PartSize = fileSize - (chunkCount-1)*sliceSize
	return nil
}
func (cp *uploadCheckpoint) isValid() (bool, error) {
	// Compare the CP's magic number and MD5.
	cpb := *cp
	cpb.MD5 = ""
	js, _ := json.Marshal(cpb)
	sum := md5.Sum(js)
	b64 := base64.StdEncoding.EncodeToString(sum[:])

	if cp.Magic != uploadCpMagic || b64 != cp.MD5 {
		return false, nil
	}

	// Make sure if the local file is not updated.
	fd, err := os.Open(cp.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer fd.Close()

	st, err := fd.Stat()
	if err != nil {
		return false, err
	}

	// Compare the file size, file's last modified time and file's MD5
	if cp.FileStat.Size != st.Size() || cp.FileStat.LastModified != st.ModTime() {
		return false, nil
	}

	return true, nil
}

// load loads from the file
func (cp *uploadCheckpoint) load(filePath string) error {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(contents, cp)
	return err
}

// dump dumps to the local file
func (cp *uploadCheckpoint) dump() error {
	if cp.cpPath == "" {
		return nil
	}
	bcp := *cp

	// Calculate MD5
	bcp.MD5 = ""
	js, err := json.Marshal(bcp)
	if err != nil {
		return err
	}
	sum := md5.Sum(js)
	b64 := base64.StdEncoding.EncodeToString(sum[:])
	bcp.MD5 = b64

	// Serialization
	js, err = json.Marshal(bcp)
	if err != nil {
		return err
	}

	// Dump
	return ioutil.WriteFile(cp.cpPath, js, os.FileMode(0664))
}
func (cp *uploadCheckpoint) completePart(partID int) {
	cp.Parts[partID].Completed = true
}

// todoParts returns unfinished parts
func (cp *uploadCheckpoint) todoParts() []*uploadPart {
	parts := []*uploadPart{}
	for _, part := range cp.Parts {
		if !part.Completed {
			parts = append(parts, part)
		}
	}
	return parts
}

// getCompletedBytes returns completed bytes count
func (cp *uploadCheckpoint) getCompletedBytes() int64 {
	var completedBytes int64
	for _, part := range cp.Parts {
		if part.Completed {
			completedBytes += part.PartSize
		}
	}
	return completedBytes
}
func (cp *uploadCheckpoint) remove() {
	if cp.cpPath != "" {
		os.Remove(cp.cpPath)
	}
}

func getCpFileName(src, dest string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(src))
	srcCheckSum := hex.EncodeToString(md5Ctx.Sum(nil))

	md5Ctx.Reset()
	md5Ctx.Write([]byte(dest))
	destCheckSum := hex.EncodeToString(md5Ctx.Sum(nil))

	return fmt.Sprintf("%v-%v.json", srcCheckSum, destCheckSum)
}
