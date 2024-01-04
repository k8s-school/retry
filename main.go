package main

import (
    "bufio"
    "flag"
    "fmt"
    "io"
    "os"
    "os/exec"
    "syscall"
    "time"
)

func retry(command string, maxRetries int, delay time.Duration, args ...string) int {
    n := 1

    for {
        cmd := exec.Command(command, args...)

        stdout, _ := cmd.StdoutPipe()
        stderr, _ := cmd.StderrPipe()

        go stream(stdout)
        go stream(stderr)

        err := cmd.Start()

        if err == nil && cmd.Wait() == nil {
            return 0
        } else {
            if exitError, ok := err.(*exec.ExitError); ok {
                if n < maxRetries {
                    n++
                    fmt.Printf("Command failed with exit code %d. Attempt %d/%d:\n", exitError.Sys().(syscall.WaitStatus).ExitStatus(), n, maxRetries)
                    time.Sleep(delay)
                } else {
                    fmt.Fprintf(os.Stderr, "The command has failed after %d attempts with exit code %d.\n", n, exitError.Sys().(syscall.WaitStatus).ExitStatus())
                    return exitError.Sys().(syscall.WaitStatus).ExitStatus()
                }
            }
        }
    }
    return 1
}

func stream(reader io.Reader) {
    scanner := bufio.NewScanner(reader)
    for scanner.Scan() {
        fmt.Println(scanner.Text())
    }
}

func main() {
    maxRetries := flag.Int("retries", 5, "the maximum number of retries")
    delay := flag.Int("delay", 5, "the delay between retries in seconds")
    flag.Parse()
    args := flag.Args()

    if len(args) < 1 {
        fmt.Println("Please provide a command to execute")
        os.Exit(1)
    }

    command := args[0]
    commandArgs := args[1:]

    exitCode := retry(command, *maxRetries, time.Duration(*delay)*time.Second, commandArgs...)
    os.Exit(exitCode)
}
