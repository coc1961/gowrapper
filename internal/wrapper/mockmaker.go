package wrapper

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"strings"
)

type MockMaker struct {
	StructName string
	Funcs      []*Func
	AddPackage bool
	Package    string
}

type Func struct {
	FuncName string
	Params   []*DataType
	Returns  []*DataType
}

type DataType struct {
	Name string
	Type string
}

func (mm *MockMaker) CreateMock(filePath, structName string, addPackage bool) *MockMaker {
	m := MockMaker{
		Funcs:      make([]*Func, 0),
		AddPackage: addPackage,
	}

	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, filePath, nil, 0)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	mm.AddPackage = addPackage
	mm.Package = f.Name.Name
	m.StructName = structName
	m.Package = mm.Package

	for _, dec := range f.Decls {
		if gen, ok := dec.(*ast.GenDecl); ok {
			if gen.Tok != token.TYPE {
				continue
			}
			for _, specs := range gen.Specs {
				if ts, ok := specs.(*ast.TypeSpec); ok {
					if ts.Name.String() != structName {
						continue
					}
					if iFace, ok := ts.Type.(*ast.InterfaceType); ok {
						mm.ProcessInterface(iFace, &m)
					}
				}
			}
		}
	}
	return &m
}

func (mm *MockMaker) ProcessInterface(iFace *ast.InterfaceType, m *MockMaker) {
	for _, meths := range iFace.Methods.List {
		if len(meths.Names) == 0 {
			// interface composition
			if meths.Type == nil {
				return
			}
			if id, ok := meths.Type.(*ast.Ident); ok {
				if id.Obj == nil || id.Obj.Decl == nil {
					return
				}
				if te, ok := id.Obj.Decl.(*ast.TypeSpec); ok {
					if if1, ok := te.Type.(*ast.InterfaceType); ok {
						mm.ProcessInterface(if1, m)
					}
				}
			}
			continue
		}
		for _, name := range meths.Names {
			ff := Func{
				FuncName: name.String(),
				Params:   make([]*DataType, 0),
				Returns:  make([]*DataType, 0),
			}
			m.Funcs = append(m.Funcs, &ff)

			if ft, ok := meths.Type.(*ast.FuncType); ok {
				mm.AddParams(ft, &ff)
				mm.AddReturns(ft, &ff)
			}
		}
	}
}

func (mm *MockMaker) AddParams(ft *ast.FuncType, ff *Func) {
	for num, p := range ft.Params.List {
		if p.Names != nil && len(p.Names) > 0 {
			for _, n := range p.Names {
				dt := &DataType{
					Name: n.String(),
				}
				ff.Params = append(ff.Params, dt)

				dt.Type = mm.GetType(p.Type)
			}
		} else {
			dt := &DataType{
				Name: fmt.Sprintf("parVar%d", num),
			}
			ff.Params = append(ff.Params, dt)

			dt.Type = mm.GetType(p.Type)
		}
	}
}

func (mm *MockMaker) AddReturns(ft *ast.FuncType, ff *Func) {
	if ft.Results != nil {
		for num, r := range ft.Results.List {
			if r.Names != nil && len(r.Names) > 0 {
				for _, n := range r.Names {
					dt := &DataType{
						Name: n.String(),
					}

					ff.Returns = append(ff.Returns, dt)

					dt.Type = mm.GetType(r.Type)
				}
			} else {
				dt := &DataType{
					Name: fmt.Sprintf("retVar%d", num),
				}

				ff.Returns = append(ff.Returns, dt)

				dt.Type = mm.GetType(r.Type)
			}
		}
	}
}

