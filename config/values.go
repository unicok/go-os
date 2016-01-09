package config

import (
	"encoding/json"
	"fmt"
	"time"

	simple "github.com/bitly/go-simplejson"
)

type jsonValues struct {
	ch *ChangeSet
	sj *simple.Json
}

type jsonValue struct {
	*simple.Json
}

func newValues(ch *ChangeSet) (Values, error) {
	sj := new(simple.Json)
	err := sj.UnmarshalJSON(ch.Data)
	if err != nil {
		return nil, err
	}
	return &jsonValues{ch, sj}, nil
}

func (j *jsonValues) Get(path ...string) Value {
	return &jsonValue{j.sj.GetPath(path...)}
}

func (j *jsonValue) Bool(def bool) bool {
	return j.Json.MustBool(def)
}

func (j *jsonValue) Int(def int) int {
	return j.Json.MustInt(def)
}

func (j *jsonValue) String(def string) string {
	return j.Json.MustString(def)
}

func (j *jsonValue) Float64(def float64) float64 {
	return j.Json.MustFloat64(def)
}

func (j *jsonValue) Duration(def time.Duration) time.Duration {
	v, err := j.Json.String()
	if err != nil {
		return def
	}

	value, err := time.ParseDuration(v)
	if err != nil {
		return def
	}

	return value
}

func (j *jsonValue) StringSlice(def []string) []string {
	return j.Json.MustStringArray(def)
}

func (j *jsonValue) StringMap(def map[string]string) map[string]string {
	m, err := j.Json.Map()
	if err != nil {
		return def
	}

	var res map[string]string
	for k, v := range m {
		res[k] = fmt.Sprintf("%v", v)
	}

	return res
}

func (j *jsonValue) Scan(v interface{}) error {
	b, err := j.Json.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func (j *jsonValue) Bytes() []byte {
	b, err := j.Json.Bytes()
	if err != nil {
		return []byte{}
	}
	return b
}
