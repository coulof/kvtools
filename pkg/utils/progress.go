package utils

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Progress manages CLI status output, spinning or printing steps.
type Progress struct {
	mu      sync.Mutex
	quiet   bool
	writer  io.Writer
	active  bool
	stopCh  chan struct{}
	doneCh  chan struct{}
	message string
}

// NewProgress creates a new Progress instance.
func NewProgress(quiet bool) *Progress {
	return &Progress{
		quiet:  quiet,
		writer: os.Stderr,
	}
}

// Start starts a simple terminal spinner with a message.
func (p *Progress) Start(msg string) {
	if p.quiet {
		return
	}
	p.mu.Lock()
	if p.active {
		p.mu.Unlock()
		p.Update(msg)
		return
	}
	p.active = true
	p.message = msg
	p.stopCh = make(chan struct{})
	p.doneCh = make(chan struct{})
	p.mu.Unlock()

	go p.spin()
}

func (p *Progress) spin() {
	spinnerChars := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
	idx := 0
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCh:
			// Clear line
			fmt.Fprintf(p.writer, "\r\033[K")
			close(p.doneCh)
			return
		case <-ticker.C:
			p.mu.Lock()
			msg := p.message
			p.mu.Unlock()
			fmt.Fprintf(p.writer, "\r\033[36m%c\033[0m %s...", spinnerChars[idx], msg)
			idx = (idx + 1) % len(spinnerChars)
		}
	}
}

// Update updates the current spinner message.
func (p *Progress) Update(msg string) {
	if p.quiet {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.message = msg
}

// Success stops the spinner and prints a checkmark with a success message.
func (p *Progress) Success(msg string) {
	p.stop()
	if !p.quiet {
		fmt.Fprintf(p.writer, "\r\033[32m✔\033[0m %s\n", msg)
	}
}

// Warn stops the spinner and prints a warning message.
func (p *Progress) Warn(msg string) {
	p.stop()
	if !p.quiet {
		fmt.Fprintf(p.writer, "\r\033[33m▲\033[0m %s\n", msg)
	}
}

// Error stops the spinner and prints an error message.
func (p *Progress) Error(msg string) {
	p.stop()
	if !p.quiet {
		fmt.Fprintf(p.writer, "\r\033[31m✖\033[0m %s\n", msg)
	}
}

// Info prints an informative line directly.
func (p *Progress) Info(msg string) {
	p.stop()
	if !p.quiet {
		fmt.Fprintf(p.writer, "  \033[90m→\033[0m %s\n", msg)
	}
}

func (p *Progress) stop() {
	p.mu.Lock()
	if !p.active {
		p.mu.Unlock()
		return
	}
	p.active = false
	close(p.stopCh)
	p.mu.Unlock()
	<-p.doneCh
}
