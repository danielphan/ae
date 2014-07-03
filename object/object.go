package object

import (
	"appengine"
	"appengine/datastore"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"time"
)

var ErrNotChanged = errors.New("object: Object not changed")

type Entity struct {
	Kind string
	ID string
}

func (e Entity) Key(c appengine.Context) *datastore.Key {
	if e != (Entity{}) {
		return datastore.NewKey(c, e.Kind, e.ID, 0, nil)
	}
	return nil
}

type Object struct {
	Entity
	Group Entity
	// Version is the version number for this Object.
	// It is the hash of the Object's exported fields,
	// calculated at the last time Modified was called.
	Version    string
	ModifiedAt time.Time
	CreatedAt  time.Time
}

func New(kind string, id string) Object {
	return Object{
		Entity: Entity{
			Kind: kind,
			ID: id,
		},
	}
}

func (o *Object) Key(c appengine.Context) *datastore.Key {
	groupKey := o.Group.Key(c)
	return datastore.NewKey(c, o.Kind, o.ID, 0, groupKey)
}

type objecter interface {
	object() *Object
}

func (o *Object) object() *Object {
	return o
}

func Get(c appengine.Context, dst objecter) error {
	o := dst.object()
	c.Debugf("object: Getting object %q.", o.ID)

	err := datastore.Get(c, o.Key(c), dst)
	if err != nil {
		return err
	}
	return nil
}

func Save(c appengine.Context, src objecter) error {
	o := src.object()
	c.Debugf("object: Saving object %q.", o.ID)

	changed, err := modified(c, src)
	if err != nil {
		return err
	}
	if !changed {
		return ErrNotChanged
	}

	c.Debugf("object: Storing modified object to datastore.")
	_, err = datastore.Put(c, o.Key(c), src)
	if err != nil {
		return err
	}

	return nil
}

func modified(c appengine.Context, src objecter) (changed bool, err error) {
	c.Debugf("object: Checking object for modifications.")
	o := src.object()

	if o.ID == "" {
		c.Debugf("object: Creating new ID.")
		i, _, err := datastore.AllocateIDs(c, o.Kind, nil, 1)
		if err != nil {
			return false, err
		}
		o.ID = strconv.FormatInt(i, 16)
	}

	if o.CreatedAt.IsZero() {
		o.CreatedAt = time.Now()
	}

	// Ignore Version and ModifiedAt when computing the hash.
	// Reinstate them if something fails.
	oldVersion := o.Version
	o.Version = ""

	oldModifiedAt := o.ModifiedAt
	o.ModifiedAt = time.Time{}

	c.Debugf("object: Serializing object for hashing.")
	// While SHA-256 is slow, xxhash32 is too short,
	// so collisions would be probable.
	// A 160-bit hash would work.
	h := sha256.New()
	b, err := json.Marshal(src)
	_, err = io.Copy(h, bytes.NewReader(b))
	if err != nil {
		o.Version = oldVersion
		o.ModifiedAt = oldModifiedAt
		return false, err
	}

	c.Debugf("object: Computing hash.")
	v := hex.EncodeToString(h.Sum(nil))
	if v != oldVersion {
		o.Version = v
		o.ModifiedAt = time.Now()
		return true, nil
	}

	o.Version = oldVersion
	o.ModifiedAt = oldModifiedAt
	return false, err
}
