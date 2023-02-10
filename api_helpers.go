package coinbasetrade

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"time"

	"github.com/shopspring/decimal"
)

// timeToString converts a time object into the string format used by the API
func timeToString(t time.Time) string {
	return t.Format(time.RFC3339)
}

// parametersToValues takes a pointer to a Parameters struct, and converts the values
// to url.Values which can be used in GET queries. Any field in the struct with a `cbt`
// tag will have that tag used as the key in the url.Values.
func parametersToValues(p interface{}) (u url.Values, err error) {
	if p == nil {
		err = errors.New("nil value passed to parametersToValues")
		return
	}

	u = make(url.Values)

	params := reflect.ValueOf(p)

	if params.Kind() != reflect.Struct {
		err = errors.New("struct not provided as source of reflect")
		return
	}

	for i := 0; i < params.NumField(); i++ {
		val := params.Field(i)
		tag := params.Type().Field(i).Tag.Get("cbt")
		if tag != "" {
			switch val.Type().Kind() {

			// strings
			case reflect.String:
				if s := val.String(); s != "" {
					u.Add(tag, s)
				}

				// ints
			case reflect.Int, reflect.Int64:
				if i := val.Int(); i != 0 {
					u.Add(tag, fmt.Sprintf("%d", i))
				}
				// slice of strings: add each separately
			case reflect.Slice:
				if val.Len() > 0 {
					for i := 0; i < val.Len(); i++ {
						u.Add(tag, val.Index(i).String())
					}
				}

				// structs
			case reflect.Struct:
				switch val.Type() {

				// time.Time
				case reflect.TypeOf(time.Time{}):
					if t := val.Interface().(time.Time); !t.IsZero() {
						u.Add(tag, timeToString(t))
					}

					// decimal.Decimal
				case reflect.TypeOf(decimal.Decimal{}):
					if d := val.Interface().(decimal.Decimal); !d.IsZero() {
						u.Add(tag, d.String())
					}
				}
			}
		}
	}

	return
}
