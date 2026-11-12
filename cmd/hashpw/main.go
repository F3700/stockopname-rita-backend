// Command hashpw prints the AUTH_PASSWORD_HASH value for a password.
//
//	go run ./cmd/hashpw            # reads the password from stdin
//	go run ./cmd/hashpw 'secret'   # convenient, but lands in the shell history
//
// The output is safe to paste into .env: the salt is random and the hash is a
// one-way PBKDF2-SHA256 digest.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"stockopname-rita-backend/internal/auth"
)

func main() {
	password, err := readPassword(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	fmt.Println(hash)
}

func readPassword(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	fmt.Fprint(os.Stderr, "Password: ")

	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("reading password: %w", err)
	}

	return strings.TrimRight(line, "\r\n"), nil
}
