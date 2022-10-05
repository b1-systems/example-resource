package main

import (
  "encoding/json"
  "log"
  "net/http"
  "os"
  "path/filepath"
  "strings"
  "github.com/coreos/go-oidc/v3/oidc"
  "golang.org/x/net/context"
  "golang.org/x/oauth2"
  "github.com/golang-jwt/jwt/v4"
  "gopkg.in/ini.v1"
)

var (
  clientName = "example-resource"
  clientID = ""
  clientSecret = ""
  providerUrl = ""
  listenAddress = ""
)

func readIni() {
  ex, err := os.Executable()

  if err != nil {
    panic(err)
  }

  cfg, err := ini.Load(filepath.Join(filepath.Dir(ex), clientName + ".ini"))

  if err != nil {
    panic(err)
  }

  cs := cfg.Section(clientName)

  clientID = cs.Key("clientID").String()

  if clientID == "" {
    log.Fatal(clientName + ".ini does not specify clientID")
    os.Exit(1)
  }

  clientSecret = cs.Key("clientSecret").String()

  if clientSecret == "" {
    log.Fatal(clientName + ".ini does not specify clientSecret")
    os.Exit(1)
  }

  providerUrl = cs.Key("providerUrl").String()

  if providerUrl == "" {
    log.Fatal(clientName + ".ini does not specify providerUrl")
    os.Exit(1)
  }

  listenAddress = cs.Key("listenAddress").String()

  if listenAddress == "" {
    log.Fatal(clientName + ".ini does not specify listenAddress")
    os.Exit(1)
  }

  log.Printf(
    "Read configuration:\n" +
    " clientID = %s\n" +
    " clientSecret = %s\n" +
    " providerUrl = %s\n" +
    " listenAddress = %s\n",
    clientID,
    "*REDACTED*",
    providerUrl,
    listenAddress,
  )
}

func main() {
  readIni()

  ctx := context.Background()
  provider, err := oidc.NewProvider(ctx, providerUrl)

  if err != nil {
    log.Fatal(err)
    os.Exit(1)
  }

  config := oauth2.Config{
    ClientID: clientID,
    ClientSecret: clientSecret,
    Endpoint: provider.Endpoint(),
    Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
  }

  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    auth_header := r.Header.Get("Authorization")

    if auth_header == "" {
      log.Printf("Authorization header missing from request")
      http.Error(w, "Bad request", http.StatusBadRequest)
    } else if !strings.HasPrefix(auth_header, "Bearer") {
      log.Printf("Authorization header is not a Bearer token")
      http.Error(w, "Bad request", http.StatusBadRequest)
    } else {
      access_token := strings.TrimPrefix(auth_header, "Bearer ")

      tokenSource := config.TokenSource(context.Background(), &oauth2.Token{
        AccessToken: access_token,
        TokenType: "Bearer",
      })

      if _, err := tokenSource.Token(); err != nil {
        log.Printf("Invalid access_token: ", err)
        http.Error(w, "Bad request", http.StatusBadRequest)
			} else {
        if userInfo, err := provider.UserInfo(ctx, tokenSource) ; err != nil {
          log.Printf("Error getting UserInfo: ", err)
          http.Error(w, "Internal error", http.StatusInternalServerError)
        } else {
          var claims struct {
            jwt.StandardClaims
            Email string `json:"email"`
            EmailVerified bool `json:"email_verified"`
            Name string `json:"name"`
            ClientRoles []string `json:"client_roles"`
          }

          if err := userInfo.Claims(&claims); err != nil {
            log.Printf("Error parsing claims from UserInfo: ", err)
            http.Error(w, "Internal error", http.StatusInternalServerError)
          } else {
		        data, err := json.MarshalIndent(userInfo, "", "    ")

            if err != nil {
			        http.Error(w, err.Error(), http.StatusInternalServerError)
		        } else {
		          w.Write(data)
            }
          }
        }
      }
    }
  })

  log.Printf("Listening on http://%s/", listenAddress)
  log.Fatal(http.ListenAndServe(listenAddress, nil))
}
