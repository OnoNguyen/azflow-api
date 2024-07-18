package auth

import (
	"crypto/rsa"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var testPrivateKey *rsa.PrivateKey

func init() {
	// Initialize environment variables for testing
	os.Setenv("AZURE_ENTRA_CLIENT_ID", "8e269668-5996-4bef-87e1-a803cec7ed67")
	os.Setenv("AZURE_ENTRA_CLIENT_SECRET", "pJ78Q~BN6YSpqMOJWE6KmP4.u3X1gFD_wrx-nbRc")
	os.Setenv("AZURE_ENTRA_AUTHORITY", "https://azflowext.ciamlogin.com/474d5211-4d1a-41e8-904c-25b3c4bb5677")
	os.Setenv("AZURE_ENTRA_TENANT_ID", "474d5211-4d1a-41e8-904c-25b3c4bb5677")

	// Call the Init function to set up JWKS and other variables
	Init()
}

func TestMiddleware(t *testing.T) {
	// Generate a test token
	//token, err := acquireToken()
	//if err != nil {
	//	t.Fatal(err)
	//}

	// paste the token from web
	token := "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6Ik1HTHFqOThWTkxvWGFGZnBKQ0JwZ0I0SmFLcyJ9.eyJhdWQiOiI4ZTI2OTY2OC01OTk2LTRiZWYtODdlMS1hODAzY2VjN2VkNjciLCJpc3MiOiJodHRwczovLzQ3NGQ1MjExLTRkMWEtNDFlOC05MDRjLTI1YjNjNGJiNTY3Ny5jaWFtbG9naW4uY29tLzQ3NGQ1MjExLTRkMWEtNDFlOC05MDRjLTI1YjNjNGJiNTY3Ny92Mi4wIiwiaWF0IjoxNzIxMTE1ODg1LCJuYmYiOjE3MjExMTU4ODUsImV4cCI6MTcyMTExOTc4NSwiYWlvIjoiQVRRQXkvOFhBQUFBN0t0dlN4Qm4wZHBnWmpjTm5YaWp1ZFVqdkl1VTF3cVZyQ21Mako4QlBxdGVHMHJVTHVUMlVjRktjTVBjR0lkUiIsIm5hbWUiOiJ1bmtub3duIiwibm9uY2UiOiIwMTkwYmE4NC1jZGRkLTc0ZGEtOGE0Zi04Y2E5M2RhNDQ2OTgiLCJvaWQiOiIxYTBiM2I2Zi01MmM2LTQwMzktYWZjYS1kN2Y5M2Y5ZmY5NjMiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJuZ3V5ZW5vbm9AZ21haWwuY29tIiwicmgiOiIwLkFiZ0FFVkpOUnhwTjZFR1FUQ1d6eEx0V2QyaVdKbzZXV2U5TGgtR29BODdIN1dlNEFKRS4iLCJzdWIiOiJqYXY3WnhEcDl4bjV2TTJtQkRmQ21XcmtSTFUyZ3ZxWld1Q25KOGFYYjVNIiwidGlkIjoiNDc0ZDUyMTEtNGQxYS00MWU4LTkwNGMtMjViM2M0YmI1Njc3IiwidXRpIjoiYVk2VS1reHd6VVNaaloxamFfQUFBQSIsInZlciI6IjIuMCJ9.wm3zspTR5ePYCA6H8CEVXPlyNmhhu57oIDI7IT81goxOcSf1yaacrrX6fLXEcQa_BDWqaO5gvO4AwSTes-Nm_JRa5t71ft24SEz1AUKi13LrhwgrQj7pDo93bOGyFk0i8djCvPaIDMeA3z8iao8OL1KVNPdpFSTWk86BgKItegngbxMtVOvb9xoyW6gfzJw3grQyrMofjMp3lt52AaBDxIcbBdrcaclV6nebV4pPUIWblopV55WgTaKcK4Hejf8uk0Vgaf6ouIgfBbDNSmZC-BV99Wr4KC27Qa_dnFBg15Wk6QeFd_-odcrBKqVgGyxvuh667muEsZGzG8mxQb-Fzg"

	// Define a handler that will use the middleware
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		member := ForContext(r.Context())
		if member != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Authenticated account: " + member.Email))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Unauthenticated request"))
		}
	})

	// Wrap the test handler with the middleware
	handlerToTest := Middleware()(testHandler)

	// Create a new HTTP request with the valid token
	req := httptest.NewRequest("GET", "/api", nil)
	req.Header.Set("Authorization", token)

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	handlerToTest.ServeHTTP(rr, req)

	// Check the response status code and body
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Authenticated account:")

	// Test with no Authorization header (unauthenticated)
	req = httptest.NewRequest("GET", "/api", nil)
	rr = httptest.NewRecorder()

	handlerToTest.ServeHTTP(rr, req)

	// Check the response status code and body
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "Unauthenticated request", rr.Body.String())
}
