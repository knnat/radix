package radix_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	. "github.com/knnat/radix"
)

type testWrapper struct {
	label    string
	priority int
	depth    int
	value    interface{}
}

func TestEscape(t *testing.T){
	tr := New()

	err := tr.Add("/abc@abc@/", 0)
	if assert.Error(t, err){
		assert.Equal(t, err, ErrInvalid)
	}

	assert.Nil(t, tr.Add("/@abc", 0))
	assert.Error(t, tr.Add("/@ab", 0))
	assert.Error(t, tr.Add("/@abcd", 0))

	assert.Nil(t, tr.Add("/@abc", 2))
}

func TestTree(t *testing.T) {
	testCases := []struct {
		labels      []string
		wrappers    []testWrapper
		length      int
		size        int
		params      map[string]string
		placeholder byte
		delim       byte
	}{
		{
			labels: []string{"foobar"},
			wrappers: []testWrapper{
				{label: "foobar", priority: 1, depth: 1, value: "bazqux"},
			},
			length: 2,
			size:   len("bazqux"),
		},
		{
			labels: []string{"a", "b"},
			wrappers: []testWrapper{
				{label: "a", priority: 1, depth: 1, value: 1},
				{label: "b", priority: 1, depth: 1, value: 2},
			},
			length: 3,
			size:   len("a") + len("b"),
		},
		{
			labels: []string{"a", "ab", "abc"},
			wrappers: []testWrapper{
				{label: "a", priority: 3, depth: 1, value: 1},
				{label: "ab", priority: 2, depth: 2, value: 2},
				{label: "abc", priority: 1, depth: 3, value: 3},
			},
			length: 4,
			size:   len("abc"),
		},
		{
			labels: []string{"a", "ab", "abc", "d"},
			wrappers: []testWrapper{
				{label: "a", priority: 3, depth: 1, value: 1},
				{label: "ab", priority: 2, depth: 2, value: 2},
				{label: "abc", priority: 1, depth: 3, value: 3},
				{label: "d", priority: 1, depth: 1, value: 4},
			},
			length: 5,
			size:   len("abc") + len("d"),
		},
		{
			labels: []string{"ab", "a", "abc"},
			wrappers: []testWrapper{
				{label: "ab", priority: 2, depth: 2, value: 2},
				{label: "a", priority: 3, depth: 1, value: 1},
				{label: "abc", priority: 1, depth: 3, value: 3},
			},
			length: 4,
			size:   len("abc"),
		},
		{
			labels: []string{"ab", "abc", "a"},
			wrappers: []testWrapper{
				{label: "ab", priority: 2, depth: 2, value: 2},
				{label: "abc", priority: 1, depth: 3, value: 3},
				{label: "a", priority: 3, depth: 1, value: 1},
			},
			length: 4,
			size:   len("abc"),
		},
		{
			labels: []string{"abc", "a", "ab"},
			wrappers: []testWrapper{
				{label: "abc", priority: 1, depth: 3, value: 3},
				{label: "a", priority: 3, depth: 1, value: 1},
				{label: "ab", priority: 2, depth: 2, value: 2},
			},
			length: 4,
			size:   len("abc"),
		},
		{
			labels: []string{"a", "b", "c"},
			wrappers: []testWrapper{
				{label: "a", priority: 1, depth: 1, value: 1},
				{label: "b", priority: 1, depth: 1, value: 2},
				{label: "c", priority: 1, depth: 1, value: 3},
			},
			length: 4,
			size:   len("a") + len("b") + len("c"),
		},
		{
			labels: []string{"/path/123"},
			wrappers: []testWrapper{
				{label: "/path/@id", priority: 1, depth: 1, value: "foobar"},
			},
			length:      2,
			size:        len("/path/@id"),
			params:      map[string]string{"id": "123"},
			placeholder: '@',
			delim:       '/',
		},
		{
			labels: []string{"/path/123/subpath/456"},
			wrappers: []testWrapper{
				{label: "/path/@id/subpath/@id2", priority: 1, depth: 1, value: "foobar"},
			},
			length:      2,
			size:        len("/path/@id/subpath/@id2"),
			params:      map[string]string{"id": "123", "id2": "456"},
			placeholder: '@',
			delim:       '/',
		},
		{
			labels: []string{"/path/123", "/path/123/subpath/456"},
			wrappers: []testWrapper{
				{label: "/path/@id", priority: 2, depth: 1, value: "foobar"},
				testWrapper{
					label:    "/path/@id/subpath/@id2",
					priority: 1,
					depth:    2,
					value:    "bazqux",
				},
			},
			length:      3,
			size:        len("/path/@id/subpath/@id2"),
			params:      map[string]string{"id": "123", "id2": "456"},
			placeholder: '@',
			delim:       '/',
		},
		{
			labels: []string{"/api/user/123"},
			wrappers: []testWrapper{
				{label: "/api/user/@id", priority: 1, depth: 1, value: "foobar"},
			},
			length:      2,
			size:        len("/api/user/@id"),
			params:      map[string]string{"id": "123"},
			placeholder: '@',
			delim:       '/',
		},
	}
	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			tr := New()
			// if tc.placeholder != 0 && tc.delim != 0 {
			// 	tr.SetBoundaries(tc.placeholder, tc.delim)
			// }
			for _, w := range tc.wrappers {
				tr.Add(w.label, w.value)
			}
			t.Log(tr.String())

			if want, got := tc.length, tr.Len(); want != got {
				t.Errorf("want %d, got %d", want, got)
			}
			if want, got := tc.size, tr.Size(); want != got {
				t.Errorf("want %d, got %d", want, got)
			}
			var (
				n *Node
				p map[string]string
			)
			for i, w := range tc.wrappers {
				n, p = tr.Get(tc.labels[i])
				if want, got := w.value, n.Value; !reflect.DeepEqual(want, got) {
					t.Errorf("want %v, got %v", want, got)
				}
				// if want, got := w.priority, n.Priority(); want != got {
				// 	t.Errorf("want %d, got %d", want, got)
				// }
				if want, got := w.depth, n.Depth(); want != got {
					t.Errorf("want %d, got %d", want, got)
				}
			}
			if want, got := tc.params, p; !reflect.DeepEqual(want, got) {
				t.Errorf("want %v, got %v", want, got)
			}

			for i, w := range tc.wrappers {
				tr.Del(w.label)
				n, _ = tr.Get(tc.labels[i])
				if want, got := (*Node)(nil), n; want != got {
					t.Errorf("want %v, got %v", want, got)
				}
			}
		})
	}
}
