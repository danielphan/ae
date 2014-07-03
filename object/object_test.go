package object

import (
	"appengine"
	"appengine/aetest"
	"appengine/datastore"
	"fmt"
	"testing"
	"time"
)

const testObjectKind = "testObject"

type testObject struct {
	Object
	Foo string
	Bar string
}

func TestNewSaveGet(t *testing.T) {
	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	t.Log("Save the object.")
	out := testObject{
		Object: New(testObjectKind, ""),
		Foo:    "foo",
		Bar:    "bar",
	}
	err = Save(c, &out)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Get the saved object.")
	in := testObject{
		Object: New(testObjectKind, out.ID),
	}
	err = Get(c, &in)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("time.Time is stored inexactly. Round to milliseconds.")
	for _, at := range []*time.Time{
		&in.CreatedAt, &in.ModifiedAt,
		&out.CreatedAt, &out.ModifiedAt,
	} {
		*at = at.Round(time.Millisecond)
	}

	t.Log("See if the objects match.")
	if in != out {
		t.Fatalf("saved %#v but got %#v", out, in)
	}
}

func TestVersionModifiedAt(t *testing.T) {
	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	t.Log("Save the object to compute Version and ModifiedAt.")
	to1 := testObject{
		Object: New(testObjectKind, ""),
		Foo:    "foo",
		Bar:    "bar",
	}
	err = Save(c, &to1)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Copy the object, modify and save the copy.")
	to2 := to1
	to2.Foo = "FOO"
	err = Save(c, &to2)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("See if the Versions are different.")
	if to1.Version == to2.Version {
		t.Errorf("%#v and %#v have the same Version", to1, to2)
	}

	t.Log("See if the modified copy has a later ModifiedAt.")
	if !to1.ModifiedAt.Before(to2.ModifiedAt) {
		t.Errorf("%#v ModifiedAt earlier than %#v ModifiedAt", to1, to2)
	}
}

func TestUnchanged(t *testing.T) {
	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	t.Log("Save the object to compute Version and ModifiedAt.")
	to := testObject{
		Object: New(testObjectKind, ""),
		Foo:    "foo",
		Bar:    "bar",
	}
	err = Save(c, &to)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Save it again without changing it.")
	err = Save(c, &to)
	if err != ErrNotChanged {
		t.Fatal("Expected to get an ErrNotChanged.")
	}
}

func TestGroup(t *testing.T) {
	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	g := Entity{Kind: "testGroup", ID: "tg"}

	o := New(testObjectKind, "")
	o.Group = g
	to1 := testObject{
		Object: o,
		Foo:    "foo",
		Bar:    "bar",
	}
	err = datastore.RunInTransaction(c, func(c appengine.Context) error {
		t.Log("Save the object.")
		err = Save(c, &to1)
		if err != nil {
			return err
		}
		return nil
	}, nil)

	t.Log("Get the object back.")
	o.ID = to1.ID
	to2 := testObject{
		Object: o,
	}
	err = Get(c, &to2)
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkModified(b *testing.B) {
	c, err := aetest.NewContext(nil)
	if err != nil {
		b.Fatal(err)
	}
	defer c.Close()

	b.Log("Create a new object.")
	to := &testObject{
		Object: New(testObjectKind, ""),
		Foo:    "foo",
		Bar:    "bar",
	}
	_, err = modified(c, to)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.Log("Modify the object.")
		to.Foo = fmt.Sprintf("foo%d", i)
		to.Bar = fmt.Sprintf("bar%d", i)

		b.Log("Update Version and ModifiedAt.")
		_, err := modified(c, to)
		if err != nil {
			b.Fatal(err)
		}
	}
}
