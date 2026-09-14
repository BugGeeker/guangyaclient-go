package guangyaclient

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func GenerateDID() string {
	var seed [16]byte
	if _, err := rand.Read(seed[:]); err != nil {
		panic(fmt.Errorf("generate device id: %w", err))
	}
	sum := md5.Sum(seed[:])
	return hex.EncodeToString(sum[:])
}

func GenerateTraceparent() string {
	var traceID [16]byte
	var parentID [8]byte
	if _, err := rand.Read(traceID[:]); err != nil {
		panic(fmt.Errorf("generate trace id: %w", err))
	}
	if _, err := rand.Read(parentID[:]); err != nil {
		panic(fmt.Errorf("generate parent id: %w", err))
	}
	return fmt.Sprintf("00-%s-%s-01", hex.EncodeToString(traceID[:]), hex.EncodeToString(parentID[:]))
}

// CalculateCID hashes the whole file up to 60 KiB; larger files use three
// 20 KiB samples at the start, one-third offset, and end.
func CalculateCID(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	const sampleSize int64 = 0x5000
	hash := sha1.New()
	if info.Size() <= 3*sampleSize {
		if _, err := io.Copy(hash, file); err != nil {
			return "", err
		}
	} else {
		for _, offset := range []int64{0, info.Size() / 3, info.Size() - sampleSize} {
			if _, err := file.Seek(offset, io.SeekStart); err != nil {
				return "", err
			}
			if _, err := io.CopyN(hash, file, sampleSize); err != nil {
				return "", err
			}
		}
	}
	return stringsToUpper(hex.EncodeToString(hash.Sum(nil))), nil
}

func CalculateGCID(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	chunkSize := int64(262144)
	switch {
	case info.Size() <= 0x8000000:
	case info.Size() <= 0x10000000:
		chunkSize = 524288
	case info.Size() <= 0x20000000:
		chunkSize = 1048576
	default:
		chunkSize = 2097152
	}
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	var hashes []byte
	buf := make([]byte, chunkSize)
	for {
		n, readErr := io.ReadFull(file, buf)
		if readErr == io.EOF {
			break
		}
		if readErr != nil && readErr != io.ErrUnexpectedEOF {
			return "", readErr
		}
		sum := sha1.Sum(buf[:n])
		hashes = append(hashes, sum[:]...)
		if readErr == io.ErrUnexpectedEOF {
			break
		}
	}
	final := sha1.Sum(hashes)
	return stringsToUpper(hex.EncodeToString(final[:])), nil
}

func stringsToUpper(value string) string {
	const lower = "abcdefghijklmnopqrstuvwxyz"
	const upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	out := []byte(value)
	for i, c := range out {
		for j := range lower {
			if c == lower[j] {
				out[i] = upper[j]
				break
			}
		}
	}
	return string(out)
}
