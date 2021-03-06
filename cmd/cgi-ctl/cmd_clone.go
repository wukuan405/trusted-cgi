package main

import (
	"bytes"
	"fmt"
	"github.com/alecthomas/units"
	"github.com/reddec/trusted-cgi/cmd/internal"
	internal_app "github.com/reddec/trusted-cgi/internal"
	"log"
	"os"
	"os/exec"
)

type clone struct {
	remoteLink
	UID    string `short:"U" long:"uid" env:"UID" description:"Lambda UID" required:"yes"`
	Output string `short:"o" long:"output" env:"OUTPUT" description:"Output directory (empty - same as UID)" default:""`
}

func (cmd *clone) Execute(args []string) error {
	ctx, closer := internal.SignalContext()
	defer closer()
	log.Println("login...")
	token, err := cmd.Token(ctx)
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}

	if cmd.Output == "" {
		cmd.Output = cmd.UID
	}

	err = os.MkdirAll(cmd.Output, 0755)
	if err != nil {
		return fmt.Errorf("prepare output: %w", err)
	}

	err = os.Chdir(cmd.Output)
	if err != nil {
		return fmt.Errorf("change dir: %w", err)
	}

	log.Println("download...")
	tarball, err := cmd.Lambdas().Download(ctx, token, cmd.UID)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	log.Println("downloaded", units.Base2Bytes(len(tarball)))
	log.Println("extract to", cmd.Output, "...")
	untar := exec.CommandContext(ctx, "tar", "zxf", "-")
	untar.Stderr = os.Stderr
	untar.Stdout = os.Stdout
	untar.Stdin = bytes.NewReader(tarball)
	err = untar.Run()
	if err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	var cf controlFile
	cf.URL = cmd.URL
	cf.UID = cmd.UID
	err = cf.Save(controlFilename)
	if err != nil {
		return fmt.Errorf("save control file: %w", err)
	}

	err = appendIfNoLineFile(internal_app.CGIIgnore, controlFilename)
	if err != nil {
		return fmt.Errorf("update cgiignore file: %w", err)
	}

	log.Println("done")
	return nil
}
