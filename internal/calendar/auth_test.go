package calendar

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

type errorOpener struct {
	err error
}

func (opener errorOpener) Open(url string) error {
	return opener.err
}

func TestFetchTokenTimesOut(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()

	_, err := fetchToken(ctx, &oauth2.Config{}, &fakeOpener{})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timed out") && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected timeout-ish error, got %v", err)
	}
}

func TestFetchTokenContinuesAfterOpenBrowserError(t *testing.T) {
	openErr := errors.New("permission denied")
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()

	_, err := fetchToken(ctx, &oauth2.Config{}, errorOpener{err: openErr})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if errors.Is(err, openErr) {
		t.Fatalf("expected opener error to be non-fatal, got %v", err)
	}
}
