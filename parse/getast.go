package parse

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/aggronmagi/csmsgp2go/gen"
)

// A FileSet is the in-memory representation of a
// parsed file.
type FileSet struct {
	Package       string              // package name
	Specs         map[string]ast.Expr // type specs in file
	Identities    map[string]gen.Elem // processed from specs
	Directives    []string            // raw preprocessor directives
	Imports       []*ast.ImportSpec   // imports
	CompactFloats bool                // Use smaller floats when feasible
	ClearOmitted  bool                // Set omitted fields to zero value
	NewTime       bool                // Set to use -1 extension for time.Time
	tagName       string              // tag to read field names from
	pointerRcv    bool                // generate with pointer receivers.

	FSet *token.FileSet // use for prompt error
}

// File parses a file at the relative path
// provided and produces a new *FileSet.
// If you pass in a path to a directory, the entire
// directory will be parsed.
// If unexport is false, only exported identifiers are included in the FileSet.
// If the resulting FileSet would be empty, an error is returned.
func File(name string, unexported bool) (*FileSet, error) {
	pushstate(name)
	defer popstate()
	fs := &FileSet{
		Specs:      make(map[string]ast.Expr),
		Identities: make(map[string]gen.Elem),
	}

	fset := token.NewFileSet()
	finfo, err := os.Stat(name)
	if err != nil {
		return nil, err
	}
	if finfo.IsDir() {
		pkgs, err := parser.ParseDir(fset, name, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		if len(pkgs) != 1 {
			return nil, fmt.Errorf("multiple packages in directory: %s", name)
		}
		var one *ast.Package
		for _, nm := range pkgs {
			one = nm
			break
		}
		fs.Package = one.Name
		for _, fl := range one.Files {
			pushstate(fl.Name.Name)
			fs.Directives = append(fs.Directives, yieldComments(fl.Comments)...)
			if !unexported {
				ast.FileExports(fl)
			}
			fs.getTypeSpecs(fl)
			popstate()
		}
	} else {
		f, err := parser.ParseFile(fset, name, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		fs.Package = f.Name.Name
		fs.Directives = yieldComments(f.Comments)
		if !unexported {
			ast.FileExports(f)
		}
		fs.getTypeSpecs(f)
	}

	if len(fs.Specs) == 0 {
		return nil, fmt.Errorf("no definitions in %s", name)
	}

	fs.FSet = fset

	if err = fs.applyEarlyDirectives(); err != nil {
		return nil, err
	}
	if err = fs.process(); err != nil {
		return nil, err
	}
	if err = fs.applyDirectives(); err != nil {
		return nil, err
	}
	if err = fs.propInline(); err != nil {
		return nil, err
	}

	fs.sortAndFillMsgFields()

	return fs, nil
}

// Format printer.Fprint wrap.
func (f *FileSet) Format(node interface{}) string {
	buf := &bytes.Buffer{}
	_ = printer.Fprint(buf, f.FSet, node)
	return buf.String()
}

func (f *FileSet) sortAndFillMsgFields() {
	spaceHolder := &gen.NilPlaceholder{}
	for _, elem := range f.Identities {
		// fix struct
		if s, ok := elem.(*gen.Struct); ok {
			if len(s.Fields) < 1 {
				continue
			}
			flagSet := make(map[uint16]struct{})
			maxId := uint16(0)
			for _, field := range s.Fields {
				flagSet[field.FieldTag] = struct{}{}
				if field.FieldTag > maxId {
					maxId = field.FieldTag
				}
				field.FieldElem = fixCsharpString(field.FieldElem)
			}
			if int(maxId)+1 == len(s.Fields) {
				continue
			}
			for i := uint16(0); i <= maxId; i++ {
				if _, ok := flagSet[i]; ok {
					continue
				}
				s.Fields = append(s.Fields, gen.StructField{
					FieldTag:  i,
					FieldElem: spaceHolder,
				})
			}
			sort.Slice(s.Fields, func(i, j int) bool {
				return s.Fields[i].FieldTag < s.Fields[j].FieldTag
			})
		}
	}
}

func fixCsharpString(elem gen.Elem) gen.Elem {
	switch v := elem.(type) {
	case *gen.BaseElem:
		if v.Value == gen.String {
			return &gen.CsharpString{}
		}
	case *gen.Ptr:
		v.Value = fixCsharpString(v.Value)
	case *gen.Map:
		// map key not set csharp string.
		v.Value = fixCsharpString(v.Value)
	case *gen.Slice:
		v.Els = fixCsharpString(v.Els)
	case *gen.Array:
		v.Els = fixCsharpString(v.Els)
	case *gen.Struct:
		for i := range v.Fields {
			v.Fields[i].FieldElem = fixCsharpString(v.Fields[i].FieldElem)
		}
	}
	return elem
}

// applyDirectives applies all of the directives that
// are known to the parser. additional method-specific
// directives remain in f.Directives
func (f *FileSet) applyDirectives() error {
	newdirs := make([]string, 0, len(f.Directives))
	for _, d := range f.Directives {
		chunks := strings.Split(d, " ")
		if len(chunks) > 0 {
			if fn, ok := directives[chunks[0]]; ok {
				pushstate(chunks[0])
				err := fn(chunks, f)
				if err != nil {
					warnf("directive error: %s", err)
					return fmt.Errorf("directive %v error: %w", chunks, err)
				}
				popstate()
			} else {
				newdirs = append(newdirs, d)
			}
		}
	}
	f.Directives = newdirs
	return nil
}

// applyEarlyDirectives applies all early directives needed before process() is called.
// additional directives remain in f.Directives for future processing
func (f *FileSet) applyEarlyDirectives() error {
	newdirs := make([]string, 0, len(f.Directives))
	for _, d := range f.Directives {
		parts := strings.Split(d, " ")
		if len(parts) == 0 {
			continue
		}
		if fn, ok := earlyDirectives[parts[0]]; ok {
			pushstate(parts[0])
			err := fn(parts, f)
			if err != nil {
				warnf("early directive error: %s", err)
				err = fmt.Errorf("early directive error: %w", err)
				return err
			}
			popstate()
		} else {
			newdirs = append(newdirs, d)
		}
	}
	f.Directives = newdirs
	return nil
}

// A linkset is a graph of unresolved
// identities.
//
// Since gen.Ident can only represent
// one level of type indirection (e.g. Foo -> uint8),
// type declarations like `type Foo Bar`
// aren't resolve-able until we've processed
// everything else.
//
// The goal of this dependency resolution
// is to distill the type declaration
// into just one level of indirection.
// In other words, if we have:
//
//	type A uint64
//	type B A
//	type C B
//	type D C
//
// ... then we want to end up
// figuring out that D is just a uint64.
type linkset map[string]*gen.BaseElem

func (f *FileSet) resolve(ls linkset) error {
	progress := true
	for progress && len(ls) > 0 {
		progress = false
		for name, elem := range ls {
			real, ok := f.Identities[elem.TypeName()]
			if ok {
				// copy the old type descriptor,
				// alias it to the new value,
				// and insert it into the resolved
				// identities list
				progress = true
				nt := real.Copy()
				nt.Alias(name)
				f.Identities[name] = nt
				delete(ls, name)
			}
		}
	}

	if len(ls) < 1 {
		return nil
	}

	// what's left can't be resolved
	errTip := "couldn't resolve type."
	for name, elem := range ls {
		warnf("couldn't resolve type %s (%s)\n", name, elem.TypeName())
		errTip += fmt.Sprintf(" %s-(%s)", name, elem.TypeName())
	}
	return errors.New(errTip)
}

// process takes the contents of f.Specs and
// uses them to populate f.Identities
func (f *FileSet) process() error {
	deferred := make(linkset)
parse:
	for name, def := range f.Specs {
		pushstate(name)
		el, err := f.parseExpr(def)
		if err != nil {
			return fmt.Errorf("parse %s failed,%w", f.Format(def), err)
		}
		if el == nil {
			warnf("failed to parse %s", f.Format(def))
			popstate()
			continue parse
		}
		el.AlwaysPtr(&f.pointerRcv)
		// push unresolved identities into
		// the graph of links and resolve after
		// we've handled every possible named type.
		if be, ok := el.(*gen.BaseElem); ok && be.Value == gen.IDENT {
			deferred[name] = be
			popstate()
			continue parse
		}
		el.Alias(name)
		f.Identities[name] = el
		popstate()
	}

	if len(deferred) > 0 {
		return f.resolve(deferred)
	}
	return nil
}

func strToMethod(s string) gen.Method {
	switch s {
	case "encode":
		return gen.Encode
	case "decode":
		return gen.Decode
	case "test":
		return gen.Test
	case "size":
		return gen.Size
	case "marshal":
		return gen.Marshal
	case "unmarshal":
		return gen.Unmarshal
	default:
		return 0
	}
}

func (f *FileSet) applyDirs(p *gen.Printer) {
	// apply directives of the form
	//
	// 	//msgp:encode ignore {{TypeName}}
	//
loop:
	for _, d := range f.Directives {
		chunks := strings.Split(d, " ")
		if len(chunks) > 1 {
			for i := range chunks {
				chunks[i] = strings.TrimSpace(chunks[i])
			}
			m := strToMethod(chunks[0])
			if m == 0 {
				warnf("unknown pass name: %q\n", chunks[0])
				continue loop
			}
			if fn, ok := passDirectives[chunks[1]]; ok {
				pushstate(chunks[1])
				err := fn(m, chunks[2:], p)
				if err != nil {
					warnf("error applying directive: %s\n", err)
				}
				popstate()
			} else {
				warnf("unrecognized directive %q\n", chunks[1])
			}
		} else {
			warnf("empty directive: %q\n", d)
		}
	}
	p.CompactFloats = f.CompactFloats
	p.ClearOmitted = f.ClearOmitted
	p.NewTime = f.NewTime
}

func (f *FileSet) PrintTo(p *gen.Printer) error {
	f.applyDirs(p)
	names := make([]string, 0, len(f.Identities))
	for name := range f.Identities {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		el := f.Identities[name]
		el.SetVarname("z")
		pushstate(el.TypeName())
		err := p.Print(el)
		popstate()
		if err != nil {
			return err
		}
	}
	return nil
}

// getTypeSpecs extracts all of the *ast.TypeSpecs in the file
// into fs.Identities, but does not set the actual element
func (fs *FileSet) getTypeSpecs(f *ast.File) {
	// collect all imports...
	fs.Imports = append(fs.Imports, f.Imports...)

	// check all declarations...
	for i := range f.Decls {
		// for GenDecls...
		if g, ok := f.Decls[i].(*ast.GenDecl); ok {
			// and check the specs...
			for _, s := range g.Specs {
				// for ast.TypeSpecs....
				if ts, ok := s.(*ast.TypeSpec); ok {
					switch ts.Type.(type) {
					// this is the list of parse-able
					// type specs
					case *ast.StructType,
						*ast.ArrayType,
						*ast.StarExpr,
						*ast.MapType,
						*ast.Ident:
						fs.Specs[ts.Name.Name] = ts.Type
					}
				}
			}
		}
	}
}

func fieldName(f *ast.Field) string {
	switch len(f.Names) {
	case 0:
		return stringify(f.Type)
	case 1:
		return f.Names[0].Name
	default:
		return f.Names[0].Name + " (and others)"
	}
}

func (fs *FileSet) parseFieldList(fl *ast.FieldList, maxIdx *uint16) ([]gen.StructField, error) {
	if fl == nil || fl.NumFields() == 0 {
		return nil, nil
	}
	out := make([]gen.StructField, 0, fl.NumFields())
	for _, field := range fl.List {
		pushstate(fieldName(field))
		fds, err := fs.getField(field, maxIdx)
		if err != nil {
			return nil, fmt.Errorf("field %s parse failed,%w", fs.Format(fl), err)
		}
		if len(fds) > 0 {
			out = append(out, fds...)
		} else {
			warnf("ignored")
		}
		popstate()
	}
	return out, nil
}

// translate *ast.Field into []gen.StructField
func (fs *FileSet) getField(f *ast.Field, maxIdx *uint16) (sf []gen.StructField, err error) {
	sf = make([]gen.StructField, 1)
	var extension, flatten bool
	// parse tag; otherwise field name is field tag
	if f.Tag != nil {
		var body string
		if fs.tagName != "" {
			body = reflect.StructTag(strings.Trim(f.Tag.Value, "`")).Get(fs.tagName)
		}
		if body == "" {
			body = reflect.StructTag(strings.Trim(f.Tag.Value, "`")).Get("msg")
		}
		if body == "" {
			body = reflect.StructTag(strings.Trim(f.Tag.Value, "`")).Get("msgpack")
		}
		tags := strings.Split(body, ",")
		if len(tags) >= 2 {
			switch tags[1] {
			case "extension":
				extension = true
			case "flatten":
				flatten = true
			}
		}
		// ignore "-" fields
		if tags[0] == "-" {
			sf = nil
			return
		}
		if !flatten {
			if len(tags[0]) > 0 {
				idx, err := strconv.ParseUint(tags[0], 10, 16)
				if err != nil {
					return nil, fmt.Errorf("invalid index %q: %w", tags[0], err)
				}
				sf[0].FieldTag = uint16(idx)
				*maxIdx = uint16(idx) + 1
			} else {
				sf[0].FieldTag = *maxIdx
				*maxIdx++
			}

			sf[0].FieldTagParts = tags
			sf[0].RawTag = f.Tag.Value
		} else {
			if len(tags[0]) > 0 {
				idx, err := strconv.ParseUint(tags[0], 10, 16)
				if err != nil {
					return nil, fmt.Errorf("invalid index %q: %w", tags[0], err)
				}
				*maxIdx = uint16(idx)
			}
			if len(f.Names) != 0 {
				return nil, fmt.Errorf("flatten field must be anonymous field")
			}
		}
	} else {
		sf[0].FieldTag = *maxIdx
		*maxIdx++
	}

	ex, err := fs.parseExpr(f.Type)
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return nil, nil
	}

	// parse field name
	switch len(f.Names) {
	case 0:
		if flatten {
			return fs.getFieldsFromEmbeddedStruct(f.Type, maxIdx)
		} else {
			sf[0].FieldName = embedded(f.Type)
		}
	case 1:
		sf[0].FieldName = f.Names[0].Name
	default:
		//return nil, fmt.Errorf("multiple field names not supported: %s", f.Names)
		// this is for a multiple in-line declaration,
		// e.g. type A struct { One, Two int }
		sf = sf[0:0]
		for _, nm := range f.Names {
			sf = append(sf, gen.StructField{
				FieldTag:  *maxIdx, //nm.Name,
				FieldName: nm.Name,
				FieldElem: ex.Copy(),
			})
			*maxIdx++
		}
		return sf, nil
	}
	sf[0].FieldElem = ex
	// if sf[0].FieldTag == "" {
	// 	sf[0].FieldTag = sf[0].FieldName
	// 	if len(sf[0].FieldTagParts) <= 1 {
	// 		sf[0].FieldTagParts = []string{sf[0].FieldTag}
	// 	} else {
	// 		sf[0].FieldTagParts = append([]string{sf[0].FieldName}, sf[0].FieldTagParts[1:]...)
	// 	}
	// }

	// validate extension
	if extension {
		switch ex := ex.(type) {
		case *gen.Ptr:
			if b, ok := ex.Value.(*gen.BaseElem); ok {
				b.Value = gen.Ext
			} else {
				warnf("couldn't cast to extension.")
				return nil, fmt.Errorf("couldn't cast to extension %s.", fs.Format(ex.Value))
			}
		case *gen.BaseElem:
			ex.Value = gen.Ext
		default:
			warnf("couldn't cast to extension.")
			return nil, fmt.Errorf("couldn't cast to extension %s.", fs.Format(ex))
		}
	}
	return sf, nil
}

func (fs *FileSet) getFieldsFromEmbeddedStruct(f ast.Expr, maxIdx *uint16) ([]gen.StructField, error) {
	switch f := f.(type) {
	case *ast.Ident:
		s := fs.Specs[f.Name]
		switch s := s.(type) {
		case *ast.StructType:
			return fs.parseFieldList(s.Fields, maxIdx)
		default:
			warnf("%s disabled, not struct type", fs.Format(f))
			return nil, fmt.Errorf("%s not struct type", fs.Format(f))
		}
	case *ast.StarExpr:
		warnf("%s disabled, StarExpr not support type", fs.Format(f))
		return nil, fmt.Errorf("%s not support StarExpr type", fs.Format(f))
	default:
		// other possibilities are disallowed
		warnf("%s other possibilities are disallowed", fs.Format(f))
		return nil, fmt.Errorf("%s. other possibilities are disallowed", fs.Format(f)) // errors.New("other possibilities are disallowed")
	}
}

// extract embedded field name
//
// so, for a struct like
//
//		type A struct {
//			io.Writer
//	 }
//
// we want "Writer"
func embedded(f ast.Expr) string {
	switch f := f.(type) {
	case *ast.Ident:
		return f.Name
	case *ast.StarExpr:
		return embedded(f.X)
	case *ast.SelectorExpr:
		return f.Sel.Name
	default:
		// other possibilities are disallowed
		return ""
	}
}

// stringify a field type name
func stringify(e ast.Expr) string {
	switch e := e.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + stringify(e.X)
	case *ast.SelectorExpr:
		return stringify(e.X) + "." + e.Sel.Name
	case *ast.ArrayType:
		if e.Len == nil {
			return "[]" + stringify(e.Elt)
		}
		return fmt.Sprintf("[%s]%s", stringify(e.Len), stringify(e.Elt))
	case *ast.InterfaceType:
		if e.Methods == nil || e.Methods.NumFields() == 0 {
			return "interface{}"
		}
	}
	return "<BAD>"
}

