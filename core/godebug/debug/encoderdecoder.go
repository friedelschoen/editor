package debug

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"unsafe"
)

var edReg = newEDRegistry()

//----------
//----------
//----------

// encode/decode, id/type, registry
type EDRegistry struct {
	idc      EDRegId
	typeToId map[reflect.Type]EDRegId
	idToType map[EDRegId]reflect.Type

	nilId EDRegId
	//nilPointerId   EDRegId
	//nilInterfaceId EDRegId
}

func newEDRegistry() *EDRegistry {
	reg := &EDRegistry{}
	reg.typeToId = map[reflect.Type]EDRegId{}
	reg.idToType = map[EDRegId]reflect.Type{}

	// default entries
	// start at non-zero to detect zero value as error
	reg.idc = 5
	reg.nilId = reg.newId(nil)
	//reg.nilPointerId = reg.newId(nil)
	//reg.nilInterfaceId = reg.newId(nil)
	reg.idc = 10

	return reg
}
func (reg *EDRegistry) register(v any) EDRegId {
	typ := bypassPointersAndInterfaces(reflect.ValueOf(v)).Type()
	id, ok := reg.typeToId[typ]
	if ok {
		return id
	}
	id = reg.newId(typ)
	reg.typeToId[typ] = id
	reg.idToType[id] = typ
	return id
}
func (reg *EDRegistry) newId(typ reflect.Type) EDRegId {
	id := reg.idc
	reg.idc++

	// DEBUG
	//fmt.Printf("reg: %v: %v\n", id, typ)

	return id
}

//----------

type EDRegId byte

//----------
//----------
//----------

func encode(w io.Writer, v any, logOn bool, logPrefix string) error {
	enc := newEncoder(w, edReg)
	enc.logOn = logOn
	enc.logPrefix = logPrefix + enc.logPrefix
	return enc.reflect(v)
}
func decode(r io.Reader, v any, logOn bool, logPrefix string) error {
	dec := newDecoder(r, edReg)
	dec.logOn = logOn
	dec.logPrefix = logPrefix + dec.logPrefix
	return dec.reflect(v)

}

//----------
//----------
//----------

type Encoder struct {
	w   io.Writer
	reg *EDRegistry

	Logger
}

func newEncoder(w io.Writer, reg *EDRegistry) *Encoder {
	enc := &Encoder{w: w, reg: reg}
	enc.logPrefix = "enc: "
	return enc
}

//----------
//----------
//----------

type Decoder struct {
	r              io.Reader
	reg            *EDRegistry
	firstInterface bool
	firstPointer   bool

	Logger
}

func newDecoder(r io.Reader, reg *EDRegistry) *Decoder {
	dec := &Decoder{r: r, reg: reg}
	dec.logPrefix = "dec: "
	return dec
}

//----------
//----------
//----------

func (enc *Encoder) sliceLen(n int) error {
	return enc.writeBinary(uint16(n))
}
func (dec *Decoder) sliceLen(v *int) error {
	n := uint16(0)
	err := dec.readBinary(&n)
	*v = int(n)
	return err
}

//----------

func (enc *Encoder) id(id EDRegId) error {
	return enc.writeBinary(id)
}
func (dec *Decoder) id() (EDRegId, error) {
	id := EDRegId(0)
	err := dec.readBinary(&id)
	return id, err
}

//----------

func (enc *Encoder) id2(v reflect.Value) error {
	typ := bypassPointersAndInterfaces(v).Type()
	id, ok := enc.reg.typeToId[typ]
	if !ok {
		return enc.errorf("type has no id: %v", typ)
	}
	return enc.id(id)
}
func (dec *Decoder) id2(id EDRegId) (reflect.Type, error) {
	typ, ok := dec.reg.idToType[id]
	if !ok {
		return nil, dec.errorf("id has no type: %v", id)
	}
	return typ, nil
}

//----------

