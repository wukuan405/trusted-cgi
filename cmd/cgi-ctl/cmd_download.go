package main

import (
	"fmt"
	"github.com/alecthomas/units"
	"github.com/reddec/trusted-cgi/cmd/internal"
	"io"
	"log"
	"os"
)

type download struct {
	remoteLink
	uidLocator
	Output string `short:"o" long:"output" env:"OUTPUT" description:"Output data (- means stdout, empty means as UID)" default:""`
}

func (cmd *download) Execute(args []string) error {
	ctx, closer := internal.SignalContext()
	defer closer()
	if err := cmd.parseUID(); err != nil {
		return err
	}
	log.Println("login...")
	token, err := cmd.Token(ctx)
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}
	log.Println("download...")
	tarball, err := cmd.Lambdas().Download(ctx, token, cmd.UID)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	log.Println("downloaded", units.Base2Bytes(len(tarball)))
	var out io.Writer
	if cmd.Output == "" {
		cmd.Output = cmd.UID + ".tar.gz"
	}
	if cmd.Output == "-" {
		out = os.Stdout
	} else {
		log.Println("saving to", cmd.Output, "...")
		f, err := os.Create(cmd.Output)
		if err != nil {
			return fmt.Errorf("create destination file: %w", err)
		}
		defer f.Close()
		out = f
	}
	var w int
	for w < len(tarball) {
		n, err := out.Write(tarball[w:])
		if err != nil {
			return err
		}
		if n <= 0 {
			break
		}
		w += n
	}
	log.Println("done")
	return nil
}
