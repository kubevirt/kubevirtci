package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	expect "github.com/google/goexpect"
	"google.golang.org/grpc/codes"
)

const (
	PromptExpression = `(\$ |\# )`
	CRLF             = "\r\n"
	UTFPosEscape     = "\u001b\\[[0-9]+;[0-9]+H"
)

var (
	ShellSuccess       = retValue("0")
	ShellFail          = retValue("[1-9].*")
	ShellSuccessRegexp = regexp.MustCompile(ShellSuccess)
	ShellFailRegexp    = regexp.MustCompile(ShellFail)
)

func retValue(retcode string) string {
	return "\n" + retcode + CRLF + ".*" + PromptExpression
}

func command(command string) []expect.Batcher {
	return []expect.Batcher{
		&expect.BSnd{S: "\n"},
		&expect.BExp{R: PromptExpression},
		&expect.BSnd{S: command + "\n"},
		&expect.BExp{R: PromptExpression},
		&expect.BSnd{S: "echo $?\n"},
		&expect.BCas{C: []expect.Caser{
			&expect.Case{
				R: ShellSuccessRegexp,
				T: expect.OK(),
			},
			&expect.Case{
				R: ShellFailRegexp,
				T: expect.Fail(expect.NewStatus(codes.Unavailable, command+" failed")),
			},
		}},
	}
}

func main() {

	opts := []expect.Option{}
	opts = append(opts, expect.Verbose(true))

	expecter, _, err := expect.Spawn(fmt.Sprintf("virsh console %s", os.Getenv("DOMAIN")), 2*time.Second, opts...)
	if err != nil {
		log.Fatalf("failed running virsh: %v", err)
	}
	defer expecter.Close()

	err = expecter.Send("\n")
	if err != nil {
		log.Fatalf("failed sending new line: %v", err)
	}

	// Do not login, if we already logged in
	b := append([]expect.Batcher{
		&expect.BSnd{S: "\n"},
		&expect.BExp{R: "localhost:~\\# "},
	})
	_, err = expecter.ExpectBatch(b, 5*time.Second)
	if err == nil {
		return
	}

	b = append([]expect.Batcher{
		&expect.BSnd{S: "\n"},
		&expect.BExp{R: "localhost login:"},
		&expect.BSnd{S: "root\n"},
		&expect.BExp{R: `(\$ |\# )`}})
	_, err = expecter.ExpectBatch(b, 180*time.Second)
	if err != nil {
		log.Fatalf("failed sending login: %v", err)
	}

	_, err = expecter.ExpectBatch(command("mkdir /tmp/setup"), 180*time.Second)
	if err != nil {
		log.Fatalf("failed running setup script: %v", err)
	}

	_, err = expecter.ExpectBatch(command("mount /dev/sr1 /tmp/setup"), 180*time.Second)
	if err != nil {
		log.Fatalf("failed mounting setup script: %v", err)
	}

	_, err = expecter.ExpectBatch(command("/tmp/setup/setup.sh"), 180*time.Second)
	if err != nil {
		log.Fatalf("failed running setup script: %v", err)
	}
}
