package middleware

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/coffee4ai/school-api-b3/config"
	"github.com/cristalhq/jwt/v3"
	"golang.org/x/crypto/bcrypt"
)

type UserClaims struct {
	jwt.RegisteredClaims
	User_ID string `json:"user_id"`
}

func GenerateHash(password string) ([]byte, error) {
	hashPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error generating bcrypt hash from password %w", err)
	}
	return hashPass, nil
}

func ComparePassword(password string, hashPass []byte) bool {
	fmt.Println("ComparePassword", password, hashPass)
	err := bcrypt.CompareHashAndPassword(hashPass, []byte(password))
	fmt.Println(err)
	return err == nil
}

func VerifyTokenRole(tk, role string) (bool, error) {
	var err error
	splitToken := strings.Split(tk, "Bearer ")
	if len(splitToken) != 2 {
		return false, fmt.Errorf("invalid bearer token format")
	}
	tokenStr := strings.TrimSpace(splitToken[1])

	// create a Verifier (HMAC in this example)
	sKey := config.GetApiSecret()
	verifier, err := jwt.NewVerifierHS(jwt.HS256, []byte(sKey))
	if err != nil {
		return false, fmt.Errorf("internal server error")
	}

	// parse a Token
	token, err := jwt.ParseString(tokenStr)
	if err != nil {
		return false, fmt.Errorf("invalid token")
	}

	// and verify it's signature
	err = verifier.Verify(token.Payload(), token.Signature())
	if err != nil {
		return false, fmt.Errorf("invalid token")
	}

	var newClaims UserClaims
	errClaims := json.Unmarshal(token.RawClaims(), &newClaims)
	if errClaims != nil {
		return false, fmt.Errorf("internal server error")
	}

	if !newClaims.IsValidAt(time.Now()) {
		return false, fmt.Errorf("expired token")
	}

	auth := newClaims.IsForAudience(role)
	if !auth {
		err = fmt.Errorf("not authorised")
	}
	return auth, err
}
