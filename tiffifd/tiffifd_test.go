package tiffifd

import "testing"

func TestParseLittleEndian(t *testing.T) {
	data := BuildIFD([]entry{{0x0100, 3, 1, 64}})
	file, err := Parse(data, &Counter{})
	if err != nil {
		t.Fatalf("解析失败：%v", err)
	}
	if len(file.IFDs) != 1 {
		t.Fatalf("应有一个 IFD：%d", len(file.IFDs))
	}
}

func TestBadHeader(t *testing.T) {
	if _, err := Parse([]byte("XX"), &Counter{}); err == nil {
		t.Fatal("头不对应报错")
	}
}

func TestTagCount(t *testing.T) {
	data := BuildIFD([]entry{{0x0100, 3, 1, 64}, {0x0101, 3, 1, 32}})
	file, _ := Parse(data, &Counter{})
	if len(file.IFDs[0]) != 2 {
		t.Fatalf("应有两个标签：%d", len(file.IFDs[0]))
	}
}

func TestGetMissing(t *testing.T) {
	data := BuildIFD([]entry{{0x0100, 3, 1, 64}})
	file, _ := Parse(data, &Counter{})
	if file.Get(0, 0x9999, &Counter{}) != nil {
		t.Fatal("缺失标签应返回 nil")
	}
}

func TestGetOutOfRange(t *testing.T) {
	data := BuildIFD([]entry{{0x0100, 3, 1, 64}})
	file, _ := Parse(data, &Counter{})
	if file.Get(9, 0x0100, &Counter{}) != nil {
		t.Fatal("越界 IFD 应返回 nil")
	}
}

func TestCounterCounts(t *testing.T) {
	c := &Counter{}
	Parse(BuildIFD([]entry{{0x0100, 3, 1, 64}}), c)
	if c.Scanned == 0 {
		t.Fatal("应统计比较次数")
	}
}
