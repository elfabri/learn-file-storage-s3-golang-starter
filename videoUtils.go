package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

func GetVideoAspectRatio(filePath string) (string, error) {
    var out bytes.Buffer
    cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
    cmd.Stdout = &out

    err := cmd.Run()
    if err != nil {
        return "", fmt.Errorf("Invalid Path: %s; error: %v\n", filePath, err)
    }

    type values struct {
        Height float32 `json:"height"`
        Width float32 `json:"width"`
    }

    type streams struct {
        V []values `json:"streams"`
    }

    var s streams

    err = json.Unmarshal(out.Bytes(), &s)
    if err != nil {
        return "", fmt.Errorf("Error unmarshalling video aspect ratio: %v\n", err)
    }

    aspRat := float32(s.V[0].Width) / float32(s.V[0].Height)
    horizontal := float32(16) / float32(9)          // 1.7777
    vertical := float32(9) / float32(16)            // 0.5625
    tol := float32(0.01)

    if aspRat <= (horizontal + tol) && aspRat >= (horizontal - tol) {
        return "landscape", nil
    }
    if aspRat <= (vertical + tol) && aspRat >= (vertical - tol) {
        return "portrait", nil
    }

    return "other", nil
}

func ProcessVideoForFastStart(filePath string) (string, error) {
    // creates and returns a new path to a file with “fast start” encoding
    outFilePath := filePath + ".processing"

    var out bytes.Buffer
    cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outFilePath)
    cmd.Stdout = &out

    err := cmd.Run()
    if err != nil {
        return "", fmt.Errorf("Invalid Path: %s; error: %v\n", filePath, err)
    }

    return outFilePath, nil
}
