package pogoicon

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
)

type Options struct {
	FFMPEG      string
	ScaleWidth  int
	ScaleHeight int
}

func newError(err error, log []byte) *Error {
	return &Error{
		Err: err,
		Log: log,
	}
}

type Error struct {
	Err error
	Log []byte
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func (e *Error) LogString() string {
	return string(e.Log)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func Generate(ctx context.Context, pokemon io.Reader, background io.Reader, options Options) (io.Reader, error) {
	backgroundPipeReader, backgroundPipeWriter, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("error creating background pipe: %w", err)
	}
	pokemonPipeReader, pokemonPipeWriter, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("error creating pokemon pipe: %w", err)
	}
	go func() {
		defer backgroundPipeWriter.Close()
		if _, err := io.Copy(backgroundPipeWriter, background); err != nil {
			slog.ErrorContext(ctx, "error copying to background pipe", slog.Any("err", err))
		}
	}()
	go func() {
		defer pokemonPipeWriter.Close()
		if _, err := io.Copy(pokemonPipeWriter, pokemon); err != nil {
			slog.ErrorContext(ctx, "error copying to pokemon pipe", slog.Any("err", err))
		}
	}()

	cmd := exec.CommandContext(ctx, options.FFMPEG,
		"-i", "pipe:0",
		"-i", "pipe:3",
		"-filter_complex", fmt.Sprintf(
			"[1:v]scale=%d:%d[overlay];[0:v][overlay]overlay=(main_w-overlay_w)/2:(main_h-overlay_h)/2",
			options.ScaleWidth, options.ScaleHeight,
		),
		"-c:v", "png", // PNG encoding
		"-pix_fmt", "rgba", // Ensure proper PNG format with transparency
		"-f", "image2pipe", // Force output as image2pipe format
		"pipe:1", // Output to stdout
	)

	cmd.Stdin = backgroundPipeReader
	cmd.ExtraFiles = []*os.File{pokemonPipeReader}

	var buf bytes.Buffer
	cmd.Stdout = &buf
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	// Run FFmpeg
	if err = cmd.Run(); err != nil {
		return nil, newError(err, stderrBuf.Bytes())
	}

	_, _ = io.Copy(os.Stdout, &stderrBuf)

	return &buf, nil
}
