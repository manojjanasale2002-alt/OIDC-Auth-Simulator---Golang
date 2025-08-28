package main

import (
    "context"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "fmt"
    "html/template"
    "log"
    "math/big"
    "net/http"
    "os"
    "strings"
    "time"

    oidc "github.com/coreos/go-oidc/v3/oidc"
    "golang.org/x/oauth2"
)

// Simple in-memory session store for demo only.
var sessions = map[string]map[string]string{}

func randString(n int) string {
    const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    b := make([]byte, n)
    for i := range b {
        num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
        b[i] = letters[num.Int64()]
    }
    return string(b)
}

func base64URLEncode(b []byte) string { return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=") }

func main() {
    // Config via env
    issuer := getenv("OIDC_ISSUER", "http://localhost:8080/realms/demo-realm")
    clientID := getenv("OIDC_CLIENT_ID", "go-web-app")
    clientSecret := os.Getenv("OIDC_CLIENT_SECRET") // empty for public client
    redirectURL := getenv("OIDC_REDIRECT_URL", "http://localhost:3000/callback")
    addr := getenv("ADDR", ":3000")

    ctx := context.Background()

    provider, err := oidc.NewProvider(ctx, issuer)
    if err != nil {
        log.Fatalf("discover provider: %v", err)
    }

    // Oauth2 config
    conf := oauth2.Config{
        ClientID:     clientID,
        ClientSecret: clientSecret,
        RedirectURL:  redirectURL,
        Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
        Endpoint:     provider.Endpoint(),
    }

    // Verifier for ID Token signature + issuer + audience
    verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

    // Routes
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        tmpl := template.Must(template.New("index").Parse(`
		<html>
        <head>
          <style>
            body {
              font-family: Arial, sans-serif;
              background: #f4f6f9;
              display: flex;
              justify-content: center;
              align-items: center;
              height: 100vh;
              margin: 0;
            }
            .container {
              text-align: center;
              background: white;
              padding: 40px;
              border-radius: 12px;
              box-shadow: 0 4px 12px rgba(0,0,0,0.15);
            }
            h2 {
              color: #333;
              margin-bottom: 20px;
            }
            a.login-btn {
              display: inline-block;
              padding: 12px 24px;
              font-size: 16px;
              color: white;
              background: #4a90e2;
              border-radius: 6px;
              text-decoration: none;
              transition: background 0.3s ease;
            }
            a.login-btn:hover {
              background: #357ab8;
            }
          </style>
        </head>
        <body>
          <div class="container">
            <h2>🔑 Keycloak Broker Demo</h2>
            <p><a class="login-btn" href="/login">Login with Keycloak</a></p>
          </div>
        </body>
        </html>`))
        _ = tmpl.Execute(w, nil)
    })

    http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
        // CSRF state
        state := randString(24)
        // PKCE
        codeVerifier := base64URLEncode([]byte(randString(64)))
        h := sha256.Sum256([]byte(codeVerifier))
        codeChallenge := base64URLEncode(h[:])

        // Save in-memory by a cookie key
        sid := randString(24)
        sessions[sid] = map[string]string{"state": state, "code_verifier": codeVerifier}
        http.SetCookie(w, &http.Cookie{Name: "sid", Value: sid, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})

        // If you want to force Google/Azure button immediately, add kc_idp_hint=google|azure
        authURL := conf.AuthCodeURL(state,
            oauth2.SetAuthURLParam("code_challenge", codeChallenge),
            oauth2.SetAuthURLParam("code_challenge_method", "S256"),
            // oauth2.SetAuthURLParam("kc_idp_hint", "google"), // or "azure"
        )
        http.Redirect(w, r, authURL, http.StatusFound)
    })

    http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
        // Validate state + exchange code
        c, _ := r.Cookie("sid")
        sess := map[string]string{}
        if c != nil {
            sess = sessions[c.Value]
        }
        if r.URL.Query().Get("state") == "" || sess["state"] != r.URL.Query().Get("state") {
            http.Error(w, "invalid state", http.StatusBadRequest)
            return
        }

        // Exchange code for tokens (with PKCE)
        tok, err := conf.Exchange(r.Context(), r.URL.Query().Get("code"),
            oauth2.SetAuthURLParam("code_verifier", sess["code_verifier"]))
        if err != nil {
            http.Error(w, fmt.Sprintf("token exchange failed: %v", err), http.StatusBadRequest)
            return
        }

        rawIDToken, ok := tok.Extra("id_token").(string)
        if !ok {
            http.Error(w, "no id_token in token response", http.StatusBadRequest)
            return
        }

        // Verify ID token
        idt, err := verifier.Verify(r.Context(), rawIDToken)
        if err != nil {
            http.Error(w, fmt.Sprintf("verify id_token: %v", err), http.StatusBadRequest)
            return
        }

        // Parse claims into a map
        claims := map[string]any{}
        if err := idt.Claims(&claims); err != nil {
            http.Error(w, fmt.Sprintf("parse claims: %v", err), http.StatusBadRequest)
            return
        }

        tmpl := template.Must(template.New("callback").Parse(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Keycloak Callback</title>
  <style>
    body {
      font-family: "Segoe UI", sans-serif;
      background: #f4f6f9;
      margin: 0;
      padding: 20px;
    }
    .card {
      background: white;
      padding: 20px;
      border-radius: 12px;
      box-shadow: 0 4px 12px rgba(0,0,0,0.1);
      max-width: 800px;
      margin: auto;
    }
    h3 { color: #2c3e50; }
    textarea, pre {
      width: 100%;
      background: #272822;
      color: #f8f8f2;
      font-family: monospace;
      border-radius: 6px;
      padding: 12px;
      overflow-x: auto;
    }
    .label {
      font-weight: bold;
      margin-top: 12px;
      display: block;
    }
    a {
      color: #3498db;
      text-decoration: none;
    }
    a:hover { text-decoration: underline; }
  </style>
</head>
<body>
  <div class="card">
    <h3>Token Exchange Success 🎉</h3>
    <div>
      <span class="label">Access Token (truncated):</span>
      <textarea rows="3">{{.AccessToken}}...</textarea>
    </div>
    <div>
      <span class="label">ID Token:</span>
      <textarea rows="6">{{.RawIDToken}}</textarea>
    </div>
    <div>
      <span class="label">ID Token Claims:</span>
      <pre>{{.Claims}}</pre>
    </div>
    <p>Paste the ID token at <a href="https://jwt.io" target="_blank">jwt.io</a> to inspect.</p>
  </div>
</body>
</html>`))

data := map[string]any{
    "AccessToken": tok.AccessToken[:32],
    "RawIDToken":  rawIDToken,
    "Claims":      claims,
}

_ = tmpl.Execute(w, data)
    })

    srv := &http.Server{Addr: addr, ReadHeaderTimeout: 5 * time.Second}
    log.Printf("listening on %s", addr)
    log.Fatal(srv.ListenAndServe())
}

func getenv(k, def string) string {
    if v := os.Getenv(k); v != "" { return v }
    return def
}