// recursively translate ast.Expr to gen.Elem; nil means type not supported
// expected input types:
// - *ast.MapType (map[T]J)
// - *ast.Ident (name)
// - *ast.ArrayType ([(sz)]T)
// - *ast.StarExpr (*T)
// - *ast.StructType (struct {})
// - *ast.SelectorExpr (a.B)
// - *ast.InterfaceType (interface {})
func (fs *FileSet) parseExpr(e ast.Expr) (gen.Elem, error) {
	switch e := e.(type) {

	case *ast.MapType:
		// parse key type
		kt, err := fs.parseExpr(e.Key)
		if err != nil {
			return nil, err
		}
		if kt == nil {
			return nil, nil
		}
		// check map key valid key
		kname := kt.TypeName()
		if kname != "string" && !strings.HasPrefix(kname, "uint") && !strings.HasPrefix(kname, "int") {
			// 仅支持 string,int...,uint...
			return nil, errors.New("map key only support string, int...,uint...")
		}

		// parse value type
		value, err := fs.parseExpr(e.Value)
		if err != nil {
			return nil, fmt.Errorf("map value parse failed,%w", err)
		}
		if value != nil {
			return &gen.Map{Key: kt, Value: value}, nil
		}
		return nil, fmt.Errorf("not support map value type")

	case *ast.Ident:
		b := gen.Ident(e.Name)

		// work to resolve this expression
		// can be done later, once we've resolved
		// everything else.
		if b.Value == gen.IDENT {
			if _, ok := fs.Specs[e.Name]; !ok {
				warnf("non-local identifier: %s\n", e.Name)
			} else {
				println("local identifier: ", e.Name)
			}
		}
		return b, nil

	case *ast.ArrayType:

		// special case for []byte
		if e.Len == nil {
			if i, ok := e.Elt.(*ast.Ident); ok && i.Name == "byte" {
				return &gen.BaseElem{Value: gen.Bytes}, nil
			}
		}

		// return early if we don't know
		// what the slice element type is
		els, err := fs.parseExpr(e.Elt)
		if err != nil {
			return nil, err
		}
		if els == nil {
			return nil, nil
		}

		// array and not a slice
		if e.Len != nil {
			switch s := e.Len.(type) {
			case *ast.BasicLit:
				return &gen.Array{
					Size: s.Value,
					Els:  els,
				}, nil

			case *ast.Ident:
				return &gen.Array{
					Size: s.String(),
					Els:  els,
				}, nil

			case *ast.SelectorExpr:
				return &gen.Array{
					Size: stringify(s),
					Els:  els,
				}, nil

			default:
				return nil, nil
			}
		}
		return &gen.Slice{Els: els}, nil

	case *ast.StarExpr:
		// if v, err := fs.parseExpr(e.X); err != nil {
		// 	return nil, err
		// } else if v != nil {
		// 	return &gen.Ptr{Value: v}, nil
		// }
		// return nil, fmt.Errorf("star expr [%s] parse failed", fs.Format(e))
		// NOTE: 不支持指针类型,否则会导致行为和csharp不一致
		return nil, fmt.Errorf("not support star expr [%s]", fs.Format(e))

	case *ast.StructType:
		var maxIdx uint16
		fields, err := fs.parseFieldList(e.Fields, &maxIdx)
		if err != nil {
			return nil, err
		}

		return &gen.Struct{Fields: fields}, nil

	case *ast.SelectorExpr:
		return gen.Ident(stringify(e)), nil

	case *ast.InterfaceType:
		// support `interface{}`
		if len(e.Methods.List) == 0 {
			return &gen.BaseElem{Value: gen.Intf}, nil
		}
		return nil, errors.New("invalid interface type")

	default: // other types not supported
		return nil, errors.New("types not supported")
	}
}

var Logf func(s string, v ...interface{})

func infof(s string, v ...interface{}) {
	if Logf != nil {
		pushstate(s)
		Logf("info: "+strings.Join(logctx, ": "), v...)
		popstate()
	}
}

func warnf(s string, v ...interface{}) {
	if Logf != nil {
		pushstate(s)
		Logf("warn: "+strings.Join(logctx, ": "), v...)
		popstate()
	}
}

var logctx []string

// push logging state
func pushstate(s string) {
	logctx = append(logctx, s)
}

// pop logging state
func popstate() {
	logctx = logctx[:len(logctx)-1]
}
