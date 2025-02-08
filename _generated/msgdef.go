package _generated

import "time"

//go:generate csmsgp2go

type AllowNil0 struct {
	ABool       bool       `msg:"0"`
	AInt        int        `msg:"1"`
	AInt8       int8       `msg:"2"`
	AInt16      int16      `msg:"3"`
	AInt32      int32      `msg:"4"`
	AInt64      int64      `msg:"5"`
	AUint       uint       `msg:"6"`
	AUint8      uint8      `msg:"7"`
	AUint16     uint16     `msg:"8"`
	AUint32     uint32     `msg:"9"`
	AUint64     uint64     `msg:"10"`
	AFloat32    float32    `msg:"11"`
	AFloat64    float64    `msg:"12"`
	AComplex64  complex64  `msg:"13"`
	AComplex128 complex128 `msg:"14"`

	ANamedBool    bool    `msg:"15"`
	ANamedInt     int     `msg:"16"`
	ANamedFloat64 float64 `msg:"17"`

	AMapStrStr map[string]string `msg:"18"`

	APtrNamedStr NamedString `msg:"19"`

	AString      string `msg:"20"`
	ANamedString string `msg:"21"`
	AByteSlice   []byte `msg:"22"`

	ASliceString      []string      `msg:"23"`
	ASliceNamedString []NamedString `msg:"24"`

	ANamedStruct    NamedStructAN `msg:"25"`
	APtrNamedStruct NamedStructAN `msg:"26"`

	AUnnamedStruct struct {
		A string
	} `msg:"27"` // allownil not supported on unnamed struct

	EmbeddableStructAN `msg:"32,flatten"` // embed flat

	EmbeddableStruct2AN `msg:"28"` // embed non-flat

	AArrayInt [5]int `msg:"30"` // not supported

	ATime time.Time `msg:"31"`

	MapStrStr map[string]string `msg:"35"`
	MapIntStr map[int32]string  `msg:"36"`
}

// type AllowNil1 struct {
// 	ABool       bool       `msg:"0"`
// 	AInt        int        `msg:"1"`
// 	AInt8       int8       `msg:"2"`
// 	AInt16      int16      `msg:"3"`
// 	AInt32      int32      `msg:"4"`
// 	AInt64      int64      `msg:"5"`
// 	AUint       uint       `msg:"6"`
// 	AUint8      uint8      `msg:"7"`
// 	AUint16     uint16     `msg:"8"`
// 	AUint32     uint32     `msg:"9"`
// 	AUint64     uint64     `msg:"10"`
// 	AFloat32    float32    `msg:"11"`
// 	AFloat64    float64    `msg:"12"`
// 	AComplex64  complex64  `msg:"13"`
// 	AComplex128 complex128 `msg:"14"`

// 	ANamedBool    bool    `msg:"15"`
// 	ANamedInt     int     `msg:"16"`
// 	ANamedFloat64 float64 `msg:"17"`

// 	AMapStrStr map[string]string `msg:"18"`

// 	APtrNamedStr *NamedString `msg:"19"`

// 	AString      string `msg:"20"`
// 	ANamedString string `msg:"21"`
// 	AByteSlice   []byte `msg:"22"`

// 	ASliceString      []string      `msg:"23"`
// 	ASliceNamedString []NamedString `msg:"24"`

// 	ANamedStruct    NamedStructAN  `msg:"25"`
// 	APtrNamedStruct NamedStructAN `msg:"26"`

// 	AUnnamedStruct struct {
// 		A string
// 	} `msg:"27"` // allownil not supported on unnamed struct

// 	EmbeddableStructAN `msg:"32,flatten"` // embed flat

// 	*EmbeddableStruct2AN `msg:"28"` // embed non-flat

// 	AArrayInt [5]int `msg:"30"` // not supported

// 	ATime time.Time `msg:"31"`
// }

type EmbeddableStructAN struct {
	SomeEmbed []string
	SomeInt32 []int32
}

type EmbeddableStruct2AN struct {
	SomeEmbed2 []string
}

type NamedStructAN struct {
	A []string
	B []string
}

type AllowNilHalfFull struct {
	Field00 []string `msg:"0"`
	Field01 []string `msg:"1"`
	Field02 []string `msg:"2"`
	Field03 []string `msg:"3"`
}

