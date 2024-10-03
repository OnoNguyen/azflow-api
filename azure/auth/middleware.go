package auth

import (
	"azflow-api/domain/account"
	"context"
	"crypto/rsa"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/jwk"
	"net/http"
	"os"
)

type contextKey string

const userCtxKey contextKey = "account"

var (
	ClientID     string
	ClientSecret string
	Authority    string
)

func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				// No auth header, pass the request along as unauthenticated
				next.ServeHTTP(w, r)
				return
			}

			// Check if it's a local dev request
			if authHeader == "azflow@local.dev" {
				env := os.Getenv("ENV")
				if env == "local" {
					user := &account.Member{
						Email: authHeader,
						ExtId: "12345",
					}

					// Pass the request to the next handler with account info in context
					ctx := context.WithValue(r.Context(), userCtxKey, user)
					next.ServeHTTP(w, r.WithContext(ctx))

					return
				}
			}

			tokenString := authHeader[len("Bearer "):]

			verified, err := verifyToken(tokenString, r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			// Extract account information from the token claims
			if claims, ok := verified.Claims.(jwt.MapClaims); ok && verified.Valid {
				user := &account.Member{
					Email: claims["preferred_username"].(string),
					ExtId: claims["oid"].(string),
				}

				// Pass the request to the next handler with account info in context
				ctx := context.WithValue(r.Context(), userCtxKey, user)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			}
		})
	}
}

func verifyToken(tokenString string, context context.Context) (*jwt.Token, error) {

	keySet, err := jwk.Fetch(context, "https://login.microsoftonline.com/common/discovery/v2.0/keys")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid header not found")
		}

		keys, ok := keySet.LookupKeyID(kid)
		if !ok {
			return nil, fmt.Errorf("key %v not found", kid)
		}

		publickey := &rsa.PublicKey{}
		err = keys.Raw(publickey)
		if err != nil {
			return nil, fmt.Errorf("could not parse pubkey")
		}

		return publickey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		aud := claims["aud"].(string)
		if aud != ClientID {
			return nil, fmt.Errorf("invalid audience: %v", aud)
		}
	} else {
		return nil, fmt.Errorf("invalid token claims")
	}

	return token, nil
}

func Init() {
	ClientID = os.Getenv("AZURE_ENTRA_CLIENT_ID")
	ClientSecret = os.Getenv("AZURE_ENTRA_CLIENT_SECRET")
	Authority = os.Getenv("AZURE_ENTRA_AUTHORITY")
}

func ForContext(ctx context.Context) *account.Member {
	u, ok := ctx.Value(userCtxKey).(*account.Member)
	if !ok {
		fmt.Println("no account in context")
		return nil
	}
	return u
}

func GetMember(ctx context.Context) (*account.Member, error) {
	u, ok := ctx.Value(userCtxKey).(*account.Member)

	if !ok {
		return nil, fmt.Errorf("unauthenticated")
	}

	println("Member id", u.Email)

	return u, nil
}
