// Package bash provides utilities for executing bash commands and processing their output.
// It includes functionality for process management, system monitoring, and output formatting.
package bash

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	// ansi is a regular expression pattern for matching ANSI escape sequences.
	// These sequences are used for terminal color formatting and control.
	// The pattern matches:
	// - CSI (Control Sequence Introducer) sequences starting with ESC[
	// - OSC (Operating System Command) sequences
	// - Other common terminal control sequences.
	ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))" //nolint:lll

	// commandBash specifies the default command to use for bash shell execution.
	commandBash = "bash"

	// zeroValue is the default string value returned when numeric operations fail.
	zeroValue = "0.0"

	// uptimeZeroValue is the default string value returned when uptime cannot be determined.
	uptimeZeroValue = "00:00:00"
)

// ansiRegexp is a precompiled regular expression created from the ansi pattern.
// This is used for efficient stripping of ANSI escape sequences from strings.
var ansiRegexp = regexp.MustCompile(ansi)

// Common error definitions used throughout the package.
var (
	// ErrInvalidCommand indicates that command is invalid.
	ErrInvalidCommand = errors.New("invalid command")

	// ErrEmptyPID indicates that a process lookup returned an empty process ID.
	ErrEmptyPID = errors.New("empty process id")

	// ErrNotRunning indicates that the requested process is not currently running.
	ErrNotRunning = errors.New("process is not running")

	// ErrBashExecuteFailed indicates that a bash command execution failed.
	ErrBashExecuteFailed = errors.New("bash execute failed")
)

// Strip removes all ANSI escape sequences and trailing newline characters from a string.
// This is useful for cleaning up colored terminal output or formatted text.
//
// Parameters:
//   - str: The input string potentially containing ANSI codes
//
// Returns:
//   - A cleaned string with all ANSI sequences and trailing newlines removed
//
// Example:
//
//	cleaned := Strip("\033[32mHello\033[0m\n") // Returns "Hello".
func Strip(str string) string {
	return strings.TrimRight(ansiRegexp.ReplaceAllString(str, ""), "\n")
}

// Execute runs a system command and captures its output streams.
// It provides a convenient wrapper around exec.Command with integrated error handling.
//
// Parameters:
//   - name: The name/path of the command to execute (e.g. "ls", "/bin/bash")
//   - args: Variadic arguments to pass to the command (e.g. "-l", "-a")
//
// Returns:
//   - string: The combined stdout output of the command
//   - error:  Returns ErrBashExecuteFailed if stderr contains output,
//     or the original exec error if the command failed to run.
//     Returns nil if execution was successful with empty stderr.
//
// Behavior:
//   - Captures both stdout and stderr streams separately
//   - Considers any stderr output as an error condition
//   - Preserves the command's exit status error if present
//   - Trims no output - returned strings may contain trailing newlines
//
// Example:
//
//	output, err := Execute("ls", "-l", "/tmp")
//	if err != nil {
//	    // Handle error (either from stderr or command failure)
//	}
//	fmt.Println(output)
//
// Notes:
//   - For bash commands, consider using commandBash constant as name
//   - Command output is not stripped of ANSI codes (use Strip() separately)
//   - Not suitable for interactive commands requiring stdin.
func Execute(name string, args ...string) (string, error) {
	var stdout, stder bytes.Buffer

	cmd := exec.Command(name, args...)

	cmd.Stdout = &stdout
	cmd.Stderr = &stder

	err := cmd.Run()

	// Return stderr if present, even if command technically succeeded.
	if stder.String() != "" {
		return stdout.String(), fmt.Errorf("%w: %s", ErrBashExecuteFailed, stder.String())
	}

	return stdout.String(), err
}

// GetLargeFileList finds large files with specific extension in given path
// Parameters:
//   - path: directory to search
//   - ext: file extension to match (e.g. ".log")
//   - params: optional count parameter (default 20)
//
// Returns list of files or error if command fails.
func GetLargeFileList(path, mask string, params ...int) (string, error) {
	count := "20"
	if len(params) > 0 {
		count = strconv.Itoa(params[0])
	}

	args := []string{
		"-c", "ls " + path + " -hSRs | egrep '" + mask + "' | head -" + count,
	}

	return Execute(commandBash, args...)
}