func (enc *Encoder) reflect(v any) error {
	// log encoded bytes at the end
	if enc.logOn {
		buf := &bytes.Buffer{}
		enc.w = io.MultiWriter(enc.w, buf)
		defer func() {
			enc.logf("encoded byte: %v\n", buf.Bytes())
		}()
	}

	enc.logf("reflect: %T\n", v)

	vv := reflect.ValueOf(v)

	switch vv.Kind() {
	case reflect.Pointer:
	case reflect.Interface:
	default:
		// use interface for other type in order to have an id
		//vv = reflect.ValueOf(vv.Interface())

		return enc.errorf("not a pointer or interface: %T", v)
	}

	return enc.reflect2(vv)
}
func (dec *Decoder) reflect(v any) error {
	dec.logf("%T\n", v)

	vv := reflect.ValueOf(v)
	switch vv.Kind() {
	case reflect.Pointer:
	default:
		return dec.errorf("not a pointer: %T", v)
	}

	if vv.IsNil() {
		vv.Set(reflect.New(vv.Type().Elem()))
	}

	switch vv.Elem().Kind() {
	case reflect.Pointer:
		vv = vv.Elem()
	case reflect.Interface:
		dec.firstInterface = true // allow not knowing the first type
		vv = vv.Elem()
	}

	return dec.reflect2(vv)
}

//----------

