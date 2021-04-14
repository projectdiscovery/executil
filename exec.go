package executil

import (
	"errors"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
)

const (
	// token of prefix for space, escape char '\'
	tkSpPerfix = '\\'
	// the space
	tkSp = ' '
)

// split cmds with " ".
// recognize "\ " as a " " is a part of one argument
// '\' is escape char only effect with space ' '.
// so "\\a" is also the "\\a" in argment, not "\a".
// "b\a" -> "b\a"
func splitCmdAgrs(cmds string) []string {

	raw := []byte(cmds)
	length := len(raw)

	argments := make([]string, 0, 4)

	escape := false
	// argment recognizing
	argRec := false
	var rawArg []byte

	fillArgChar := func(c byte) []byte {
		if rawArg == nil {
			rawArg = make([]byte, 0, 16)
		}
		rawArg = append(rawArg, c)
		return rawArg
	}

	finishArg := func() {
		// finish one argment recognized
		if len(rawArg) > 0 {
			argment := string(rawArg)
			argments = append(argments, argment)
		}
		rawArg = nil
	}

	for i := 0; i < length; i++ {
		c := raw[i]
		switch c {
		case tkSp:
			if escape {
				fillArgChar(tkSp)
				escape = false // exit excape
			} else if argRec {
				argRec = false
				finishArg()
			}
		case tkSpPerfix:
			if escape {
				// like "\\"
				fillArgChar(tkSpPerfix) // first '\'
			} else {
				escape = true
			}
			if !argRec {
				argRec = true
			}
		default:
			if !argRec {
				argRec = true
			}
			if escape {
				escape = false
				fillArgChar(tkSpPerfix)
			}
			fillArgChar(c)
		}
	}

	if escape {
		escape = false
		fillArgChar(tkSpPerfix)
	}

	if argRec {
		argRec = false
		finishArg()
	}

	return argments
}

// convert spaces in arg to "\ "
// so it can split with func [splitCmdAgrs]
func safeArg(arg string) string {
	return strings.ReplaceAll(arg, " ", "\\ ")
}

func splitt(s string) (tokens []string) {
	for _, ss := range strings.Split(s, " ") {
		tokens = append(tokens, strings.Split(ss, "\n")...)
	}
	return
}

func Run(cmd string) (string, error) {
	return RunSh(splitCmdAgrs(cmd)...)
}

func RunSh(cmd ...string) (string, error) {
	cmdExec := exec.Command(cmd[0], cmd[1:]...)
	in, _ := cmdExec.StdinPipe()
	errorOut, _ := cmdExec.StderrPipe()
	out, _ := cmdExec.StdoutPipe()
	defer closeIO(in)
	defer closeIO(errorOut)
	defer closeIO(out)

	if err := cmdExec.Start(); err != nil {
		return "", errors.New("start sh process error")
	}

	outData, _ := ioutil.ReadAll(out)
	errorData, _ := ioutil.ReadAll(errorOut)

	var adbError error = nil

	if err := cmdExec.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			adbError = errors.New("sh return error")
			outData = errorData
		} else {
			return "", errors.New("start sh process error")
		}
	}

	return string(outData), adbError
}

// close a stream
func closeIO(c io.Closer) {
	if c != nil {
		c.Close()
	}
}
