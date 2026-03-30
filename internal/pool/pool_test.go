package pool

import "testing"

type testObject struct {
	value int
	used  bool
}

func (o *testObject) Reset() {
	o.value = 0
	o.used = false
}

func TestPoolGetNew(t *testing.T) {
	called := 0
	p := New(func() *testObject {
		called++
		return &testObject{value: 42}
	})

	obj := p.Get()
	if called != 1 {
		t.Errorf("expected newFunc to be called once, got %d", called)
	}
	if obj.value != 42 {
		t.Errorf("expected value 42, got %d", obj.value)
	}
}

func TestPoolGetReused(t *testing.T) {
	p := New(func() *testObject {
		return &testObject{value: 42}
	})

	obj1 := p.Get()
	obj1.value = 100
	obj1.used = true

	p.Put(obj1)

	obj2 := p.Get()
	if obj2 != obj1 {
		t.Errorf("expected same object, got different")
	}
	if obj2.value != 0 {
		t.Errorf("expected value 0 after reset, got %d", obj2.value)
	}
	if obj2.used != false {
		t.Errorf("expected used false after reset, got %v", obj2.used)
	}
}

func TestPoolMultipleObjects(t *testing.T) {
	p := New(func() *testObject {
		return &testObject{}
	})

	obj1 := p.Get()
	obj1.value = 1
	p.Put(obj1)

	obj2 := p.Get()
	obj2.value = 2
	p.Put(obj2)

	obj3 := p.Get()
	if obj3.value != 0 {
		t.Errorf("expected value 0 after reset, got %d", obj3.value)
	}

	p.Put(obj3)

	obj4 := p.Get()
	if obj4.value != 0 {
		t.Errorf("expected value 0 after reset, got %d", obj4.value)
	}
}
