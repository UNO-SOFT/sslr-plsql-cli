// Copyright 2023 Tamás Gulácsi. All rights reserved.

package main

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
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

	"github.com/UNO-SOFT/zlog/v2"
	//"golang.org/x/exp/slog"

	"github.com/tgulacsi/go/iohlp"
	"golang.org/x/sync/errgroup"
	"golang.org/x/text/encoding/charmap"

	"github.com/godror/godror"
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

	verbose zlog.VerboseVar
	logger  = zlog.NewLogger(zlog.MaybeConsoleHandler(&verbose, os.Stderr)).SLog()
)

func main() {
	if err := Main(); err != nil {
		log.Fatal(err)
	}
}

func Main() error {
	flagServer := flag.String("server", "http://localhost:8003", "SSLR server")
	flagXML := flag.Bool("xml", false, "output raw XML")
	flagFormat := flag.String("format", "{{.FullName}}:{{.Begin}}:{{.End}}\t{{range .Calls}}{{.Other}},{{end}}\n", "format to print")
	flagLine := flag.Int("line", 0, "line number to get the function name")
	flagConnect := flag.String("connect", os.Getenv("BRUNO_OWNER_ID"), "connect to this database to get function/procedure names")
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

	var grp errgroup.Group

	var name string
	var funcs []Function
	grp.Go(func() error {
		var err error
		name, funcs, err = GetFunctions(out)
		return err
	})
	var procedures map[string][2]string
	if *flagConnect != "" {
		grp.Go(func() error {
			db, err := sql.Open("godror", *flagConnect)
			if err != nil {
				logger.Warn("connect to database", "dsn", *flagConnect, "error", err)
				return nil
			}
			defer db.Close()
			const qry = "SELECT NVL2(procedure_name, object_name, NULL) AS package_name, NVL(procedure_name, object_name) AS procedure_name FROM user_procedures"
			rows, err := db.QueryContext(ctx, qry, godror.FetchArraySize(4<<10), godror.PrefetchCount(4<<10+1))
			if err != nil {
				return fmt.Errorf("%s: %w", qry, err)
			}
			defer rows.Close()
			procedures = make(map[string][2]string)
			for rows.Next() {
				var pkg, proc string
				if err := rows.Scan(&pkg, &proc); err != nil {
					return fmt.Errorf("scan %q: %w", qry, err)
				}
				k := proc
				if pkg != "" {
					k = pkg + "." + proc
				}
				procedures[k] = [2]string{pkg, proc}
			}
			return rows.Err()
		})
	}
	if err := grp.Wait(); err != nil {
		return err
	}
	for _, f := range funcs {
		procedures[name+"."+f.FullName()] = [2]string{name, f.FullName()}
	}

	for i := range funcs {
		f := &funcs[i]
		for j := 0; j < len(f.Calls); j++ {
			c := &f.Calls[j]
			if c.Procedure && !strings.HasSuffix(c.Other, ".DELETE") {
				continue
			}
			k := c.Other
			if strings.IndexByte(k, '.') < 0 {
				if _, ok := procedures[k]; ok {
					continue
				}
				k = name + "." + k
			}
			if _, ok := procedures[k]; ok {
				c.Other = k
			} else {
				f.Calls[j] = f.Calls[len(f.Calls)-1]
				f.Calls = f.Calls[:len(f.Calls)-1]
				j--
			}
		}
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
