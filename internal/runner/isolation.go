package runner

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func RunIsolation(ctx context.Context, execPath string, inputData string, time_limit, memory_limit, boxID int, execArg []string) (*TaskResult, error) {
	boxStr := strconv.Itoa(boxID)

	initCmd := exec.CommandContext(ctx, "sudo", "isolate", "--box-id="+boxStr, "--init")

	path, err := initCmd.Output()

	if err != nil {
		return nil, err
	}

	boxDir := strings.TrimSpace(string(path)) + "/box"

	defer exec.CommandContext(ctx, "sudo", "isolate", "--box-id="+boxStr, "--cleanup").Run()

	inputPath := filepath.Join(boxDir, "input.in")

	if err := os.WriteFile(inputPath, []byte(inputData), 0644); err != nil {
		return nil, fmt.Errorf("error with write input.in %w", err)
	}

	fileName := filepath.Base(execPath)

	targetPath := filepath.Join(boxDir, fileName)

	readBytes, err := os.ReadFile(execPath)

	if err != nil {
		return nil, fmt.Errorf("error with read %s : %w", execPath, err)
	}

	if err := os.WriteFile(targetPath, readBytes, 0755); err != nil {
		return nil, fmt.Errorf("error with write bytes into %s : %w", targetPath, err)
	}

	metaPath := filepath.Join(os.TempDir(), fmt.Sprintf("isolate_meta_%d", boxID))
	defer os.Remove(metaPath)

	args := []string{
		"isolate",
		"--box-id=" + boxStr,
		fmt.Sprintf("--time=%.3f", float64(time_limit)),
		fmt.Sprintf("--wall-time=%.3f", float64(time_limit)*2),
		fmt.Sprintf("--mem=%d", memory_limit),
		"--dir=/usr/bin",
		"--dir=/usr/lib",
		"--dir=/lib64",
		"--dir=/etc",
		"--processes=1",
		"--stdin=input.in",
		"--stdout=output.out",
		"--stderr=error.err",
		"--meta=" + metaPath,
		"--run",
		"--",
	}

	args = append(args, execArg...)

	cmd := exec.CommandContext(ctx, "sudo", args...)

	_ = cmd.Run()

	stdoutBytes, _ := os.ReadFile(filepath.Join(boxDir, "output.out"))
	stderrBytes, _ := os.ReadFile(filepath.Join(boxDir, "error.err"))

	res, err := parseMetaFile(metaPath)

	if err != nil {
		return nil, fmt.Errorf("metafile parse error : %w", err)
	}

	res.Stdout = string(stdoutBytes)
	res.Stderr = string(stderrBytes)

	return res, nil
}

func parseMetaFile(metaPath string) (*TaskResult, error) {
	res := &TaskResult{Verdict: "OK"}

	file, err := os.Open(metaPath)

	if err != nil {
		return nil, fmt.Errorf("file open err at parseMetaFile : %w", err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	var status, message string
	var exitcode, exitsig int

	for scanner.Scan() {
		line := scanner.Text()

		content := strings.SplitN(line, ":", 2)

		if len(content) != 2 {
			continue
		}

		key, val := content[0], content[1]

		switch key {
		case "time":
			timeSeconds, _ := strconv.ParseFloat(val, 64)
			res.TimeUsed = int(timeSeconds * 1000)
		case "max-rss":
			memoryKb, _ := strconv.Atoi(val)
			res.MemoryUsed = (memoryKb + 1023) / 1024
		case "status":
			status = val
		case "exitcode":
			exitcode, _ = strconv.Atoi(val)
		case "message":
			message = val
		case "exitsig":
			exitsig, _ = strconv.Atoi(val)
		}

	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	//Формирование вердикта
	if status != "" {
		switch status {
		case "TO":
			res.Verdict = "TLE"
		case "MEM":
			res.Verdict = "MLE"
		case "FO":
			res.Verdict = "OLE"
		case "SG":
			if (exitsig == 11 || exitsig == 9) && strings.Contains(message, "Memory limit") {
				res.Verdict = "MLE"
			} else {
				res.Verdict = "RE"
			}
		case "RE":
			res.Verdict = "RE"
		default:
			res.Verdict = "SE"
		}
	} else if exitcode != 0 {
		res.Verdict = "RE"
	}

	return res, nil

}
