// Copyright 2023 Tamás Gulácsi. All rights reserved.

package main

import (
	"bufio"
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
	"text/template"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"google.golang.org/protobuf/proto"

	"github.com/UNO-SOFT/sslr-plsql-cli/pb"
	"github.com/tgulacsi/go/iohlp"

	"github.com/UNO-SOFT/zlog/v2"
)

//go:generate go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
//go:generate protoc --proto_path=pb --go_out=pb --go_opt=paths=source_relative pb/functions.proto

var (
	//go:embed sslr-plsql-toolkit-3.8.0.4948.jar
	sslrJAR []byte
	//go:embed commons-text-1.10.0.jar
	commonsTextJAR []byte
	//go:embed commons-lang3-3.12.0.jar
	commonsLangJAR []byte

	//go:embed out/production/sslr/sslr/Main.class
	mainClass []byte

	verbose zlog.VerboseVar
	logger  = zlog.NewLogger(zlog.MaybeConsoleHandler(&verbose, os.Stderr)).SLog()
)

func main() {
	if err := Main(); err != nil {
		log.Fatal(err)
	}
}

func Main() error {
	flagServer := flag.String("server", "http://localhost:7003", "SSLR server")
	flagXML := flag.Bool("xml", false, "output raw XML")
	flagPB := flag.Bool("pb", false, "output protobuf")
	flagFormat := flag.String("format", "{{.FullName}}:{{.Begin}}:{{.End}}\t{{range .Calls}}{{.Other}},{{end}}\n", "format to print")
	flagLine := flag.Int("line", 0, "line number to get the function name")
	flag.Var(&verbose, "v", "verbose logging")
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
	if stdin, err = toUTF8(io.NewSectionReader(sr, 0, sr.Size())); err != nil {
		return err
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
				logger.Info("result from HTTP server")
				// buf.Write(res.Body)
			} else {
				logger.Error("http", "error", res.Err)
			}
		}
	}
	if buf.Len() == 0 {
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%q: %w", cmd.Args, err)
		}
	}
	out, err := toUTF8(io.NewSectionReader(bytes.NewReader(buf.Bytes()), 0, int64(buf.Len())))
	if err != nil {
		return err
	}

	if *flagXML {
		_, err = io.Copy(os.Stdout, out)
		return err
	}

	name, funcs, err := GetFunctions(out)
	if err != nil {
		return err
	}
	if *flagPB {
		obj := pb.Object{Name: name, Functions: make([]*pb.Function, 0, len(funcs))}
		for _, f := range funcs {
			pbf := pb.Function{
				Name:  f.FullName(),
				Calls: make([]*pb.Call, 0, len(f.Calls)),
				Begin: uint32(f.Begin), End: uint32(f.End),
				Level: uint32(f.Level),
			}
			if f.Parent != nil {
				pbf.Parent = f.Parent.FullName()
			}
			for _, c := range f.Calls {
				pbf.Calls = append(pbf.Calls, &pb.Call{
					Other: c.Other, Line: uint32(c.Line),
					Procedure: c.Procedure,
				})
			}

			obj.Functions = append(obj.Functions, &pbf)
		}
		b, err := proto.Marshal(&obj)
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(b)
		os.Stdout.Close()
		return err
	}

	if line := *flagLine; line != 0 {
		for _, f := range funcs {
			if f.Begin <= line && line <= f.End {
				fmt.Println(f.FullName())
			}
		}
		return nil
	}

	defer os.Stdout.Close()
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	tmpl := template.Must(template.New("").Parse(*flagFormat))
	for _, f := range funcs {
		if err = tmpl.Execute(w, f); err != nil {
			return err
		}
	}

	return nil
}

type Call struct {
	Other     string
	Line      int
	Procedure bool
}

type Function struct {
	Name              string
	Calls             []Call
	Parent            *Function
	Begin, End, Level int
}

