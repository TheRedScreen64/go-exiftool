package exiftool

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

type ExifTool struct {
	lock             sync.Mutex
	cmd              *exec.Cmd
	stdin            io.WriteCloser
	mergedOut        io.ReadCloser
	mergedOutScanner *bufio.Scanner
}

func NewExifTool(path string) (*ExifTool, error) {
	et := ExifTool{}

	args := []string{"-stay_open", "True", "-@", "-"}

	et.cmd = exec.Command(path, args...)

	read, write := io.Pipe()
	et.mergedOut = read
	et.cmd.Stdout = write
	et.cmd.Stderr = write

	var err error
	et.stdin, err = et.cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("error piping stdin: %w", err)
	}

	scanBuf := make([]byte, 64*1000)
	et.mergedOutScanner = bufio.NewScanner(read)
	et.mergedOutScanner.Buffer(scanBuf, 1024*1000)
	et.mergedOutScanner.Split(splitByReadyToken)

	err = et.cmd.Start()
	if err != nil {
		return nil, err
	}

	return &et, nil
}

func (et *ExifTool) Close() error {
	et.lock.Lock()
	defer et.lock.Unlock()

	closeArgs := []string{"-stay_open", "False"}
	for _, closeArg := range closeArgs {
		_, err := fmt.Fprintln(et.stdin, closeArg)
		if err != nil {
			return err
		}
	}

	err := et.mergedOut.Close()
	if err != nil {
		return err
	}

	err = et.stdin.Close()
	if err != nil {
		return err
	}

	err = et.cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

// NOTE: Param -j/-json for JSON formatted output
func (et *ExifTool) Exec(params ...string) (out string, outErrors []string, outWarnings []string, error error) {
	et.lock.Lock()
	defer et.lock.Unlock()

	for _, param := range params {
		_, err := fmt.Fprintln(et.stdin, param)
		if err != nil {
			return "", nil, nil, err
		}
	}

	_, err := fmt.Fprintln(et.stdin, "-execute")
	if err != nil {
		return "", nil, nil, err
	}

	scanOk := et.mergedOutScanner.Scan()
	scanErr := et.mergedOutScanner.Err()
	if scanErr != nil {
		return "", nil, nil, fmt.Errorf("error reading from scanner: %w", scanErr)
	}
	if !scanOk {
		return "", nil, nil, fmt.Errorf("error reading from scanner")
	}

	out = et.mergedOutScanner.Text()
	lines := strings.Split(out, "\n")

	for _, line := range lines {
		if strings.Contains(line, "Warning:") {
			warning := strings.TrimSpace(strings.TrimPrefix(line, "Warning:"))
			outWarnings = append(outWarnings, warning)
		} else if strings.Contains(line, "Error:") {
			err := strings.TrimSpace(strings.TrimPrefix(line, "Error:"))
			outErrors = append(outErrors, err)
		}
	}

	return out, outErrors, outWarnings, nil
}

func (et *ExifTool) GetMetadata(file ...string) ([]map[string]interface{}, error) {
	et.lock.Lock()
	defer et.lock.Unlock()

	params := []string{"-json"}
	params = append(params, file...)

	for _, param := range params {
		_, err := fmt.Fprintln(et.stdin, param)
		if err != nil {
			return nil, err
		}
	}

	_, err := fmt.Fprintln(et.stdin, "-execute")
	if err != nil {
		return nil, err
	}

	scanOk := et.mergedOutScanner.Scan()
	scanErr := et.mergedOutScanner.Err()
	if scanErr != nil {
		return nil, fmt.Errorf("error reading from scanner: %w", scanErr)
	}
	if !scanOk {
		return nil, fmt.Errorf("error reading from scanner")
	}

	var resMap []map[string]interface{}
	err = json.Unmarshal(et.mergedOutScanner.Bytes(), &resMap)
	if err != nil {
		return nil, fmt.Errorf("error during unmarshaling (%v): %w)", string(et.mergedOutScanner.Bytes()), err)
	}

	return resMap, nil
}
