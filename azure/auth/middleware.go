package auth

import (
	"azflow-api/domain/user"
	"context"
	"crypto/rsa"
	"fmt"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/jwk"
	"net/http"
	"os"
)

type contextKey string

const userCtxKey contextKey = "user"

var (
	ClientID     string
	ClientSecret string
	Authority    string
	TenantID     string
	JwksURL      string
	Jwks         *keyfunc.JWKS
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

			tokenString := authHeader[len("Bearer "):]
			verified, err := verifyToken(tokenString, r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			// Extract user information from the token claims
			if claims, ok := verified.Claims.(jwt.MapClaims); ok && verified.Valid {
				user := &user.User{
					Username: claims["preferred_username"].(string),
					Email:    claims["preferred_username"].(string),
				}

				// Pass the request to the next handler with user info in context
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
		//return Jwks.Keyfunc(token)
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

func acquireToken() (string, error) {
	cred, err := confidential.NewCredFromSecret(ClientSecret)
	if err != nil {
		return "", err
	}

	clientApp, err := confidential.New(Authority, ClientID, cred)
	if err != nil {
		return "", err
	}

	result, err := clientApp.AcquireTokenByCredential(context.Background(), []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return "", err
	}

	return result.AccessToken, nil
}

func Init() {
	ClientID = os.Getenv("AZURE_ENTRA_CLIENT_ID")
	ClientSecret = os.Getenv("AZURE_ENTRA_CLIENT_SECRET")
	Authority = os.Getenv("AZURE_ENTRA_AUTHORITY")
	TenantID = os.Getenv("AZURE_ENTRA_TENANT_ID")
	JwksURL = fmt.Sprintf("%s/discovery/v2.0/keys", Authority)

	var err error
	//initializes the JWKS from Microsoft Entra
	Jwks, err = keyfunc.Get(JwksURL, keyfunc.Options{})
	if err != nil {
		panic(err)
	}
}

func ForContext(ctx context.Context) *user.User {
	u, ok := ctx.Value(userCtxKey).(*user.User)
	if !ok {
		fmt.Println("no user in context")
		return nil
	}
	return u
}

func GetUiserId(ctx context.Context) (string, error) {
	u, ok := ctx.Value(userCtxKey).(*user.User)

	if !ok {
		return "", fmt.Errorf("unauthenticated")
	}

	println("current user id", u.Email)

	return u.Email, nil
}
