package leetcode

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

const (
	leetcodeLoginURL = "https://leetcode.com/accounts/login/"
)

func (c *Client) browserSessionAuth(ctx context.Context) error {
	b, err := detectChromiumBrowser()
	if err != nil {
		return fmt.Errorf("%s", unsupportedBrowserMessage())
	}

	profileDir := filepath.Join(c.dataDir, "browser-profile")
	if err := os.MkdirAll(profileDir, 0o700); err != nil {
		return fmt.Errorf("create browser profile: %w", err)
	}

	fmt.Printf("Opening %s for LeetCode authentication...\n", b.Name)
	fmt.Println("Log in to LeetCode in the browser window. Leetgo will close it after the Session is connected.")

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx,
		chromedp.ExecPath(b.Path),
		chromedp.UserDataDir(profileDir),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
	)
	defer cancelAlloc()

	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx,
		chromedp.WithLogf(func(string, ...any) {}),
		chromedp.WithErrorf(func(string, ...any) {}),
	)
	defer cancelBrowser()

	authCtx, cancelAuth := context.WithTimeout(browserCtx, 5*time.Minute)
	defer cancelAuth()

	if err := chromedp.Run(authCtx,
		clearLeetCodeCookies(),
		chromedp.Navigate(leetcodeLoginURL),
	); err != nil {
		return fmt.Errorf("open LeetCode login: %w", err)
	}

	session, err := waitForSignedInSession(authCtx)
	if err != nil {
		return err
	}
	c.session = session
	if err := c.saveSession(); err != nil {
		return fmt.Errorf("save Session: %w", err)
	}

	cancelAuth()
	cancelBrowser()
	cancelAlloc()

	fmt.Println("LeetCode Session connected.")
	fmt.Println("Browser closed. You can now run `leetgo submit .`.")
	return nil
}

func clearLeetCodeCookies() chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		for _, name := range []string{"csrftoken", "LEETCODE_SESSION"} {
			if err := network.DeleteCookies(name).WithURL(leetcodeBaseURL).Do(ctx); err != nil {
				return err
			}
			if err := network.DeleteCookies(name).WithDomain(".leetcode.com").WithPath("/").Do(ctx); err != nil {
				return err
			}
			if err := network.DeleteCookies(name).WithDomain("leetcode.com").WithPath("/").Do(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

func waitForSignedInSession(ctx context.Context) (*Session, error) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("authentication timed out")
		case <-ticker.C:
			var cookies []*network.Cookie
			if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
				var err error
				cookies, err = network.GetCookies().WithURLs([]string{leetcodeBaseURL}).Do(ctx)
				return err
			})); err != nil {
				return nil, fmt.Errorf("read browser cookies: %w", err)
			}

			session := sessionFromCookies(cookies)
			if session != nil {
				signedIn, err := browserShowsSignedIn(ctx)
				if err != nil {
					continue
				}
				if signedIn {
					return session, nil
				}
			}
		}
	}
}

func browserShowsSignedIn(ctx context.Context) (bool, error) {
	var currentURL string
	if err := chromedp.Run(ctx, chromedp.Location(&currentURL)); err != nil {
		return false, err
	}
	if !strings.Contains(currentURL, "leetcode.com") {
		return false, nil
	}

	var signedIn bool
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`(() => {
			const signedOut = document.querySelector('a[href*="/accounts/login"], a[href*="/accounts/signup"]') || /\bSign\s*In\b|\bSign\s*Up\b/i.test(document.body ? document.body.innerText : '');
			const signedIn = document.querySelector('[data-cy*="user"], [class*="user"], button[class*="avatar"], a[href^="/u/"], a[href^="/profile/"], img[src*="avatar"], img[alt*="avatar"]');
			return Boolean(!signedOut && signedIn);
		})()`, &signedIn),
	)
	if err != nil {
		return false, err
	}
	return signedIn, nil
}

func sessionFromCookies(cookies []*network.Cookie) *Session {
	var csrf string
	var sessionID string
	var expiresAt *time.Time

	for _, cookie := range cookies {
		switch cookie.Name {
		case "csrftoken":
			csrf = cookie.Value
		case "LEETCODE_SESSION":
			sessionID = cookie.Value
			if cookie.Expires > 0 {
				expiry := time.Unix(int64(cookie.Expires), 0)
				expiresAt = &expiry
			}
		}
	}

	if csrf == "" || sessionID == "" {
		return nil
	}
	return &Session{
		CSRFToken: csrf,
		SessionID: sessionID,
		ExpiresAt: expiresAt,
		Source:    "chromium-cdp",
	}
}
