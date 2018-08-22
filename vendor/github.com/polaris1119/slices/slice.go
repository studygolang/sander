package slices

import "reflect"

// StructInt64Slice returns a slice of int64. For more info refer to Slice types StructInt64Slice() method.
func StructsInt64Slice(s interface{}, fieldName string) []int64 {
	return New(s).StructInt64Slice(fieldName)
}

// StructsIntSlice returns a slice of int. For more info refer to Slice types StructIntSlice() method.
func StructsIntSlice(s interface{}, fieldName string) []int {
	return New(s).StructIntSlice(fieldName)
}

type Slice struct {
	value reflect.Value
}

// New returns a new *Slice with the slice s. It panics if the s's kind is
// not slice.
func New(s interface{}) *Slice {
	return &Slice{
		value: sliceVal(s),
	}
}

// StructIntSlice extracts the given s slice's every element, which is struct, to []int by the field.
// It panics if the s's element is not struct, or field is not exits, or the value of field is not signed integer.
func (this *Slice) StructIntSlice(fieldName string) []int {
	length := this.value.Len()
	intSlice := make([]int, length)

	for i := 0; i < length; i++ {
		v := this.strucFieldVal(i, fieldName)
		switch v.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int, reflect.Int32, reflect.Int64:
			intSlice[i] = int(v.Int())
		default:
			panic("polaris1119/slices: the value of field is not signed integer.")
		}
	}

	return intSlice
}

// StructInt64Slice extracts the given s slice's every element, which is struct, to []int64 by the field.
// It panics if the s's element is not struct, or field is not exits, or the value of field is not signed integer.
func (this *Slice) StructInt64Slice(fieldName string) []int64 {
	length := this.value.Len()
	intSlice := make([]int64, length)

	for i := 0; i < length; i++ {
		v := this.strucFieldVal(i, fieldName)
		switch v.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int, reflect.Int32, reflect.Int64:
			intSlice[i] = v.Int()
		default:
			panic("polaris1119/slices: the value of field is not signed integer.")
		}
	}

	return intSlice
}

func (this *Slice) strucFieldVal(i int, fieldName string) reflect.Value {
	val := this.value.Index(i)
	if !this.isStruct(val) {
		panic("polaris1119/slices: the slice's element is not struct or pointer of struct!")
	}

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	v := val.FieldByName(fieldName)
	if !v.IsValid() {
		panic("polaris1119/slices: the struct of slice's element has not the field:" + fieldName)
	}

	return v
}

func (this *Slice) isStruct(v reflect.Value) bool {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// uninitialized zero value of a struct
	if v.Kind() == reflect.Invalid {
		return false
	}

	return v.Kind() == reflect.Struct
}

// Name returns the slice's type name within its package. For more info refer
// to Name() function.
func (this *Slice) Name() string {
	return this.value.Type().Name()
}

func sliceVal(s interface{}) reflect.Value {
	v := reflect.ValueOf(s)

	// if pointer get the underlying element≤
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Slice {
		panic("not slice")
	}

	return v
}
