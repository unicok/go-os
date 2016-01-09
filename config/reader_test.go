package config

import (
	"encoding/json"
	"testing"
)

func TestReader(t *testing.T) {
	source1 := map[string]interface{}{
		"a": "b",
		"c": map[string]interface{}{
			"d": "e",
		},
		"d": "a",
		"e": 1,
		"f": 1.02,
		"g": true,
	}

	source2 := map[string]interface{}{
		"a": "c",
		"c": map[string]interface{}{
			"d": "f",
			"g": 1,
		},
	}

	s1, err := json.Marshal(source1)
	if err != nil {
		t.Error(err)
	}
	s2, err := json.Marshal(source2)
	if err != nil {
		t.Error(err)
	}

	r := NewReader()

	// Try 1 changeset
	v, err := r.Values(&ChangeSet{
		Data: s1,
	})
	if err != nil {
		t.Error(err)
	}

	if res := v.Get("a").String(""); res != "b" {
		t.Errorf("Expected %s got %s", "a", res)
	}

	if res := v.Get("c", "d").String(""); res != "e" {
		t.Errorf("Expected %s got %s", "e", res)
	}

	if res := v.Get("g").Bool(false); res != true {
		t.Errorf("Expected %t got %t", true, res)
	}

	// Try merged ChangeSet
	ch, err := r.Parse(&ChangeSet{Data: s1}, &ChangeSet{Data: s2})
	if err != nil {
		t.Error(err)
	}

	v, err = r.Values(ch)
	if err != nil {
		t.Error(err)
	}

	if res := v.Get("e").Int(0); res != 1 {
		t.Errorf("Expected %d got %d", 1, res)
	}

	if res := v.Get("c", "d").String(""); res != "f" {
		t.Errorf("Expected %s got %s", "f", res)
	}
}
