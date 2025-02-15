package parse

import (
	"errors"
	"fmt"
	"sort"

	"github.com/aggronmagi/csmsgp2go/gen"
)

// This file defines when and how we
// propagate type information from
// one type declaration to another.
// After the processing pass, every
// non-primitive type is marshalled/unmarshalled/etc.
// through a function call. Here, we propagate
// the type information into the caller's type
// tree *if* the child type is simple enough.
//
// For example, types like
//
//    type A [4]int
//
// will get pushed into parent methods,
// whereas types like
//
//    type B [3]map[string]struct{A, B [4]string}
//
// will not.

// this is an approximate measure
// of the number of children in a node
const maxComplex = 5

// begin recursive search for identities with the
// given name and replace them with e
func (f *FileSet) findShim(id string, e gen.Elem, addID bool) {
	for name, el := range f.Identities {
		pushstate(name)
		switch el := el.(type) {
		case *gen.Struct:
			for i := range el.Fields {
				f.nextShim(&el.Fields[i].FieldElem, id, e)
			}
		case *gen.Array:
			f.nextShim(&el.Els, id, e)
		case *gen.Slice:
			f.nextShim(&el.Els, id, e)
		case *gen.Map:
			f.nextShim(&el.Value, id, e)
		case *gen.Ptr:
			f.nextShim(&el.Value, id, e)
		}
		popstate()
	}
	if addID {
		f.Identities[id] = e
	}
}

func (f *FileSet) nextShim(ref *gen.Elem, id string, e gen.Elem) {
	if (*ref).TypeName() == id {
		vn := (*ref).Varname()
		*ref = e.Copy()
		(*ref).SetVarname(vn)
	} else {
		switch el := (*ref).(type) {
		case *gen.Struct:
			for i := range el.Fields {
				f.nextShim(&el.Fields[i].FieldElem, id, e)
			}
		case *gen.Array:
			f.nextShim(&el.Els, id, e)
		case *gen.Slice:
			f.nextShim(&el.Els, id, e)
		case *gen.Map:
			f.nextShim(&el.Value, id, e)
		case *gen.Ptr:
			f.nextShim(&el.Value, id, e)
		}
	}
}

// propInline identifies and inlines candidates
func (f *FileSet) propInline() (err error) {
	type gelem struct {
		name string
		el   gen.Elem
	}

	all := make([]gelem, 0, len(f.Identities))

	for name, el := range f.Identities {
		all = append(all, gelem{name: name, el: el})
	}

	// make sure we process inlining determinstically:
	// start with the least-complex elems;
	// use identifier names as a tie-breaker
	sort.Slice(all, func(i, j int) bool {
		ig, jg := &all[i], &all[j]
		ic, jc := ig.el.Complexity(), jg.el.Complexity()
		return ic < jc || (ic == jc && ig.name < jg.name)
	})

	for i := range all {
		name := all[i].name
		pushstate(name)
		switch el := all[i].el.(type) {
		case *gen.Struct:
			tagsValue := [256]uint16{}
			for i := range el.Fields {
				err2 := f.nextInline(&el.Fields[i].FieldElem, name)
				if err2 != nil {
					err = err2
				}
				if tagsValue[el.Fields[i].FieldTag] != 0 {
					err = errors.New("tag value repeated")
					warnf("tag value %d repeated", el.Fields[i].FieldTag)
					continue
				}
				tagsValue[el.Fields[i].FieldTag] = el.Fields[i].FieldTag
			}
		case *gen.Array:
			err = f.nextInline(&el.Els, name)
		case *gen.Slice:
			err = f.nextInline(&el.Els, name)
		case *gen.Map:
			err = f.nextInline(&el.Value, name)
		case *gen.Ptr:
			err = f.nextInline(&el.Value, name)
		}
		popstate()
		if err != nil {
			return err
		}
	}

	return nil
}

const fatalloop = `detected infinite recursion in inlining loop!
Please file a bug at github.com/aggronmagi/csmsgp2go/issues!
Thanks!
`

func (f *FileSet) nextInline(ref *gen.Elem, root string) (err error) {
	switch el := (*ref).(type) {
	case *gen.BaseElem:
		// ensure that we're not inlining
		// a type into itself
		typ := el.TypeName()
		if el.Value == gen.IDENT && typ != root {
			if node, ok := f.Identities[typ]; ok && node.Complexity() < maxComplex {
				infof("inlining %s\n", typ)

				// This should never happen; it will cause
				// infinite recursion.
				if node == *ref {
					panic(fatalloop)
				}

				*ref = node.Copy()
				err = f.nextInline(ref, node.TypeName())
				if err != nil {
					return err
				}
			} else if !ok && !el.Resolved() {
				// this is the point at which we're sure that
				// we've got a type that isn't a primitive,
				// a library builtin, or a processed type
				warnf("unresolved identifier: %s\n", typ)
				return fmt.Errorf("unresolved identifier: %s", typ)
			}
		}
	case *gen.Struct:
		tagsValue := [256]uint16{}
		for i := range el.Fields {
			err2 := f.nextInline(&el.Fields[i].FieldElem, root)
			if err2 != nil {
				err = err2
			}

			if tagsValue[el.Fields[i].FieldTag] != 0 {
				err = errors.New("tag value repeated")
				warnf("tag value %d repeated", el.Fields[i].FieldTag)
				continue
			}
			tagsValue[el.Fields[i].FieldTag] = el.Fields[i].FieldTag
		}
	case *gen.Array:
		return f.nextInline(&el.Els, root)
	case *gen.Slice:
		return f.nextInline(&el.Els, root)
	case *gen.Map:
		return f.nextInline(&el.Value, root)
	case *gen.Ptr:
		return f.nextInline(&el.Value, root)
	case *gen.CsharpString:
		return nil
	default:
		//panic("bad elem type")
		return errors.New("bad elem type")
	}
	return nil
}
