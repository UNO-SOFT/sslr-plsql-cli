// Copyright 2023 Tamás Gulácsi. All rights reserved.

package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/tgulacsi/go/iohlp"
	"golang.org/x/text/encoding/charmap"
)

var (
	//go:embed sslr-plsql-toolkit-3.8.0.4948.jar
	sslrJAR []byte
	//go:embed commons-text-1.10.0.jar
	commonsTextJAR []byte
	//go:embed commons-lang3-3.12.0.jar
	commonsLangJAR []byte

	//go:embed out/production/sslr/sslr/Main.class
	mainClass []byte
)

func main() {
	if err := Main(); err != nil {
		log.Fatal(err)
	}
}

func Main() error {
	flagServer := flag.String("server", "http://localhost:8003", "SSLR server")
	flagXML := flag.Bool("xml", false, "output raw XML")
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
	var cp strings.Builder
	cp.WriteString(filepath.Join(dn))
	for _, jar := range []struct {
		Data []byte
		Name string
	}{
		{Data: sslrJAR, Name: "sslr.jar"},
		{Data: commonsLangJAR, Name: "commons-lang3.jar"},
		{Data: commonsTextJAR, Name: "commons-text.jar"},
	} {
		jarFn := filepath.Join(dn, jar.Name)
		if err = os.WriteFile(jarFn, jar.Data, 0640); err != nil {
			return err
		}
		if cp.Len() != 0 {
			cp.WriteByte(':')
		}
		cp.WriteString(jarFn)
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
	var buf bytes.Buffer
	cmd := exec.CommandContext(ctx, "java", "-cp", cp.String(), "sslr.Main")
	cmd.Stdin = stdin
	cmd.Stdout, cmd.Stderr = &buf, os.Stderr
	if done != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res, ok := <-done:
			if ok && res.Err == nil {
				log.Println("result from HTTP server")
				buf.Write(res.Body)
			} else {
				log.Printf("http error: %+v", res.Err)
			}
		}
	}
	if buf.Len() == 0 {
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%q: %w", cmd.Args, err)
		}
	}

	if *flagXML {
		_, err = os.Stdout.Write(buf.Bytes())
		return err
	}
	dec := xml.NewDecoder(bytes.NewReader(buf.Bytes()))
	dec.Strict = false
	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		type element struct {
			Name, Value  string
			Line, Column int
		}
		if st, ok := tok.(xml.StartElement); ok {
			elt := element{Name: st.Name.Local}
			for _, a := range st.Attr {
				switch a.Name.Local {
				case "tokenValue":
					elt.Value = a.Value
				case "tokenLine":
					elt.Line, _ = strconv.Atoi(a.Value)
				case "tokenColumn":
					elt.Column, _ = strconv.Atoi(a.Value)
				}
			}
			fmt.Println(elt)
		}
	}
	return nil
}
