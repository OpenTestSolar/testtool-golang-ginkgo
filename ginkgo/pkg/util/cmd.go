package util

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/sourcegraph/conc"
)

type ReaderCallback func(line string)

func forwardStream(stream io.ReadCloser, cb ReaderCallback) {
	var buffer []byte
	buff := make([]byte, 4096)
	for {
		n, err := stream.Read(buff)
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("Read pipe %v error: %v", stream, err)
			}
			break
		}
		if n == 0 {
			break
		}
		buffer = append(buffer, buff[:n]...)
		pos := bytes.Index(buffer, []byte{'\n'})
		for {
			if pos < 0 {
				break
			}
			lineBytes := buffer[:pos]
			cb(string(lineBytes))
			buffer = buffer[pos+1:]
			pos = bytes.Index(buffer, []byte{'\n'})
		}
	}
	if len(buffer) > 0 {
		cb(string(buffer))
	}
}

func RunCommandWithEnvs(cmdline string, projPath string, envs map[string]string, isWait bool, redirect bool) (*exec.Cmd, io.ReadCloser, io.ReadCloser, error) {
	cmd := exec.Command("bash", "-c", cmdline)
	cmd.Dir = projPath
	cmd.Env = os.Environ()
	for k, v := range envs {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	var stdout, stderr bytes.Buffer
	var stdoutReader, stderrReader io.ReadCloser
	if isWait {
		if !redirect {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		} else {
			cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
			cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
			stdoutReader = io.NopCloser(bytes.NewReader(stdout.Bytes()))
			stderrReader = io.NopCloser(bytes.NewReader(stderr.Bytes()))
		}
	} else {
		stdoutReader, _ = cmd.StdoutPipe()
		stderrReader, _ = cmd.StderrPipe()
	}

	err := cmd.Start()
	if err != nil {
		return cmd, nil, nil, err
	}
	if isWait {
		err = cmd.Wait()
		return cmd, stdoutReader, stderrReader, err
	}
	return cmd, stdoutReader, stderrReader, nil
}

func RunCommand(cmdline string, projPath string, isWait bool, redirect bool) (io.ReadCloser, io.ReadCloser, error) {
	_, stdoutReader, stderrReader, err := RunCommandWithEnvs(cmdline, projPath, nil, isWait, redirect)
	return stdoutReader, stderrReader, err
}

func RunCommandWithOutput(cmdline string, projPath string) (string, string, error) {
	var stdout, stderr string
	var wg conc.WaitGroup
	outStream, errStream, err := RunCommand(cmdline, projPath, false, true)
	if err != nil {
		return "", "", err
	}
	wg.Go(
		func() {
			forwardStream(outStream, func(line string) {
				log.Printf("[OUT] %s\n", line)
				stdout += line + "\n"
			})
		},
	)
	wg.Go(
		func() {
			forwardStream(errStream, func(line string) {
				log.Printf("[ERR] %s\n", line)
				stderr += line + "\n"
			})
		},
	)
	wg.Wait()
	return stdout, stderr, nil
}
