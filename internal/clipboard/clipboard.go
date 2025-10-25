package clipboard

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
)

// ErrUnavailable indicates that no clipboard provider is accessible on this platform.
var ErrUnavailable = errors.New("clipboard unavailable")

// Provider represents something that can write text to the clipboard.
type Provider interface {
	Write(string) error
}

// ProviderFunc allows plain functions to satisfy the Provider interface.
type ProviderFunc func(string) error

// Write implements Provider.
func (f ProviderFunc) Write(text string) error { return f(text) }

var current Provider = systemProvider{}

// Copy writes text to the clipboard using the active provider.
func Copy(text string) error {
	return current.Write(text)
}

// SetProvider swaps the clipboard provider, primarily for testing. Passing nil
// restores the default system-backed provider.
func SetProvider(p Provider) {
	if p == nil {
		current = systemProvider{}
		return
	}
	current = p
}

type systemProvider struct{}

func (systemProvider) Write(text string) error {
	var lastErr error
	for _, spec := range candidateCommands() {
		if _, err := exec.LookPath(spec.bin); err != nil {
			lastErr = err
			continue
		}

		cmd := exec.Command(spec.bin, spec.args...)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			lastErr = err
			continue
		}

		if err := cmd.Start(); err != nil {
			stdin.Close()
			lastErr = err
			continue
		}

		if _, err := io.WriteString(stdin, text); err != nil {
			stdin.Close()
			cmd.Wait()
			return err
		}
		stdin.Close()

		if err := cmd.Wait(); err == nil {
			return nil
		}
		lastErr = err
	}

	if lastErr == nil {
		return ErrUnavailable
	}
	return fmt.Errorf("%w: %v", ErrUnavailable, lastErr)
}

type cmdSpec struct {
	bin  string
	args []string
}

func candidateCommands() []cmdSpec {
	switch runtime.GOOS {
	case "darwin":
		return []cmdSpec{{bin: "pbcopy"}}
	case "windows":
		return []cmdSpec{{bin: "clip"}}
	default:
		return []cmdSpec{
			{bin: "wl-copy"},
			{bin: "xclip", args: []string{"-selection", "clipboard"}},
			{bin: "xsel", args: []string{"-b"}},
		}
	}
}
