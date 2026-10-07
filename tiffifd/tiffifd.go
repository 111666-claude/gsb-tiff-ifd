// Package tiffifd 解析 TIFF 目录。
package tiffifd

import (
	"errors"
	"strconv"
)

// Counter 记录比较次数。
type Counter struct {
	Scanned int
}

// File 是解析结果。
type File struct {
	BigEndian bool
	IFDs      []map[int][]string
}

func typeSize(t int) int {
	switch t {
	case 1, 2:
		return 1
	case 3:
		return 2
	case 4:
		return 4
	case 5:
		return 8
	}
	return 1
}

func u16(b []byte) int {
	return int(b[0]) | int(b[1])<<8
}

func u32(b []byte) int {
	return int(b[0]) | int(b[1])<<8 | int(b[2])<<16 | int(b[3])<<24
}

// Parse 解析整份文件。
func Parse(data []byte, c *Counter) (*File, error) {
	if len(data) < 8 || string(data[:2]) != "II" {
		return nil, errors.New("bad-header")
	}
	file := &File{}
	offset := u32(data[4:8])
	for offset > 0 && offset+2 <= len(data) {
		c.Scanned++
		count := u16(data[offset:])
		ifd := map[int][]string{}
		for i := 0; i < count; i++ {
			base := offset + 2 + i*12
			if base+12 > len(data) {
				break
			}
			tag := u16(data[base:])
			typ := u16(data[base+2:])
			n := u32(data[base+4:])
			field := data[base+8 : base+12]
			values := readValues(data, typ, n, u32(field), c)
			ifd[tag] = values
		}
		file.IFDs = append(file.IFDs, ifd)
		break
	}
	return file, nil
}

func readValues(data []byte, typ, count, field int, c *Counter) []string {
	size := typeSize(typ)
	total := size * count
	start := field
	if start+total > len(data) {
		return nil
	}
	var out []string
	for i := 0; i < count; i++ {
		c.Scanned++
		off := start + i*size
		switch typ {
		case 3:
			out = append(out, strconv.Itoa(u16(data[off:])))
		case 4:
			out = append(out, strconv.Itoa(u32(data[off:])))
		case 5:
			out = append(out, strconv.Itoa(u32(data[off:]))+"/"+strconv.Itoa(u32(data[off+4:])))
		default:
			out = append(out, string(data[off:off+size]))
		}
	}
	return out
}

// Get 取某个 IFD 的某个标签。
func (f *File) Get(ifd, tag int, c *Counter) []string {
	if ifd < 0 || ifd >= len(f.IFDs) {
		return nil
	}
	for t, v := range f.IFDs[ifd] {
		c.Scanned++
		if t == tag {
			return v
		}
	}
	return nil
}