// PidofByProcess retrieves the process ID (PID) of a running process by its name.
// It uses the system's 'pidof' command to find the PID of the specified process.
//
// Parameters:
//   - process: Name of the process to look up (e.g., "nginx", "java").
//     Should match the exact executable name.
//
// Returns:
//   - string: The PID of the process as a string if found.
//   - error:  May return:
//   - Original error from command execution if pidof fails
//   - ErrEmptyPID if process is not running or pidof returns empty
//   - Other system errors if command cannot be executed
//
// Behavior:
//   - Executes 'pidof <process>' command internally
//   - Automatically trims trailing newline from output
//   - Returns first PID if multiple instances are running (pidof behavior)
//   - Does not validate if the process is actually running beyond PID existence
//
// Example:
//
//	pid, err := PidofByProcess("nginx")
//	if err != nil {
//	    if errors.Is(err, ErrEmptyPID) {
//	        fmt.Println("Nginx is not running")
//	    } else {
//	        log.Fatalf("Error checking nginx: %v", err)
//	    }
//	}
//	fmt.Printf("Nginx PID: %s\n", pid)
//
// Notes:
//   - Requires pidof command to be available in system PATH
//   - For more advanced process lookups, see PidofByProcessAndParam
//   - Returned PID string may need conversion to int for numeric operations
//   - On systems with multiple process instances, consider using pgrep instead.
func PidofByProcess(process string) (string, error) {
	out, err := Execute("pidof", process)
	if err != nil {
		return "", err
	}

	if l := strings.Split(out, "\n"); len(l) > 0 {
		pid := l[0]
		if pid == "" {
			return "", ErrEmptyPID
		}

		return pid, nil
	}

	return "", ErrNotRunning
}

// PidofByProcessAndParam finds a process ID by process name and matching parameter.
// It executes a command pipeline: pgrep -af <process> | grep <param> to locate
// the specific process instance containing the given parameter.
//
// Parameters:
//   - process: The name of the process to search for (e.g. "java", "nginx")
//   - param: The parameter string to match in the process command line
//     (e.g. "--config=myapp.conf", "servername")
//
// Returns:
//   - string: The PID of the matching process
//   - error: ErrEmptyPID if process is found but PID is empty,
//     ErrNotRunning if no matching process is found,
//     or other errors from command execution
//
// Example:
//
//	pid, err := PidofByProcessAndParam("java", "-Dapp.name=myapp")
//	if err != nil {
//	    // handle error
//	}.
func PidofByProcessAndParam(process, param string) (string, error) {
	if process == "" || param == "" || string(process[0]) == "-" {
		return "", ErrInvalidCommand
	}

	if string(param[0]) == "-" {
		param = "\\" + param
	}

	cmd := fmt.Sprintf("pgrep -af %q | grep -v %q | grep %q | grep -o -e %q", process, " bash ", param, `^[0-9]*`)

	out, err := Execute(commandBash, "-c", cmd)
	if err != nil {
		return "", err
	}

	if l := strings.Split(out, "\n"); len(l) > 0 {
		pid := l[0]
		if pid == "" {
			return "", ErrEmptyPID
		}

		return pid, nil
	}

	return "", ErrNotRunning
}

// GetUptimeByPID retrieves the elapsed time since a process started using its PID.
// It executes the 'ps' command to get the process's running duration in format [[DD-]HH:]MM:SS.
//
// Parameters:
//   - pid: The process ID as a string (e.g., "12345"). Must be a valid running process ID.
//
// Returns:
//   - string: The process uptime in format:
//   - "MM:SS" for processes running <1 hour
//   - "HH:MM:SS" for processes running <1 day
//   - "DD-HH:MM:SS" for processes running multiple days
//   - uptimeZeroValue ("00:00:00") if the process is not found
//   - error:  Returns:
//   - Original error if 'ps' command execution fails
//   - nil if successful (even if process not found)
//
// Behavior:
//   - Uses 'ps -o etime= -p PID' command to get process duration
//   - Automatically trims whitespace and newlines from output
//   - Returns zero value (not error) if process doesn't exist
//   - Output format matches system 'ps' command behavior
//
// Example:
//
//	uptime, err := GetUptimeByPID("12345")
//	if err != nil {
//	    log.Printf("Failed to check uptime: %v", err)
//	}
//	fmt.Printf("Process running for: %s", uptime) // e.g. "01:23:45"
//
// Notes:
//   - Requires 'ps' command to be available in system PATH
//   - Unlike other functions, returns zero value rather than error for missing process
//   - For empty/zero uptime, check against uptimeZeroValue constant
//   - Uptime resolution is seconds (no milliseconds).
func GetUptimeByPID(pid string) (string, error) {
	uptime, err := Execute("ps", "-o", "etime=", "-p", pid)
	if err != nil {
		return uptimeZeroValue, err
	}

	return strings.Trim(uptime, " \n"), nil
}