func (mm *MockMaker) GetType(e ast.Expr) string {
	str := types.ExprString(e)

	if mm.AddPackage {
		switch t := e.(type) {
		case *ast.MapType:
			return "map[" + mm.GetType(t.Key) + "]" + mm.GetType(t.Value)
		case *ast.Ellipsis:
			str = "..." + mm.GetType(t.Elt)
		case *ast.StarExpr:
			str = "*" + mm.GetType(t.X)
		case *ast.SelectorExpr:
			if !strings.Contains(str, ".") {
				str = mm.Package + "." + str
			}
		case *ast.Ident:
			if !mm.isBasic(t.Name) && !strings.Contains(str, ".") {
				str = mm.Package + "." + str
			}
		case *ast.ArrayType:
			str = "[]" + mm.GetType(t.Elt)
		}
	}

	return str
}

func (mm *MockMaker) isBasic(typ string) bool {
	basicTypes := []string{
		"string", "bool", "int8", "uint8", "int16",
		"uint16", "int32", "uint32", "int64", "uint64", "int", "uint",
		"uintptr", "float32", "float64", "complex64", "complex128", "error",
	}
	for _, s := range basicTypes {
		if typ == s || typ == "*"+typ {
			return true
		}
	}
	return false
}

func (mm *MockMaker) String() string {
	if len(mm.Funcs) == 0 {
		return ""
	}

	var str bytes.Buffer

	c := func(s ...string) {
		for _, s1 := range s {
			str.WriteString(s1)
		}
	}
	c("\n// New", mm.StructName, "Wrapper wrapper for ", mm.StructName, " interface.\n")
	c("func New", mm.StructName, "Wrapper(service ", mm.StructName, ",",
		"\n\tpreCallFunc func(name string,param ... interface{}), ",
		"\n\tposCallFunc func(name string,retValue ... interface{}),\n\t) ",
		"*", mm.StructName, "Wrapper {\n")
	c("\treturn &", mm.StructName, "Wrapper {\n")

	c("\t\tservice:\tservice,\n")
	c("\t\tpreCallFunc:\tpreCallFunc,\n")
	c("\t\tposCallFunc:\tposCallFunc,\n")

	c("\t}\n")
	c("}\n")

	c("\n// Wrapper for ", mm.StructName, " interface.\n")
	c("type ", mm.StructName, "Wrapper struct {\n")
	c("\tservice \t", mm.StructName, "\n")
	c("\tpreCallFunc func(name string,param ... interface{})\n")
	c("\tposCallFunc func(name string,retValue ... interface{})\n")
	c("}\n\n")
	for _, f := range mm.Funcs {
		c("// ", f.FuncName, " function.\n")
		c("func (m *", mm.StructName, "Wrapper) ", f.FuncName, "(")
		coma := ""
		for i, p := range f.Params {
			if p.Name == "" {
				p.Name = fmt.Sprintf("param%v", i)
			}
			c(coma, p.Name, " ", p.Type)
			coma = ", "
		}
		c(") (")
		c("\t")
		coma = ""
		for _, p := range f.Returns {
			if p.Name != "" {
				c(coma, p.Name, " ", p.Type)
			}
			coma = ", "
		}
		c(") {\n")

		c("\tm.preCallFunc(\"", f.FuncName, "\"")
		coma = ","
		for i, p := range f.Params {
			if p.Name == "" {
				p.Name = fmt.Sprintf("&param%v", i)
			}
			c(coma, "&"+p.Name)
			coma = ", "
		}
		c(")\n")

		c("\t")
		coma = ""
		for _, p := range f.Returns {
			if p.Name != "" {
				c(coma, p.Name)
			}
			coma = ", "
		}
		if len(f.Returns) > 0 {
			c(" = ")
		}

		c("m.service.", f.FuncName, "(")
		coma = ""
		for i, p := range f.Params {
			if p.Name == "" {
				p.Name = fmt.Sprintf("param%v", i)
			}
			c(coma, p.Name)
			coma = ", "
		}
		c(")\n")

		c("\tm.posCallFunc(\"", f.FuncName, "\"")
		coma = ","
		for _, p := range f.Returns {
			if p.Name != "" {
				c(coma, p.Name)
			}
			coma = ", "
		}
		c(")\n")

		c("\treturn\n")

		c("}\n\n")
	}

	return str.String()
}
