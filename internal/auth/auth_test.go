package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWT(t *testing.T) {
	id := uuid.New()
	token, err := MakeJWT(id, "secret", time.Second)
	if err != nil {
		t.Errorf("MakeJWT err: %s", err)
	}

	tokenID, err := ValidateJWT(token, "secret")
	if err != nil {
		t.Errorf("ValidateJWT err: %s", err)
	}

	if id != tokenID {
		t.Error("tokenID and input ID don't match")
	}
}

func TestInvalidJWT(t *testing.T) {
	token, err := MakeJWT(uuid.New(), "secret", time.Second)
	if err != nil {
		t.Errorf("MakeJWT err: %s", err)
	}
	time.Sleep(time.Second)

	_, err = ValidateJWT(token, "secret")
	if err == nil {
		t.Errorf("token not invalidated")
	}
}

func TestIncorrectSecret(t *testing.T) {
	token, err := MakeJWT(uuid.New(), "secret", time.Second)
	if err != nil {
		t.Errorf("MakeJWT err: %s", err)
	}
	time.Sleep(time.Second)

	_, err = ValidateJWT(token, "notSecret")
	if err == nil {
		t.Errorf("wrong secret validated")
	}
}
