package exiftool

import (
	"bytes"
	"fmt"
)

func splitByReadyToken(data []byte, atEOF bool) (int, []byte, error) {
	readyToken := []byte("{ready}")

	i := bytes.Index(data, readyToken)
	if i == -1 {
		if atEOF && len(data) > 0 {
			return 0, data, fmt.Errorf("ready token not found at EOF")
		}

		return 0, nil, nil
	}

	return i + len(readyToken), data[:i], nil
}
