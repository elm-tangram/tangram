package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/erizocosmico/elmo/ast"
	"github.com/erizocosmico/elmo/operator"

	"github.com/stretchr/testify/require"
)

const parseFixture = `
module Foo

import Foo.Bar
import Foo.Bar.Baz exposing (bar, baz)

foo : Int
foo = 45

(?) : comparable -> comparable -> Bool
(?) cmp1 cmp2 = cmp1 > cmp2

infixr 6 ?
infixl 7 :>

(:>) : comparable -> comparable -> Bool
(:>) cmp1 cmp2 = cmp1 > cmp2
`

func TestParseFile_OnlyFixity(t *testing.T) {
	require := require.New(t)

	p := stringParser(t, parseFixture)
	p.mode = SkipDefinitions
	var f *ast.Module
	func() {
		defer assertEOF(t, "TestParseFile_OnlyFixity", false)
		defer p.sess.Emit()
		f = parseFile(p)

		require.Len(f.Imports, 2, "should have 2 imports")
		name, ok := f.Module.Name.(fmt.Stringer)
		require.True(ok, "expected module name to be stringer")
		require.Equal("Foo", name.String(), "module name")

		require.Len(f.Decls, 2, "should have 2 decls")
		assertFixity(t, f.Decls[0], "?", 6, operator.Right)
		assertFixity(t, f.Decls[1], ":>", 7, operator.Left)
	}()
	require.True(p.sess.IsOK(), "no errors should be returned")
	require.NotNil(f)
}

func assertFixity(t *testing.T, d ast.Decl, op string, precedence uint, assoc operator.Associativity) {
	decl, ok := d.(*ast.InfixDecl)
	require.True(t, ok, "should be InfixDecl")
	require.Equal(t, op, decl.Op.Name)
	require.Equal(t, fmt.Sprint(precedence), decl.Precedence.Value)
	require.Equal(t, assoc, decl.Assoc)
}

func TestParseFull(t *testing.T) {
	require := require.New(t)
	wd, err := os.Getwd()
	require.NoError(err)
	path := filepath.Join(wd, "_testdata", "valid_fullparse", "src", "Main.elm")
	result, err := Parse(path, FullParse)
	require.NoError(err)

	require.Len(result.Modules, 3)
	require.Equal([]string{"Internal.Dependency", "Dependency", "Main"}, result.Resolution)

	mainExpected := File(
		Module("Main", OpenList),
		[]ImportAssert{
			Import("Internal.Dependency", nil, ClosedList(
				ExposedVar("maybeStr"),
			)),
			Import("Dependency", nil, ClosedList(
				ExposedVar("?"), ExposedVar("?:"),
			)),
		},
		Definition(
			"main",
			TypeAnnotation(NamedType("Program", NamedType("String"))),
			nil,
			BinaryOp(
				"?:",
				BinaryOp(
					"?",
					Identifier("maybeStr"),
					Literal(ast.String, `"hello"`),
				),
				Literal(ast.String, `"hello world"`),
			),
		),
	)

	internalDepExpected := File(
		Module("Internal.Dependency", ClosedList(ExposedVar("maybeStr"))),
		nil,
		Definition(
			"maybeStr",
			TypeAnnotation(NamedType("Maybe", NamedType("String"))),
			nil,
			FuncApp(
				Identifier("Just"),
				Literal(ast.String, `"hi"`),
			),
		),
	)

	depExpected := File(
		Module("Dependency", ClosedList(
			ExposedVar("?"), ExposedVar("?:"),
		)),
		nil,
		Definition(
			"?",
			TypeAnnotation(
				FuncType(
					NamedType("Maybe", VarType("a")),
					VarType("a"),
					VarType("a"),
				),
			),
			Patterns(VarPattern("m"), VarPattern("a")),
			FuncApp(
				Selector("Maybe", "withDefault"),
				Identifier("a"),
				Identifier("m"),
			),
		),
		InfixDecl("?", operator.Left, Literal(ast.Int, "2")),
		Definition(
			"?:",
			TypeAnnotation(
				FuncType(
					NamedType("Maybe", VarType("a")),
					VarType("a"),
					VarType("a"),
				),
			),
			Patterns(VarPattern("m"), VarPattern("a")),
			FuncApp(
				Selector("Maybe", "withDefault"),
				Identifier("a"),
				Identifier("m"),
			),
		),
	)

	cases := map[string]FileAssert{
		"Main":                mainExpected,
		"Dependency":          depExpected,
		"Internal.Dependency": internalDepExpected,
	}

	for mod, expected := range cases {
		f, ok := result.Modules[mod]
		require.True(ok, "expected module to exist: %s", mod)
		require.NotNil(f, mod)
		expected(t, f)
	}
}
