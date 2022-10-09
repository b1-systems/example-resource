/* Demonstration of an OIDC "Resource Server", receiving an OAuth2 access token
   and (for demonstration purpose) accessing the OIDC userinfo endpoint
   * See also: https://openid.net/specs/openid-connect-core-1_0.html#UserInfo */

package main

import (
  "encoding/json"
  "log"
  "net/http"
  "os"
  "strings"
  "github.com/coreos/go-oidc/v3/oidc"
  "golang.org/x/net/context"
  "golang.org/x/oauth2"
  "example-resource/ini"
)

var (
  clientName = "example-resource"
  providerUrl = ""
  listenAddress = ""
)

func main() {
  arr := []ini.Ref{
    {"providerUrl", &providerUrl},
    {"listenAddress", &listenAddress}}

  err := ini.ReadIni(clientName, arr)

  if err != nil {
    log.Fatal()
    os.Exit(1)
  }


  ctx := context.Background()
  provider, err := oidc.NewProvider(ctx, providerUrl)

  if err != nil {
    log.Fatal(err)
    os.Exit(1)
  }

  config := oauth2.Config{
    Endpoint: provider.Endpoint(),
    Scopes: []string{oidc.ScopeOpenID},
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
        log.Printf("Invalid access token: ", err)
        http.Error(w, "Bad request", http.StatusBadRequest)
      } else {
        if userInfo, err := provider.UserInfo(ctx, tokenSource) ; err != nil {
          log.Printf("Error getting UserInfo: ", err)
          http.Error(w, "Internal error", http.StatusInternalServerError)
        } else {
          var claims struct {
            Sub string `json:"sub"`
            Email string `json:"email"`
            EmailVerified bool `json:"email_verified"`
            PreferredUsername string `json:"preferred_username"`
            Name string `json:"name"`
            GivenName string `json:"given_name"`
            FamilyName string `json:"family_name"`
            ResourceAccess struct {
              ExampleFrontend struct {
                Roles []string `json:"roles"`
              } `json:"example-frontend"`
            } `json:"resource_access"`
          }

          if err := userInfo.Claims(&claims); err != nil {
            log.Printf("Error parsing claims from UserInfo: ", err)
            http.Error(w, "Internal error", http.StatusInternalServerError)
          } else {
            data, err := json.MarshalIndent(claims, "", "    ")

            if err != nil {
              http.Error(w, err.Error(), http.StatusInternalServerError)
            } else {
              w.Write([]byte("--------"+clientName+"--------\r\n"))
              w.Write([]byte("Parsed userinfo claims: "))
              w.Write(data)
              w.Write([]byte("\r\n"))
            }
          }
        }
      }
    }
  })

  log.Printf("Listening on http://%s/", listenAddress)
  log.Fatal(http.ListenAndServe(listenAddress, nil))
}