type AllowNilLotsOFields struct {
	Field00 []string `msg:"00"`
	Field01 []string `msg:"01"`
	Field02 []string `msg:"02"`
	Field03 []string `msg:"03"`
	Field04 []string `msg:"04"`
	Field05 []string `msg:"05"`
	Field06 []string `msg:"06"`
	Field07 []string `msg:"07"`
	Field08 []string `msg:"08"`
	Field09 []string `msg:"09"`
	Field10 []string `msg:"10"`
	Field11 []string `msg:"11"`
	Field12 []string `msg:"12"`
	Field13 []string `msg:"13"`
	Field14 []string `msg:"14"`
	Field15 []string `msg:"15"`
	Field16 []string `msg:"16"`
	Field17 []string `msg:"17"`
	Field18 []string `msg:"18"`
	Field19 []string `msg:"19"`
	Field20 []string `msg:"20"`
	Field21 []string `msg:"21"`
	Field22 []string `msg:"22"`
	Field23 []string `msg:"23"`
	Field24 []string `msg:"24"`
	Field25 []string `msg:"25"`
	Field26 []string `msg:"26"`
	Field27 []string `msg:"27"`
	Field28 []string `msg:"28"`
	Field29 []string `msg:"29"`
	Field30 []string `msg:"30"`
	Field31 []string `msg:"31"`
	Field32 []string `msg:"32"`
	Field33 []string `msg:"33"`
	Field34 []string `msg:"34"`
	Field35 []string `msg:"35"`
	Field36 []string `msg:"36"`
	Field37 []string `msg:"37"`
	Field38 []string `msg:"38"`
	Field39 []string `msg:"39"`
	Field40 []string `msg:"40"`
	Field41 []string `msg:"41"`
	Field42 []string `msg:"42"`
	Field43 []string `msg:"43"`
	Field44 []string `msg:"44"`
	Field45 []string `msg:"45"`
	Field46 []string `msg:"46"`
	Field47 []string `msg:"47"`
	Field48 []string `msg:"48"`
	Field49 []string `msg:"49"`
	Field50 []string `msg:"50"`
	Field51 []string `msg:"51"`
	Field52 []string `msg:"52"`
	Field53 []string `msg:"53"`
	Field54 []string `msg:"54"`
	Field55 []string `msg:"55"`
	Field56 []string `msg:"56"`
	Field57 []string `msg:"57"`
	Field58 []string `msg:"58"`
	Field59 []string `msg:"59"`
	Field60 []string `msg:"60"`
	Field61 []string `msg:"61"`
	Field62 []string `msg:"62"`
	Field63 []string `msg:"63"`
	Field64 []string `msg:"64"`
	Field65 []string `msg:"65"`
	Field66 []string `msg:"66"`
	Field67 []string `msg:"67"`
	Field68 []string `msg:"68"`
	Field69 []string `msg:"69"`
}

type (
	NamedBool    bool
	NamedInt     int
	NamedFloat64 float64
	NamedString  string
)

type AllowNil10 struct {
	Field00 []string `msg:"0"`
	Field01 []string `msg:"1"`
	Field02 []string `msg:"2"`
	Field03 []string `msg:"3"`
	Field04 []string `msg:"4"`
	Field05 []string `msg:"5"`
	Field06 []string `msg:"6"`
	Field07 []string `msg:"7"`
	Field08 []string `msg:"8"`
	Field09 []string `msg:"9"`
}

type NotAllowNil10 struct {
	Field00 []string `msg:"0"`
	Field01 []string `msg:"1"`
	Field02 []string `msg:"2"`
	Field03 []string `msg:"3"`
	Field04 []string `msg:"4"`
	Field05 []string `msg:"5"`
	Field06 []string `msg:"6"`
	Field07 []string `msg:"7"`
	Field08 []string `msg:"8"`
	Field09 []string `msg:"9"`
}

type AllowNilOmitEmpty struct {
	Field00 []string `msg:"0,omitempty"`
	Field01 []string `msg:"1"`
}

type AllowNilOmitEmpty2 struct {
	Field00 []string `msg:"0,omitempty"`
	Field01 []string `msg:"1,omitempty"`
}

type MapTest struct {
	MapStrStr map[string]string `msg:"1"`
	MapIntStr map[int32]string  `msg:"2"`
}

// Primitive types cannot have allownil for now.
type NoAllowNil []byte