func (f Function) FullName() string {
	names := append(make([]string, 0, 2), f.Name)
	for p := f.Parent; p != nil; p = p.Parent {
		names = append(names, p.Name)
	}
	if len(names) == 1 {
		return names[0]
	}
	var buf strings.Builder
	for i := len(names) - 1; i >= 0; i-- {
		if buf.Len() != 0 {
			buf.WriteByte('.')
		}
		buf.WriteString(names[i])
	}
	return buf.String()
}

func GetFunctions(out io.Reader) (this string, funcs []Function, _ error) {
	m := make(map[string]*Function)
	dec := xml.NewDecoder(out)
	dec.Strict = false
	var tagPath, funPath, identifiers []string
	var level, line, lastLine, callLine int
	var act *Function
	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return this, funcs, err
		}

		if st, ok := tok.(xml.StartElement); ok {
			var tokenValue string
			for _, a := range st.Attr {
				if a.Name.Local == "tokenLine" {
					line, _ = strconv.Atoi(a.Value)
				} else if a.Name.Local == "tokenValue" {
					tokenValue = a.Value
				}
			}

			switch st.Name.Local {
			case "BIN_PACKAGE":
				this = tokenValue

			case "BIN_PROCEDURE", "BIN_FUNCTION":
				if this == "" {
					this = tokenValue
				}
				f := Function{Level: level}
				f.Name = tokenValue
				f.Begin = line
				f.Parent = m[strings.Join(funPath, ".")]
				i := len(tagPath) - 1
				if len(tagPath) > 3 &&
					(tagPath[i] == "PROCEDURE_HEADING" || tagPath[i] == "FUNCTION_HEADING") &&
					(tagPath[i-1] == "PROCEDURE_DEFINITION" || tagPath[i-1] == "FUNCTION_DEFINITION") {
					level++
					funPath = append(funPath, f.Name)
					m[strings.Join(funPath, ".")] = &f
				}
				if i := len(funcs) - 1; i >= 0 && funcs[i].End == 0 && funcs[i].Level >= f.Level {
					funcs[i].End = f.Begin - 1
				}
				funcs = append(funcs, f)
				act = &funcs[len(funcs)-1]

			case "PROCEDURE_CALL":
				callLine = line

			case "BIN_IDENTIFIER":
				if p := tagPath[len(tagPath)-1]; p == "PROCEDURE_CALL" || p == "EXPRESSION_PRIMARY" {
					identifiers = append(identifiers, tokenValue)
				}
			}

			lastLine = line
			tagPath = append(tagPath, st.Name.Local)

		} else if e, ok := tok.(xml.EndElement); ok {
			tagPath = tagPath[:len(tagPath)-1]

			switch e.Name.Local {
			case "PROCEDURE_DEFINITION", "FUNCTION_DEFINITION":
				funPath = funPath[:len(funPath)-1]
				level--

			case "PROCEDURE_CALL", "EXPRESSION_PRIMARY", "PAREN_L":
				if len(identifiers) != 0 {
					if act != nil {
						act.Calls = append(act.Calls, Call{
							Line:      callLine,
							Other:     strings.Join(identifiers, "."),
							Procedure: e.Name.Local == "PROCEDURE_CALL",
						})
					}
					identifiers = identifiers[:0]
				}

			}
		}
	}
	if i := len(funcs) - 1; i >= 0 && lastLine != 0 && funcs[i].End == 0 {
		funcs[i].End = lastLine
	}
	return this, funcs, nil
}

func toUTF8(sr *io.SectionReader) (*io.SectionReader, error) {
	var invalid bool
	r := io.NewSectionReader(sr, 0, sr.Size())
	var a [4096]byte
	for {
		n, err := r.Read(a[:])
		if n == 0 {
			if err == io.EOF {
				break
			}
			return sr, err
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

	if !invalid {
		return sr, nil
	}
	logger.Info("not valid UTF-8, try to convert from ISO8859-2")
	return iohlp.MakeSectionReader(charmap.ISO8859_2.NewDecoder().Reader(io.NewSectionReader(sr, 0, sr.Size())), 1<<20)
}
