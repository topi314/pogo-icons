package pogoicons

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const (
	ScaleWidth  = 400
	ScaleHeight = -1
)

func generate() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	background, err := os.Open("background.png")
	if err != nil {
		panic(err)
	}
	defer background.Close()

	rs, err := http.Get("https://www.pokewiki.de/images/e/ed/Hauptartwork_742.png")
	if err != nil {
		panic(err)
	}
	defer rs.Body.Close()

	bgPipeReader, bgPipeWriter, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	overlayPipeReader, overlayPipeWriter, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	go func() {
		defer bgPipeWriter.Close()
		if _, err := io.Copy(bgPipeWriter, background); err != nil {
			panic(err)
		}
	}()
	go func() {
		defer overlayPipeWriter.Close()
		if _, err := io.Copy(overlayPipeWriter, rs.Body); err != nil {
			panic(err)
		}
	}()

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", "pipe:0",
		"-i", "pipe:3",
		"-filter_complex", fmt.Sprintf(
			"[1:v]scale=%d:%d[overlay];[0:v][overlay]overlay=(main_w-overlay_w)/2:(main_h-overlay_h)/2",
			ScaleWidth, ScaleHeight,
		),
		"-c:v", "png", // PNG encoding
		"-pix_fmt", "rgba", // Ensure proper PNG format with transparency
		"-f", "image2pipe", // Force output as image2pipe format
		"pipe:1", // Output to stdout
	)

	cmd.Stdin = bgPipeReader
	cmd.ExtraFiles = []*os.File{overlayPipeReader}

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr

	// Run FFmpeg
	if err = cmd.Run(); err != nil {
		fmt.Println("FFmpeg error:", err)
		return
	}

	fmt.Println("Image processing complete! Saved as output.png")

	os.WriteFile("output.png", buf.Bytes(), 0644)
}