func (enc *Encoder) reflect2(v reflect.Value) error {
	enc.logf("reflect2: %v\n", v.Type())

	switch v.Kind() {
	case reflect.Pointer:
		// has always an id because it can be nil
		if v.IsNil() {
			//return enc.id(enc.reg.nilPointerId)
			return enc.id(enc.reg.nilId)
		}
		if err := enc.id2(v); err != nil {
			return err
		}

		return enc.reflect2(v.Elem())
	case reflect.Interface:
		// has always an id because it can be nil
		if v.IsNil() {
			//return enc.id(enc.reg.nilInterfaceId)
			return enc.id(enc.reg.nilId)
		}
		if err := enc.id2(v); err != nil {
			return err
		}

		return enc.reflect2(v.Elem())
	case reflect.Struct:
		n := v.NumField()
		vt := v.Type()
		for i := 0; i < n; i++ {
			// embedded fields
			if vt.Field(i).Anonymous {
				continue
			}

			vf := v.Field(i)
			if err := enc.reflect2(vf); err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice:
		n := v.Len()
		if err := enc.sliceLen(n); err != nil {
			return err
		}

		// fast path for []byte
		if b, ok := v.Interface().([]byte); ok {
			return enc.writeBinary(b)
		}

		for i := 0; i < n; i++ {
			vi := v.Index(i)
			if err := enc.reflect2(vi); err != nil {
				return err
			}
		}
		return nil
	case reflect.String:
		u := []byte(v.Interface().(string))
		return enc.reflect2(reflect.ValueOf(u))

	case reflect.Int: // int64, 8 bytes
		u := int64(v.Interface().(int))
		bs := make([]byte, 8)
		binary.BigEndian.PutUint64(bs, uint64(u))
		_, err := enc.w.Write(bs)
		return err

	default:
		return enc.writeBinary(v.Interface())
	}
}
func (dec *Decoder) reflect2(v reflect.Value) error {
	dec.logf("reflect2: %v\n", v.Type())

	switch v.Kind() {
	case reflect.Pointer:
		// handle id
		id, err := dec.id()
		if err != nil {
			return err
		}
		dec.logf("\tdecode pointer id: %v\n", id)
		//if id == dec.reg.nilPointerId {
		if id == dec.reg.nilId {
			v.Set(reflect.Zero(v.Type()))
			return nil
		}
		typ, err := dec.id2(id)
		if err != nil {
			return err
		}

		if !typ.AssignableTo(v.Type().Elem()) {
			return dec.errorf("%v not assignable to %v", typ, v.Type().Elem())
		}

		if v.IsNil() {
			v.Set(reflect.New(typ))
		}

		return dec.reflect2(v.Elem())

	case reflect.Interface:
		// handle id
		id, err := dec.id()
		if err != nil {
			return err
		}
		dec.logf("\tdecode interface id: %v\n", id)
		//if id == dec.reg.nilInterfaceId {
		if id == dec.reg.nilId {
			v.Set(reflect.Zero(v.Type()))
			return nil
		}
		typ, err := dec.id2(id)
		if err != nil {
			return err
		}

		if !typ.AssignableTo(v.Type()) {
			return dec.errorf("%s not assignable to %s", typ, v.Type())
		}

		// assign a pointer of the type
		v.Set(reflect.New(typ))

		if dec.firstInterface {
			dec.logf("\tfirstinterface\n")
			dec.firstInterface = false
			v = v.Elem() // bypass the need for a ptr
		}

		return dec.reflect2(v.Elem())

	case reflect.Struct:
		n := v.NumField()
		vt := v.Type()
		for i := 0; i < n; i++ {
			// embedded fields
			if vt.Field(i).Anonymous {
				continue
			}

			vf := v.Field(i)
			if err := dec.reflect2(vf); err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice:
		n := 0
		if err := dec.sliceLen(&n); err != nil {
			return err
		}
		dec.logf("dec slice len: %v\n", n)

		// fast path for bytes
		if _, ok := v.Interface().([]byte); ok {
			b := make([]byte, n)
			if _, err := io.ReadFull(dec.r, b); err != nil {
				return err
			}
			v.Set(reflect.ValueOf(b))
			return nil
		}

		v.Set(reflect.MakeSlice(v.Type(), n, n))
		for i := 0; i < n; i++ {
			vi := v.Index(i)
			if err := dec.reflect2(vi); err != nil {
				return err
			}
		}
		return nil
	case reflect.String:
		u := []byte{}
		ut := reflect.ValueOf(&u).Elem()
		if err := dec.reflect2(ut); err != nil {
			return err
		}
		v.Set(reflect.ValueOf(string(u)))
		return nil
	case reflect.Int: // int64, 8 bytes
		ptr := unsafe.Pointer(v.UnsafeAddr())
		_, err := io.ReadFull(dec.r, (*[8]byte)(ptr)[:])
		return err
	default:
		return dec.readBinary(v.Addr().Interface())
	}
}

//----------

func (enc *Encoder) writeBinary(v any) error {
	if err := binary.Write(enc.w, binary.BigEndian, v); err != nil {
		return enc.errorf("writeBinary(%T): %w", v, err)
	}
	return nil
}
func (dec *Decoder) readBinary(v any) error {
	if err := binary.Read(dec.r, binary.BigEndian, v); err != nil {
		return dec.errorf("readBinary(%T): %w", v, err)
	}
	return nil
}

//----------
//----------
//----------

type Logger struct {
	logOn     bool
	logPrefix string
}

func (l *Logger) logf(f string, args ...any) {
	if l.logOn {
		// TODO: pass whitespace to before the prefix?
		f = l.logPrefix + f
		fmt.Printf(f, args...)
	}
}
func (l *Logger) errorf(f string, args ...any) error {
	return l.error(fmt.Errorf(f, args...))
}
func (l *Logger) error(err error) error {
	return fmt.Errorf("%v%w", l.logPrefix, err)
}

//----------

func wrapErrorWithType(err error, v any) error {
	if err != nil {
		return fmt.Errorf("%w (%T)", err, v)
	}
	return nil
}

//----------

func bypassPointersAndInterfaces(val reflect.Value) reflect.Value {
	for (val.Kind() == reflect.Pointer || val.Kind() == reflect.Interface) && !val.IsNil() {
		val = val.Elem()
	}
	return val
}
func bypassPointerTypes(typ reflect.Type) reflect.Type {
	for typ != nil && typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	return typ
}
