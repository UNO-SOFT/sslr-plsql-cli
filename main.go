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
	"unicode/utf8"

	"github.com/tgulacsi/go/iohlp"
	"golang.org/x/text/encoding/charmap"
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
	sr, err := iohlp.MakeSectionReader(stdin, 1<<20)
	if err != nil {
		return err
	}
	stdin = io.NewSectionReader(sr, 0, sr.Size())
	{
		var invalid bool
		r := io.NewSectionReader(sr, 0, sr.Size())
		var a [4096]byte
		for {
			n, err := r.Read(a[:])
			if n == 0 {
				if err == io.EOF {
					break
				}
				return err
			}
			b := a[:n]
			for len(b) != 0 {
				r, size := utf8.DecodeRune(b)
				if size == 0 || r == utf8.RuneError {
					invalid = true
					break
				}
				b = b[size:]
			}
		}

		if invalid {
			log.Println("input is not valid UTF-8, try to convert from ISO8859-2")
			if stdin, err = iohlp.MakeSectionReader(charmap.ISO8859_2.NewDecoder().Reader(io.NewSectionReader(sr, 0, sr.Size())), 1<<20); err != nil {
				return err
			}
		}
	}

	type result struct {
		Body []byte
		Err  error
	}
	var done chan result
	if *flagServer != "" {
		done = make(chan result, 1)
		go func() {
			defer close(done)
			done <- func() result {
				req, err := http.NewRequestWithContext(ctx, "POST", *flagServer, io.NewSectionReader(sr, 0, sr.Size()))
				if err != nil {
					return result{Err: err}
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return result{Err: fmt.Errorf("POST to %s: %w", *flagServer, err)}
				}
				if resp.StatusCode > 399 {
					return result{Err: fmt.Errorf("POST to %s: %s", *flagServer, resp.Status)}
				}
				var res result
				res.Body, res.Err = io.ReadAll(resp.Body)
				return res
			}()
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
	if false {
		cmd := exec.CommandContext(ctx, "find", dn, "-ls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	cmd := exec.CommandContext(ctx, "java", "-cp", dn+":"+jarFn, "sslr.Main")
	cmd.Stdin = stdin
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if done != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res, ok := <-done:
			if ok && res.Err == nil {
				log.Println("result from HTTP server")
				_, err = cmd.Stdout.Write(res.Body)
				return err
			}
			log.Printf("http error: %+v", res.Err)
		}
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%q: %w", cmd.Args, err)
	}
	return nil
}
