package tiffifd

import "fmt"

// Scenario 是一个固定场景。
type Scenario struct {
	Name string
	Run  func() map[string]any
}

type entry struct {
	tag, typ, count, value int
}

// BuildIFD 拼一份小端 TIFF，一个 IFD，值都内联。
func BuildIFD(entries []entry) []byte {
	out := []byte{'I', 'I', 42, 0, 8, 0, 0, 0}
	body := []byte{byte(len(entries)), byte(len(entries) >> 8)}
	for _, e := range entries {
		body = append(body, byte(e.tag), byte(e.tag>>8), byte(e.typ), byte(e.typ>>8))
		body = append(body, byte(e.count), byte(e.count>>8), 0, 0)
		body = append(body, byte(e.value), byte(e.value>>8), byte(e.value>>16), byte(e.value>>24))
	}
	body = append(body, 0, 0, 0, 0)
	return append(out, body...)
}

// Samples 返回全部场景。
func Samples() []Scenario {
	return []Scenario{
		{Name: "endian", Run: func() map[string]any {
			out := []byte{'M', 'M', 0, 42, 0, 0, 0, 8}
			body := []byte{1, 0}
			body = append(body, 0x01, 0x00, 0x00, 0x03, 0, 0, 0, 1, 0, 100, 0, 0)
			body = append(body, 0, 0, 0, 0)
			file, _ := Parse(append(out, body...), &Counter{})
			if file == nil {
				return map[string]any{"width": 0}
			}
			values := file.Get(0, 0x0100, &Counter{})
			if len(values) == 0 {
				return map[string]any{"width": 0}
			}
			return map[string]any{"width": values[0]}
		}},
		{Name: "inline", Run: func() map[string]any {
			data := BuildIFD([]entry{{0x0100, 3, 1, 100}})
			file, _ := Parse(data, &Counter{})
			values := file.Get(0, 0x0100, &Counter{})
			if len(values) == 0 {
				return map[string]any{"width": 0}
			}
			return map[string]any{"width": values[0]}
		}},
		{Name: "count2", Run: func() map[string]any {
			data := BuildIFD([]entry{{0x0100, 3, 2, 100 | 200<<16}})
			file, _ := Parse(data, &Counter{})
			values := file.Get(0, 0x0100, &Counter{})
			return map[string]any{"count": len(values)}
		}},
		{Name: "rational", Run: func() map[string]any {
			out := []byte{'I', 'I', 42, 0, 8, 0, 0, 0}
			body := []byte{1, 0}
			body = append(body, 0x1A, 0x01, 5, 0, 1, 0, 0, 0, 26, 0, 0, 0)
			body = append(body, 0, 0, 0, 0)
			body = append(body, 72, 0, 0, 0, 1, 0, 0, 0)
			file, _ := Parse(append(out, body...), &Counter{})
			values := file.Get(0, 0x011A, &Counter{})
			if len(values) == 0 {
				return map[string]any{"xres": ""}
			}
			return map[string]any{"xres": values[0]}
		}},
		{Name: "next", Run: func() map[string]any {
			first := BuildIFD([]entry{{0x0100, 3, 1, 100}})
			data := append([]byte{}, first...)
			data[22], data[23], data[24], data[25] = byte(len(data)), 0, 0, 0
			data = append(data, BuildIFD([]entry{{0x0100, 3, 1, 200}})...)
			file, _ := Parse(data, &Counter{})
			return map[string]any{"ifds": len(file.IFDs)}
		}},
	}
}

// Work 跑规模线场景。
func Work(n int) map[string]any {
	entries := make([]entry, n)
	for i := range entries {
		entries[i] = entry{tag: 0x1000 + i, typ: 3, count: 1, value: i}
	}
	file, _ := Parse(BuildIFD(entries), &Counter{})
	c := &Counter{}
	for i := 0; i < n; i++ {
		file.Get(0, 0x1000+i, c)
	}
	return map[string]any{"tags": n, "scanned": c.Scanned}
}

var _ = fmt.Sprintf
