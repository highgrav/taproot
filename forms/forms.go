package forms

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

var ErrTypeIsNotStruct = errors.New("generic type must be a struct")

func FromQueryString[T any](r *http.Request) (T, error) {
	var query map[string][]string = make(map[string][]string)
	for k, v := range r.URL.Query() {
		query[k] = v
	}
	return FromMap[T](query)
}

func FromForm[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	mimeType := r.Header.Get("Content-Type")
	if mimeType == "" {
		var t T
		return t, errors.New("empty mime type")
	}
	if mimeType == "application/x-www-form-urlencoded" {
		return FromUrlEncodedForm[T](r)
	} else if strings.HasPrefix(mimeType, "multipart/form-data") {
		return FromMultipartForm[T](w, r)
	} else if mimeType == "application/json" {
		return FromJsonBody[T](r)
	}
	var t T
	return t, errors.New("unknown mime type " + mimeType)
}

func FromUrlEncodedForm[T any](r *http.Request) (T, error) {
	err := r.ParseForm()
	if err != nil {
		var t T
		return t, err
	}
	vals := make(map[string][]string)
	for k, v := range r.Form {
		vals[strings.ToLower(k)] = v
	}
	return FromMap[T](vals)
}

func FromMultipartForm[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		var t T
		return t, err
	}
	vals := make(map[string][]string)
	for k, v := range r.Form {
		vals[strings.ToLower(k)] = v
	}
	return FromMap[T](vals)
}

func FromJsonBody[T any](r *http.Request) (T, error) {
	var t T
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		return t, err
	}
	return t, nil
}

func FromMap[T any](vals map[string][]string) (T, error) {
	var t T
	if reflect.ValueOf(t).Kind() != reflect.Struct {
		return t, ErrTypeIsNotStruct
	}
	tTyp := reflect.TypeOf(&t)
	tVal := reflect.ValueOf(&t)
	for x := 0; x < tTyp.Elem().NumField(); x++ {
		field := tTyp.Elem().Field(x)
		jt := strings.Split(field.Tag.Get("json"), ",")[0]
		key := strings.ToLower(jt)
		if key == "" {
			key = strings.ToLower(field.Name)
		}
		formVal := []string{}
		if fv, ok := vals[key]; ok {
			formVal = fv
		} else {
			for k, _ := range vals {
				if strings.ToLower(k) == key {
					formVal = vals[k]
				}
			}
		}

		kind := field.Type.Kind()
		res := tVal.Elem().Field(x)

		if len(formVal) == 0 {
			continue
		}

		switch kind {
		case reflect.String:
			res.SetString(formVal[0])
		case reflect.Uint:
		case reflect.Uint16:
		case reflect.Uint8:
		case reflect.Uint32:
		case reflect.Uint64:
			resui, err := strconv.ParseUint(formVal[0], 10, 64)
			if err != nil {
				return t, errors.New("invalid attempt to parse uint from form field " + key + ", value " + formVal[0])
			}
			res.SetUint(resui)
		case reflect.Int:
		case reflect.Int8:
		case reflect.Int16:
		case reflect.Int32:
		case reflect.Int64:
			resi, err := strconv.ParseInt(formVal[0], 10, 64)
			if err != nil {
				return t, errors.New("invalid attempt to parse int from form field " + key + ", value " + formVal[0])
			}
			res.SetInt(resi)
		case reflect.Bool:
			if strings.ToLower(formVal[0]) == "true" {
				res.SetBool(true)
			} else if strings.ToLower(formVal[0]) == "false" {
				res.SetBool(false)
			}
		case reflect.Float32:
		case reflect.Float64:
			resfloat, err := strconv.ParseFloat(formVal[0], 64)
			if err != nil {
				return t, errors.New("invalid attempt to parse float from form field " + key + ", value " + formVal[0])
			}
			res.SetFloat(resfloat)
		case reflect.Array:
		case reflect.Slice:
			// TODO
		}
	}
	return t, nil
}

// sliceToPopulate needs to be a pointer to the slice or array property
func populateSliceProperty(sliceToPopulate any, value any) error {
	sliceValue := reflect.ValueOf(sliceToPopulate)
	if sliceValue.Kind() != reflect.Ptr {
		return errors.New("sliceToPopulate must be a pointer to a slice or array")
	}
	sliceValue = sliceValue.Elem()
	if sliceValue.Kind() != reflect.Slice && sliceValue.Kind() != reflect.Array {
		return errors.New("sliceToPopulate must be a pointer to a slice or array")
	}
	elemType := sliceValue.Type().Elem()
	valueType := reflect.ValueOf(value)
	if valueType.Kind() != reflect.Slice && valueType.Kind() != reflect.Array {
		return errors.New("value must be a slice or array")
	}
	if elemType.Kind() != valueType.Type().Elem().Kind() {
		return errors.New("value type is incommensurate with sliceToPopulate type")
	}

	for x := 0; x < valueType.Len(); x++ {
		elemValue := valueType.Index(x)
		if !elemValue.Type().ConvertibleTo(elemType) {
			return errors.New("value type at position " + strconv.Itoa(x) + " is incommensurate with sliceToPopulate type")
		}
		sliceValue = reflect.Append(sliceValue, elemValue.Convert(elemType))
	}
	reflect.ValueOf(sliceToPopulate).Elem().Set(sliceValue)
	return nil

}