// CPUPercentByPID retrieves the CPU usage percentage for a specific process.
// The percentage represents the process's total CPU utilization since its start.
//
// Parameters:
//   - pid: The process ID as a string (e.g., "1234"). Must be a valid running process.
//
// Returns:
//   - string: CPU usage percentage with "%" suffix (e.g., "25.5%")
//     Returns zeroValue + "%" ("0.0%") if:
//   - Process is not found
//   - Process is using 0% CPU
//   - Command fails
//   - error:  Error from command execution if ps command fails,
//     nil if successful (even if process shows 0% usage)
//
// Behavior:
//   - Uses 'ps S -p PID -o pcpu=' command to get CPU percentage
//   - The 'S' option includes child processes in calculation
//   - Automatically trims whitespace and appends "%" symbol
//   - Returns string formatted to one decimal place
//
// Example:
//
//	cpu, err := CPUPercentByPID("1234")
//	if err != nil {
//	    log.Printf("CPU check failed: %v", err)
//	}
//	fmt.Printf("CPU Usage: %s", cpu) // e.g. "75.3%"
//
// Notes:
//   - CPU percentage is relative to a single core (may exceed 100% on multicore systems)
//   - Requires 'ps' command to be available in system PATH
//   - For containerized processes, results may differ from host metrics
//   - Values are snapshots, not averages over time
//   - Consider using multiple samples for monitoring trending usage.
func CPUPercentByPID(pid string) (string, error) {
	cpu, err := Execute("ps", "S", "-p", pid, "-o", "pcpu=")
	if err != nil {
		return zeroValue + "%", err
	}

	return strings.Trim(cpu, " \n") + "%", nil
}

// MemPercentByPID retrieves the memory usage percentage for a specific process.
// The percentage represents the process's resident memory relative to total system memory.
//
// Parameters:
//   - pid: The process ID as a string (e.g., "5678"). Must be a valid running process.
//
// Returns:
//   - string: Memory usage percentage with "%" suffix (e.g., "4.2%")
//     Returns zeroValue + "%" ("0.0%") if:
//   - Process is not found
//   - Process uses 0% memory
//   - Command execution fails
//   - error:  Error from command execution if ps command fails,
//     nil if successful (even if process shows 0% usage)
//
// Behavior:
//   - Uses 'ps S -p PID -o pmem=' command to get memory percentage
//   - The 'S' option includes child processes in calculation
//   - Automatically trims whitespace and appends "%" symbol
//   - Returns string formatted to one decimal place
//
// Example:
//
//	mem, err := MemPercentByPID("5678")
//	if err != nil {
//	    log.Printf("Memory check failed: %v", err)
//	}
//	fmt.Printf("Memory Usage: %s", mem) // e.g. "2.8%"
//
// Notes:
//   - Percentage is relative to total physical memory (RAM)
//   - Does not include shared memory or swap usage
//   - Requires 'ps' command to be available in system PATH
//   - Values represent current snapshot, not averages over time.
func MemPercentByPID(pid string) (string, error) {
	mem, err := Execute("ps", "S", "-p", pid, "-o", "pmem=")
	if err != nil {
		return zeroValue + "%", err
	}

	return strings.Trim(mem, " \n") + "%", nil
}

