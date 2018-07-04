package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
)

var supportedLanguage = []string{
	"c",
	"cpp",
	"java",
	"php",
	"python",
	"bash",
}

var cLanguage = []string{
	"c",
	"cpp",
	"java",
}

func goRun(workDir, stdin string, args ...string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	// args[0] is the program name
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = workDir
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return stdout.String(), stderr.String(), err
}

// Run will run payload, if the language need compille
// it will call CompileAndRun()
func (ar *PayLoad) Run() {

	if len(ar.A.Run) == 0 {
		ar.A.Run = []string{ar.L}
	}
	args := ar.A.Run[0:]
	workDir := ""
	if len(ar.F) != 0 {
		absFilePaths, err := writeFiles(ar.F)
		if err != nil {
			exitF("Write files failed")
		}
		workDir = filepath.Dir(absFilePaths[0])
		args = append(ar.A.Run[0:], absFilePaths...)
	}

	stdOut, stdErr, exitErr := goRun(workDir, ar.I, args...)
	returnStdOut(stdOut, stdErr, errToStr(exitErr))
}

func (ar *PayLoad) compileAndRun() {
	// if no file return
	if len(ar.F) == 0 {
		exitF("No fileds are given")
	}
	absFilePaths, err := writeFiles(ar.F)
	if err != nil {
		exitF("Write files failed")
	}
	workDir := filepath.Dir(absFilePaths[0])

	switch {
	case ar.L == "c" || ar.L == "cpp":
		if len(ar.A.Compile) == 0 {
			if ar.L == "c" {
				ar.A.Compile = []string{"gcc"}
			} else {
				ar.A.Compile = []string{"g++"}
			}
		}
		binName := "a.out"

		args := append(ar.A.Compile, []string{"-o", binName}...)
		args = append(args, absFilePaths...)
		// compile
		stdOut, stdErr, exitErr := goRun(workDir, "", args...)
		if exitErr != nil {
			if _, ok := exitErr.(*exec.ExitError); ok {
				returnStdOut(stdOut, stdErr, errToStr(exitErr))
				exitF("Compile Error")
			}
			exitF("Ric goRun Failed")
		}

		// run
		binPath := filepath.Join(workDir, binName)
		args = append(ar.A.Run, binPath)

		stdOut, stdErr, exitErr = goRun(workDir, ar.I, args...)
		returnStdOut(stdOut, stdErr, errToStr(exitErr))

	case ar.L == "java":
		if len(ar.A.Compile) == 0 {
			ar.A.Compile = []string{"javac"}
		}

		args := append(ar.A.Compile, absFilePaths...)

		fname := filepath.Base(absFilePaths[0])

		// compile
		stdOut, stdErr, exitErr := goRun(workDir, "", args...)
		if exitErr != nil {
			returnStdOut(stdOut, stdErr, errToStr(exitErr))
			exitF("Compile Error")
		}

		stdOut, stdErr, exitErr = goRun(workDir, ar.I, "java", javaClassName(fname))
		returnStdOut(stdOut, stdErr, errToStr(exitErr))

	default:
		exitF("Unsupported compile language: %s", ar.L)
	}
}

func (ar *PayLoad) needCompile() bool {
	for _, l := range cLanguage {
		if ar.L == l {
			return true
		}
	}
	return false
}

func (ar *PayLoad) isSupport() bool {
	for _, l := range supportedLanguage {
		if ar.L == l {
			return true
		}
	}
	return false
}

func javaClassName(fname string) string {
	ext := filepath.Ext(fname)
	return fname[0 : len(fname)-len(ext)]
}