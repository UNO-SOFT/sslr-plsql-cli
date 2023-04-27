// Copyright 2023 Tamás Gulácsi. All rights reserved.

package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"

	"github.com/tgulacsi/go/iohlp"
)

//go:embed sslr-plsql-toolkit-3.8.0.4948.jar
var sslrJAR []byte

//go:embed out/production/sslr/sslr/Main.class
var mainClass []byte

func main() {
	if err := Main(); err != nil {
		log.Fatal(err)
	}
}

func Main() error {
	flagServer := flag.String("server", "http://localhost:8003", "SSLR server")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	stdin := io.Reader(os.Stdin)
	if flag.NArg() > 0 {
		if fn := flag.Arg(0); fn != "" && fn != "-" {
			fh, err := os.Open(fn)
			if err != nil {
				return err
			}
			defer fh.Close()
			stdin = fh
		}
	}
	done := make(chan error, 1)
	if *flagServer != "" {
		sr, err := iohlp.MakeSectionReader(stdin, 1<<20)
		if err != nil {
			return err
		}
		stdin = io.NewSectionReader(sr, 0, sr.Size())
		go func() {
			defer close(done)
			select {
			case <-ctx.Done():
				return
			case done <- func() error {
				resp, err := http.Post(*flagServer, "text/plain", io.NewSectionReader(sr, 0, sr.Size()))
				if err != nil {
					return fmt.Errorf("POST to %s: %w", *flagServer, err)
				}
				if resp.StatusCode > 399 {
					return fmt.Errorf("POST to %s: %s", *flagServer, resp.Status)
				}
				io.Copy(os.Stdout, resp.Body)
				return nil
			}():
			}
		}()
	}

	dn, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dn)
	if err = os.MkdirAll(filepath.Join(dn, "sslr"), 0750); err != nil {
		return err
	}
	jarFn := filepath.Join(dn, "sslr.jar")
	if err = os.WriteFile(jarFn, sslrJAR, 0640); err != nil {
		return err
	}
	classFn := filepath.Join(dn, "sslr", "Main.class")
	if err = os.WriteFile(classFn, mainClass, 0640); err != nil {
		return err
	}
	{
		cmd := exec.CommandContext(ctx, "find", dn, "-ls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	cmd := exec.CommandContext(ctx, "java", "-cp", dn+":"+jarFn, "sslr.Main")
	cmd.Stdin = stdin
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err = <-done:
		if err == nil {
			return nil
		}
		log.Printf("http error: %+v", err)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%q: %w", cmd.Args, err)
	}
	return nil
}