// MemUsedByPID calculates the total resident memory usage of a process and its children in megabytes.
// It sums the RSS (Resident Set Size) memory of all process threads and converts to MB.
//
// Parameters:
//   - pid: The process ID as a string (e.g., "1234"). Must be a valid running process.
//
// Returns:
//   - string: Memory usage formatted with " MB" suffix (e.g., "24.5 MB")
//     Returns zeroValue + " MB" ("0.0 MB") if:
//   - Process is not found
//   - Process uses no resident memory
//   - Command execution fails
//   - error:  Error from command execution if the bash command fails,
//     nil if successful (even if memory usage is 0)
//
// Implementation Details:
//   - Uses bash command pipeline:
//     1. `ps -ylp PID` lists all threads with memory info
//     2. `awk` sums the RSS (column 8) and converts to MB (/1024)
//   - Automatically trims whitespace/newlines from output
//   - Adds " MB" suffix to clarify units
//
// Example:
//
//	memUsage, err := MemUsedByPID("1234")
//	if err != nil {
//	    log.Printf("Memory check failed: %v", err)
//	}
//	fmt.Printf("Memory used: %s", memUsage) // e.g. "45.2 MB"
//
// Notes:
//   - Measures physical RAM usage (RSS), not virtual memory
//   - Includes memory used by all process threads
//   - Values are in binary megabytes (MiB, 1024-based)
//   - Requires GNU ps and awk utilities.
func MemUsedByPID(pid string) (string, error) {
	args := []string{
		"-c", "ps -ylp " + pid + " | awk '{x += $8} END {print \"\" x/1024;}'",
	}

	mem, err := Execute(commandBash, args...)
	if err != nil {
		return zeroValue + " MB", err
	}

	return strings.Trim(mem, " \n") + " MB", nil
}

// MemUsed retrieves the total used system memory in megabytes (MB).
// It calculates the actively used memory excluding buffers/cache.
//
// Returns:
//   - string: Total used memory formatted with " MB" suffix (e.g., "2048 MB")
//     Returns zeroValue + " MB" ("0.0 MB") if:
//   - Command execution fails
//   - Unable to parse memory information
//   - error:  Error from command execution if the bash command fails,
//     nil if successful
//
// Implementation Details:
//   - Uses bash command pipeline:
//     1. `free` command to get memory statistics
//     2. `awk` extracts the used memory value (column 3 from second line)
//     3. Converts from kilobytes to megabytes (/1024)
//     4. Formats as integer (%.0f) to remove decimal places
//   - Automatically trims whitespace/newlines from output
//   - Adds " MB" suffix to clarify units
//
// Example:
//
//	usedMem, err := MemUsed()
//	if err != nil {
//	    log.Printf("Failed to get system memory: %v", err)
//	}
//	fmt.Printf("System memory used: %s", usedMem) // e.g. "3752 MB"
//
// Notes:
//   - Measures actual used memory excluding buffers/cache
//   - Values are in binary megabytes (MiB, 1024-based)
//   - Requires GNU free and awk utilities
//   - Represents system-wide memory usage, not per-process
//   - Consider using /proc/meminfo for more detailed breakdown.
func MemUsed() (string, error) {
	args := []string{
		"-c", "free | awk 'NR==2 { printf(\"%.0f\", $3/1024); }'",
	}

	mem, err := Execute(commandBash, args...)
	if err != nil {
		return zeroValue + " MB", err
	}

	return strings.Trim(mem, " \n") + " MB", nil
}

// MemAvail retrieves the total available system memory in megabytes (MB).
// This represents the physical RAM available for new processes, excluding buffers/cache.
//
// Returns:
//   - string: Total available memory formatted with " MB" suffix (e.g., "8192 MB")
//     Returns zeroValue + " MB" ("0.0 MB") if:
//   - Command execution fails
//   - Unable to parse memory information
//   - error:  Error from command execution if the bash command fails,
//     nil if successful
//
// Implementation Details:
//   - Uses bash command pipeline:
//     1. `free` command to get memory statistics
//     2. `awk` extracts the total memory value (column 2 from second line)
//     3. Converts from kilobytes to megabytes (/1024)
//     4. Formats as integer (%.0f) for clean output
//   - Automatically trims whitespace/newlines from output
//   - Adds " MB" suffix to clarify units
//
// Example:
//
//	availMem, err := MemAvail()
//	if err != nil {
//	    log.Printf("Failed to get available memory: %v", err)
//	}
//	fmt.Printf("Available system memory: %s", availMem) // e.g. "16384 MB"
//
// Notes:
//   - Measures physical RAM, not including swap space
//   - Values are in binary megabytes (MiB, 1024-based)
//   - Requires GNU free and awk utilities
//   - Represents system-wide available memory
//   - For accurate container memory limits, check cgroup settings.
func MemAvail() (string, error) {
	args := []string{
		"-c", "free | awk 'NR==2 { printf(\"%.0f\", $2/1024); }'",
	}

	mem, err := Execute(commandBash, args...)
	if err != nil {
		return zeroValue + " MB", err
	}

	return strings.Trim(mem, " \n") + " MB", nil
}
