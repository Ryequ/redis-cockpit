package result

import (
	"strings"
	"testing"
)

func TestNormalizeConvertsScalarAndCollectionReplies(t *testing.T) {
	str := Normalize("hello")
	if str.Kind != KindString || str.Display() != "hello" {
		t.Fatalf("string value = %#v", str)
	}

	num := Normalize(int64(42))
	if num.Kind != KindInt || num.Display() != "42" {
		t.Fatalf("int value = %#v", num)
	}

	arr := Normalize([]any{"a", int64(2)})
	if arr.Kind != KindArray || len(arr.Items) != 2 || arr.Items[1].Display() != "2" {
		t.Fatalf("array value = %#v", arr)
	}

	m := Normalize(map[any]any{"name": "redis"})
	if m.Kind != KindMap || len(m.Pairs) != 1 || m.Pairs[0].Value.Display() != "redis" {
		t.Fatalf("map value = %#v", m)
	}
}

func TestRenderFormats(t *testing.T) {
	value := Normalize([]any{"OK", int64(1)})

	if got := Render(value, FormatRaw); !strings.Contains(got, "OK") {
		t.Fatalf("raw render = %q", got)
	}
	if got := Render(value, FormatJSON); !strings.Contains(got, `"OK"`) {
		t.Fatalf("json render = %q", got)
	}
	if got := Render(value, FormatTable); !strings.Contains(got, "VALUE") {
		t.Fatalf("table render = %q", got)
	}
}

func TestPaginateReturnsRequestedPage(t *testing.T) {
	items := []Value{
		Normalize("one"), Normalize("two"), Normalize("three"), Normalize("four"),
	}

	page, hasMore := Paginate(items, 0, 2)
	if len(page) != 2 || page[0].Display() != "one" || !hasMore {
		t.Fatalf("first page = %#v, hasMore = %v", page, hasMore)
	}

	page, hasMore = Paginate(items, 1, 2)
	if len(page) != 2 || page[0].Display() != "three" || hasMore {
		t.Fatalf("last page = %#v, hasMore = %v", page, hasMore)
	}
}
