package common

import (
	"testing"
)

func TestEncrypt(t *testing.T) {
	secret := "Some secret text"
	password := "some password"

	output, err := Encrypt(secret, password)
	if err != nil {
		t.Fatal("Got err:", err)
	}

	if output == secret {
		t.Fatal("output was not encrypted")
	}

	decrypted, err := Decrypt(output, password)
	if err != nil {
		t.Fatal("Got err:", err)
	}

	if decrypted != secret {
		t.Fatalf("Expected %q, got %q", secret, decrypted)
	}

	// Wrong password
	_, err = Decrypt(output, "wrong password")
	if err == nil {
		t.Fatal("Expected decryption to fail")
	}
}
