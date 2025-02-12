package main

import (
	"fmt"
	"testing"
)

func TestThis(t *testing.T) {
    fmt.Println("Running aspect ratio test:")
    fileP1 := "./samples/boots-video-horizontal.mp4"
    fileP2 := "./samples/boots-video-vertical.mp4"

    aspRat, err := GetVideoAspectRatio(fileP1)
    if err != nil {
        t.Errorf("Unspected error: %v\n", err)
    }

    if aspRat != "landscape" {
        t.Errorf("should have been landscape: %v\n", aspRat)
    }

    aspRat, err = GetVideoAspectRatio(fileP2)
    if err != nil {
        t.Errorf("Unspected error: %v\n", err)
    }

    if aspRat != "portrait" {
        t.Errorf("should have been portrait: %v\n", aspRat)
    }

    fmt.Println("Running fast start encoding test:")

    fastPath, err := ProcessVideoForFastStart(fileP1)
    if err != nil {
        t.Errorf("Unspected error: %v\n", err)
    }

    if fastPath != fileP1+".processing" {
        t.Errorf("wrong name to file: %v\n", fastPath)
    }

    fastPath, err = ProcessVideoForFastStart(fileP2)
    if err != nil {
        t.Errorf("Unspected error: %v\n", err)
    }

    if fastPath != fileP2+".processing" {
        t.Errorf("wrong name to file: %v\n", fastPath)
    }
}
