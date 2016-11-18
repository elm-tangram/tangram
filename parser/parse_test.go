package parser

import (
	"strings"
	"testing"

	"github.com/mvader/elm-compiler/scanner"
	"github.com/stretchr/testify/require"
)

func TestParseModule(t *testing.T) {
	require := require.New(t)
	cases := []struct {
		input   string
		ok, eof bool
		module  string
		exposed [][]string
	}{
		{"module Foo", true, false, "Foo", nil},
		{"bar Foo", false, false, "", nil},
		{"module Foo.Bar", true, false, "Foo.Bar", nil},
		{"module Foo.Bar.Baz", true, false, "Foo.Bar.Baz", nil},
		{"module Foo exposing", false, true, "Foo", nil},
		{"module Foo exposing ()", false, false, "Foo", nil},
		{"module Foo exposing (..)", true, false, "Foo", [][]string{{".."}}},
		{"module Foo exposing (foo)", true, false, "Foo", [][]string{{"foo"}}},
		{"module Foo exposing (foo, bar)", true, false, "Foo", [][]string{{"foo"}, {"bar"}}},
		{"module Foo exposing (foo, bar, baz)", true, false, "Foo", [][]string{{"foo"}, {"bar"}, {"baz"}}},
		{"module Foo exposing (foo, (:>), baz)", true, false, "Foo", [][]string{{"foo"}, {":>"}, {"baz"}}},
		{"module Foo exposing ((:>), (:>), (:>))", true, false, "Foo", [][]string{{":>"}, {":>"}, {":>"}}},
		{"module Foo exposing (foo, Bar(..), Baz(A, B, C))", true, false, "Foo", [][]string{
			{"foo"},
			{"Bar", ".."},
			{"Baz", "A", "B", "C"},
		}},
	}

	for _, c := range cases {
		func() {
			defer func() {
				if r := recover(); r != nil {
					switch r.(type) {
					case bailout:
						if !c.eof {
							require.FailNow("unexpected bailout", c.input)
						}
					default:
						panic(r)
					}
				}
			}()

			p := stringParser(c.input)
			decl := p.parseModule()

			if c.ok {
				var exposed [][]string
				if decl.Exposing != nil {
					for _, e := range decl.Exposing.Idents {
						var exp = []string{e.Name}
						if e.Exposing != nil {
							for _, e := range e.Exposing.Idents {
								exp = append(exp, e.Name)
							}
						}
						exposed = append(exposed, exp)
					}
				}

				require.Equal(0, len(p.errors), c.input)
				require.Equal(c.module, decl.Name.String(), c.input)
				require.Equal(c.exposed, exposed, c.input)
			} else {
				require.NotEqual(0, len(p.errors), c.input)
			}
		}()
	}
}

func TestParseImport(t *testing.T) {
	require := require.New(t)
	cases := []struct {
		input   string
		ok, eof bool
		module  string
		alias   string
		exposed [][]string
	}{
		{"import Foo", true, false, "Foo", "", nil},
		{"bar Foo", false, false, "", "", nil},
		{"import Foo.Bar", true, false, "Foo.Bar", "", nil},
		{"import Foo.Bar.Baz", true, false, "Foo.Bar.Baz", "", nil},
		{"import Foo.Bar.Baz as Foo", true, false, "Foo.Bar.Baz", "Foo", nil},
		{"import Foo exposing", false, true, "Foo", "", nil},
		{"import Foo exposing ()", false, false, "Foo", "", nil},
		{"import Foo exposing (..)", true, false, "Foo", "", [][]string{{".."}}},
		{"import Foo as Bar exposing (..)", true, false, "Foo", "Bar", [][]string{{".."}}},
		{"import Foo exposing (foo)", true, false, "Foo", "", [][]string{{"foo"}}},
		{"import Foo exposing (foo, bar)", true, false, "Foo", "", [][]string{{"foo"}, {"bar"}}},
		{"import Foo exposing (foo, bar, baz)", true, false, "Foo", "", [][]string{{"foo"}, {"bar"}, {"baz"}}},
		{"import Foo exposing (foo, (:>), baz)", true, false, "Foo", "", [][]string{{"foo"}, {":>"}, {"baz"}}},
		{"import Foo exposing ((:>), (:>), (:>))", true, false, "Foo", "", [][]string{{":>"}, {":>"}, {":>"}}},
		{"import Foo exposing (foo, Bar(..), Baz(A, B, C))", true, false, "Foo", "", [][]string{
			{"foo"},
			{"Bar", ".."},
			{"Baz", "A", "B", "C"},
		}},
	}

	for _, c := range cases {
		func() {
			defer func() {
				if r := recover(); r != nil {
					switch r.(type) {
					case bailout:
						if !c.eof {
							require.FailNow("unexpected bailout", c.input)
						}
					default:
						panic(r)
					}
				}
			}()

			p := stringParser(c.input)
			decl := p.parseImport()

			if c.ok {
				var exposed [][]string
				if decl.Exposing != nil {
					for _, e := range decl.Exposing.Idents {
						var exp = []string{e.Name}
						if e.Exposing != nil {
							for _, e := range e.Exposing.Idents {
								exp = append(exp, e.Name)
							}
						}
						exposed = append(exposed, exp)
					}
				}

				require.Equal(0, len(p.errors), c.input)
				require.Equal(c.module, decl.Module.String(), c.input)
				require.Equal(c.exposed, exposed, c.input)
				if c.alias != "" {
					require.NotNil(decl.Alias, c.input)
					require.Equal(c.alias, decl.Alias.Name, c.input)
				}
			} else {
				require.NotEqual(0, len(p.errors), c.input)
			}
		}()
	}
}

func stringParser(str string) *parser {
	scanner := scanner.New("test", strings.NewReader(str))
	go scanner.Run()
	var p = new(parser)
	p.init("test", scanner)
	return p
}
