package calendar

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gcal "google.golang.org/api/calendar/v3"
)

func AuthorizeDefault(ctx context.Context) error {
	config, err := LoadConfigFromEnv()
	if err != nil {
		return err
	}
	return Authorize(ctx, config, SystemOpener{})
}

func Authorize(ctx context.Context, config Config, opener Opener) error {
	credentials, err := os.ReadFile(config.CredentialsPath)
	if err != nil {
		return fmt.Errorf("read google credentials: %w", err)
	}

	oauthConfig, err := google.ConfigFromJSON(credentials, gcal.CalendarReadonlyScope)
	if err != nil {
		return fmt.Errorf("parse google credentials: %w", err)
	}

	token, err := fetchToken(ctx, oauthConfig, opener)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(config.TokenPath), 0755); err != nil {
		return fmt.Errorf("create token directory: %w", err)
	}
	if err := saveToken(config.TokenPath, token); err != nil {
		return err
	}
	return nil
}

func fetchToken(ctx context.Context, oauthConfig *oauth2.Config, opener Opener) (*oauth2.Token, error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	state := randomState()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("start oauth callback listener: %w", err)
	}
	defer listener.Close()

	oauthConfig.RedirectURL = "http://" + listener.Addr().String() + "/oauth2callback"
	server := &http.Server{
		ReadHeaderTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if request.URL.Path != "/oauth2callback" {
				http.NotFound(writer, request)
				return
			}
			if request.URL.Query().Get("state") != state {
				errCh <- fmt.Errorf("oauth state mismatch")
				http.Error(writer, "authorization failed", http.StatusBadRequest)
				return
			}
			code := request.URL.Query().Get("code")
			if code == "" {
				errCh <- fmt.Errorf("oauth callback missing code")
				http.Error(writer, "authorization failed", http.StatusBadRequest)
				return
			}
			codeCh <- code
			_, _ = writer.Write([]byte("Google Calendar authorization complete. You can close this tab."))
		}),
	}
	defer server.Shutdown(context.Background())

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Printf("Opening Google Calendar authorization in your browser:\n%s\n", authURL)
	if err := opener.Open(authURL); err != nil {
		fmt.Printf("Could not open browser automatically: %v\nCopy the URL above into your browser to continue.\n", err)
	}

	select {
	case code := <-codeCh:
		token, err := oauthConfig.Exchange(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("exchange authorization code: %w", err)
		}
		return token, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("authorization timed out waiting for Google callback")
		}
		return nil, ctx.Err()
	}
}

func saveToken(path string, token *oauth2.Token) error {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open token file: %w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(token); err != nil {
		return fmt.Errorf("write token file: %w", err)
	}
	return nil
}

func randomState() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes[:])
}
