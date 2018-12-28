package handlers

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Trace traces requests.
func Trace() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)

		if err != nil {
			ip = r.RemoteAddr
		}

		output, err := RunCommand("/usr/bin/traceroute -m 5 "+ip, []byte(""), 120)

		if err != nil {
			log.Error(err)
			http.Error(w, "error running trace", http.StatusInternalServerError)
			return
		}

		for _, line := range strings.Split(string(output), "\n") {
			log.Info(line)
		}

		w.Write(output)
	})
}

// RunCommand runs commands.
func RunCommand(command string, message []byte, timeout int) ([]byte, error) {
	var output []byte

	log.Debugf("running: %s", command)

	commandArgs := strings.Fields(command)

	var args []string

	if len(commandArgs) > 1 {
		args = commandArgs[1:]
	}

	cmd := exec.Command(commandArgs[0], args...)

	input, err := cmd.StdinPipe()

	if err != nil {
		log.Error("error with stdin pipe")
		return output, err
	}

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		log.Error("error with stdout pipe")
		return output, err
	}

	defer stdout.Close()

	stderr, err := cmd.StderrPipe()

	if err != nil {
		log.Error("error with stderr pipe")
		return output, err
	}

	defer stderr.Close()

	if err := cmd.Start(); err != nil {
		log.Error("error starting command")
		return output, err
	}

	_, err = input.Write(message)

	if err != nil {
		log.Error("error writing message to stdin")
		return output, err
	}

	input.Close()

	scanner := bufio.NewScanner(stderr)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			log.Info(line)
		}
	}()

	done := make(chan error)

	go func() {
		output, err = ioutil.ReadAll(stdout)

		if err != nil {
			log.Error("error reading output")
			done <- err
			return
		}

		done <- cmd.Wait()
	}()

	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		log.Error("command timed out")

		if err := cmd.Process.Kill(); err != nil {
			log.Errorf("failed to kill cmd: %v", err)
			return output, err
		}

		return output, fmt.Errorf("command timed out")
	case cmdErr := <-done:
		if cmdErr != nil {
			log.Debug("command failed")
			return output, cmdErr
		}

		log.Debug("command exit ok")
	}

	return output, nil
}
