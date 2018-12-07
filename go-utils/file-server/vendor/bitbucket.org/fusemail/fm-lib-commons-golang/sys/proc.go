package sys

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

type Process struct {
	pid     int
	path    string
	name    string
	state   string
	cmdline []string
}

func (p Process) String() string {
	return fmt.Sprintf("%s: [%d] %s (%s) state: %s", p.name, p.pid, p.path, strings.Join(p.cmdline, " "), p.state)
}

func (p Process) Zombie() bool {
	return p.state == "Z"
}

func (p Process) Running() bool {
	return p.state == "R"
}

func FindProcessByName(name string) ([]Process, error) {
	var matches []Process

	procs, err := ProcessList()

	if err != nil {
		return nil, err
	}

	for _, proc := range procs {
		log.Debug(proc)
		if strings.HasSuffix(proc.path, name) {
			matches = append(matches, proc)
		}
	}

	return matches, nil
}

const procReadBatchSize = 20

func ProcessList() ([]Process, error) {
	var processList []Process

	d, err := os.Open("/proc")

	if err != nil {
		return nil, err
	}

	defer d.Close()

	for {
		procs, err := d.Readdir(procReadBatchSize)

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		for _, proc := range procs {
			if !proc.IsDir() {
				continue
			}

			if pid, err := strconv.Atoi(proc.Name()); err == nil {
				process, err := NewProcess(pid)

				if err != nil {
					continue
				}

				processList = append(processList, process)
			}
		}
	}

	return processList, nil
}

func NewProcess(pid int) (Process, error) {
	proc := Process{
		pid: pid,
	}

	root := path.Join("/proc", fmt.Sprintf("%d", pid))

	s, err := os.Open(path.Join(root, "status"))

	if err != nil {
		return proc, err
	}

	defer s.Close()

	name, state, err := parseStatus(s)

	if err != nil {
		log.Debug("cannot read state")
		return proc, err
	}

	proc.name = name
	proc.state = state

	if proc.state == "Z" {
		log.Debug("returning a zombie")
		return proc, nil
	}

	binary, err := os.Readlink(path.Join(root, "exe"))

	if err != nil {
		return proc, err
	}

	proc.path = binary

	f, err := os.Open(path.Join(root, "cmdline"))

	if err != nil {
		return proc, err
	}

	defer f.Close()

	data, err := ioutil.ReadAll(f)

	options := bytes.Split(data, []byte("\x00"))

	for _, option := range options {
		proc.cmdline = append(proc.cmdline, string(option))
	}

	return proc, nil
}

func parseStatus(r io.Reader) (name string, state string, err error) {
	var data []byte

	data, err = ioutil.ReadAll(r)

	if err != nil {
		return
	}

	lines := bytes.Split(data, []byte("\n"))

	for _, line := range lines {
		fields := strings.Fields(string(line))
		if len(fields) > 0 {
			switch fields[0] {
			case "Name:":
				if len(fields) > 1 {
					name = fields[1]
				}
			case "State:":
				if len(fields) > 1 {
					state = fields[1]
				}
				break
			}
		}
	}

	return
}
