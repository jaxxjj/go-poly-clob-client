package cache

import (
	"testing"
	"time"
)

func TestTTL_SetAndGet(t *testing.T) {
	c := New[string](5 * time.Minute)

	c.Set("key1", "value1")
	c.Set("key2", "value2")

	v, ok := c.Get("key1")
	if !ok || v != "value1" {
		t.Errorf("Get(key1) = %q, %v; want value1, true", v, ok)
	}

	v, ok = c.Get("key2")
	if !ok || v != "value2" {
		t.Errorf("Get(key2) = %q, %v; want value2, true", v, ok)
	}
}

func TestTTL_GetMissing(t *testing.T) {
	c := New[int](5 * time.Minute)

	v, ok := c.Get("missing")
	if ok || v != 0 {
		t.Errorf("Get(missing) = %d, %v; want 0, false", v, ok)
	}
}

func TestTTL_Expiry(t *testing.T) {
	c := New[string](10 * time.Millisecond)

	c.Set("key", "value")

	v, ok := c.Get("key")
	if !ok || v != "value" {
		t.Fatalf("Get(key) before expiry = %q, %v; want value, true", v, ok)
	}

	time.Sleep(20 * time.Millisecond)

	_, ok = c.Get("key")
	if ok {
		t.Error("Get(key) after expiry should return false")
	}
}

func TestTTL_ClearSpecific(t *testing.T) {
	c := New[string](5 * time.Minute)

	c.Set("a", "1")
	c.Set("b", "2")
	c.Set("c", "3")

	c.Clear("a", "c")

	if _, ok := c.Get("a"); ok {
		t.Error("a should be cleared")
	}
	if v, ok := c.Get("b"); !ok || v != "2" {
		t.Error("b should still exist")
	}
	if _, ok := c.Get("c"); ok {
		t.Error("c should be cleared")
	}
}

func TestTTL_ClearAll(t *testing.T) {
	c := New[int](5 * time.Minute)

	c.Set("x", 10)
	c.Set("y", 20)

	c.Clear()

	if _, ok := c.Get("x"); ok {
		t.Error("x should be cleared")
	}
	if _, ok := c.Get("y"); ok {
		t.Error("y should be cleared")
	}
}

func TestTTL_Overwrite(t *testing.T) {
	c := New[string](5 * time.Minute)

	c.Set("key", "old")
	c.Set("key", "new")

	v, ok := c.Get("key")
	if !ok || v != "new" {
		t.Errorf("Get(key) = %q, %v; want new, true", v, ok)
	}
}

func TestTTL_BoolType(t *testing.T) {
	c := New[bool](5 * time.Minute)

	c.Set("neg_risk", true)

	v, ok := c.Get("neg_risk")
	if !ok || !v {
		t.Errorf("Get(neg_risk) = %v, %v; want true, true", v, ok)
	}

	// false value should be distinguishable from missing
	c.Set("normal", false)
	v, ok = c.Get("normal")
	if !ok || v {
		t.Errorf("Get(normal) = %v, %v; want false, true", v, ok)
	}
}
