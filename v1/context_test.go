package v1

import (
	"reflect"
	"testing"
)

type lookupper interface {
	lookup(string) (reflect.Value, bool)
}

func checkLookup(t *testing.T, l lookupper, key string, value interface{}, ok bool) {
	//a := l.(*lookupper)
	v, present := l.lookup(key)
	if present != ok {
		t.Errorf("Expected %v presence, got %v\n", present, ok)
		return
	}
	if ok && v.Interface() != value {
		t.Errorf("Expected %v, got %v\n", value, v)
	}
}

func TestStructContext(t *testing.T) {
	x := struct {
		One, Two, Three int
	}{1, 2, 3}

	c, err := NewContext(x)
	if err != nil {
		t.Error(err)
	}

	for i, key := range []string{"One", "Two", "Three"} {
		checkLookup(t, c, key, i+1, true)
	}
	checkLookup(t, c, "four", nil, false)

	// test sending a pointer to a struct
	c, err = NewContext(&x)
	if err != nil {
		t.Error(err)
	}

	for i, key := range []string{"One", "Two", "Three"} {
		checkLookup(t, c, key, i+1, true)
	}
	checkLookup(t, c, "four", nil, false)
}

func TestStructMulti(t *testing.T) {
	x := struct {
		Name, Age string
	}{"Jason", "32"}
	ctx := make(contextStack, 0, 5)
	c, err := NewContext(x)
	if err != nil {
		t.Fatal(err)
	}
	ctx.push(c)

	checkLookup(t, c, "Name", "Jason", true)
	checkLookup(t, c, "Age", "32", true)
	checkLookup(t, ctx, "Name", "Jason", true)
	checkLookup(t, ctx, "Age", "32", true)
	checkLookup(t, ctx, "Foo", nil, false)

	c, err = NewContext(struct {
		Foo int
	}{1})
	if err != nil {
		t.Fatal(err)
	}

	ctx.push(c)

	checkLookup(t, ctx, "Name", "Jason", true)
	checkLookup(t, ctx, "Age", "32", true)
	checkLookup(t, ctx, "Foo", 1, true)
}

func TestMapContext(t *testing.T) {
	x := map[string]int{"one": 1, "two": 2, "three": 3}
	c, err := NewContext(x)
	if err != nil {
		t.Error(err)
	}
	for key, val := range x {
		checkLookup(t, c, key, val, true)
	}

	checkLookup(t, c, "four", nil, false)
}

func TestMapMulti(t *testing.T) {
	ctx := make(contextStack, 0, 5)
	c, err := NewContext(map[string]string{"name": "Jason", "Age": "32"})
	if err != nil {
		t.Fatal(err)
	}
	ctx.push(c)

	checkLookup(t, c, "name", "Jason", true)
	checkLookup(t, c, "Age", "32", true)
	checkLookup(t, ctx, "name", "Jason", true)
	checkLookup(t, ctx, "Age", "32", true)
	checkLookup(t, ctx, "Foo", nil, false)

	c, err = NewContext(map[string]int{"Foo": 1})
	if err != nil {
		t.Fatal(err)
	}
	ctx.push(c)

	checkLookup(t, ctx, "name", "Jason", true)
	checkLookup(t, ctx, "Age", "32", true)
	checkLookup(t, ctx, "Foo", 1, true)
}

func TestMixedMulti(t *testing.T) {
	ctx := make(contextStack, 0, 5)
	c, err := NewContext(map[string]string{"name": "Jason", "Age": "32"})
	if err != nil {
		t.Fatal(err)
	}
	ctx.push(c)

	checkLookup(t, c, "name", "Jason", true)
	checkLookup(t, c, "Age", "32", true)
	checkLookup(t, ctx, "name", "Jason", true)
	checkLookup(t, ctx, "Age", "32", true)
	checkLookup(t, ctx, "Foo", nil, false)

	c, err = NewContext(struct {
		Foo int
	}{1})
	if err != nil {
		t.Fatal(err)
	}

	ctx.push(c)

	checkLookup(t, ctx, "name", "Jason", true)
	checkLookup(t, ctx, "Age", "32", true)
	checkLookup(t, ctx, "Foo", 1, true)
}
