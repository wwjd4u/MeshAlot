package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wwjd4u/MeshAlot/agent"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "enroll":
		err = runEnroll(os.Args[2:])
	case "identity", "status":
		err = runIdentity(os.Args[2:])
	case "version":
		fmt.Println(agent.SecureEnrollmentVersion)
		return
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "meshalot-agent:", err)
		os.Exit(1)
	}
}

func runEnroll(args []string) error {
	flags := flag.NewFlagSet("enroll", flag.ContinueOnError)
	serverURL := flags.String("server", env("MESHALOT_SERVER_URL", "https://api.meshalot.com"), "MeshAlot control API URL")
	identityPath := flags.String("identity", "", "identity file path")
	codeFile := flags.String("code-file", "", "read the one-time enrollment code from a file")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("enrollment codes are not accepted as command-line arguments; use the prompt, MESHALOT_ENROLLMENT_CODE, or --code-file")
	}
	code, err := readEnrollmentCode(*codeFile)
	if err != nil {
		return err
	}
	path, err := resolveIdentityPath(*identityPath)
	if err != nil {
		return err
	}
	identity, created, err := agent.LoadOrCreateIdentity(path)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err = agent.EnrollSecure(ctx, *serverURL, identity, code); err != nil {
		return err
	}
	keyFingerprint, err := agent.PublicKeyFingerprint(identity)
	if err != nil {
		return err
	}
	fmt.Println("Enrollment accepted.")
	fmt.Printf("Node identity fingerprint: %s\n", agent.NodeIDFingerprint(identity.NodeID))
	fmt.Printf("Public key fingerprint: %s\n", keyFingerprint)
	fmt.Printf("Identity file: %s\n", path)
	if created {
		fmt.Println("A new persistent local identity was created with restrictive permissions.")
	} else {
		fmt.Println("The existing persistent local identity was reused.")
	}
	return nil
}

func runIdentity(args []string) error {
	flags := flag.NewFlagSet("identity", flag.ContinueOnError)
	identityPath := flags.String("identity", "", "identity file path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	path, err := resolveIdentityPath(*identityPath)
	if err != nil {
		return err
	}
	identity, err := agent.LoadIdentity(path)
	if err != nil {
		return err
	}
	keyFingerprint, err := agent.PublicKeyFingerprint(identity)
	if err != nil {
		return err
	}
	fmt.Printf("Node identity fingerprint: %s\n", agent.NodeIDFingerprint(identity.NodeID))
	fmt.Printf("Public key fingerprint: %s\n", keyFingerprint)
	fmt.Printf("Identity file: %s\n", path)
	return nil
}

func readEnrollmentCode(codeFile string) (string, error) {
	if strings.TrimSpace(codeFile) != "" {
		data, err := os.ReadFile(codeFile)
		if err != nil {
			return "", fmt.Errorf("read enrollment code file: %w", err)
		}
		code := strings.TrimSpace(string(data))
		if code == "" {
			return "", errors.New("enrollment code file is empty")
		}
		return code, nil
	}
	if code := strings.TrimSpace(os.Getenv("MESHALOT_ENROLLMENT_CODE")); code != "" {
		return code, nil
	}
	fmt.Fprint(os.Stderr, "Enrollment code: ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", fmt.Errorf("read enrollment code: %w", err)
	}
	code := strings.TrimSpace(line)
	if code == "" {
		return "", errors.New("enrollment code is required")
	}
	return code, nil
}

func resolveIdentityPath(value string) (string, error) {
	if strings.TrimSpace(value) != "" {
		return value, nil
	}
	return agent.DefaultIdentityPath()
}

func env(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: meshalot-agent <enroll|identity|status|version> [options]")
}
