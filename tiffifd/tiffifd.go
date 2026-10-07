// Package tiffifd 解析 TIFF 目录。
package tiffifd

import (
	"encoding/binary"
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

// Parse 解析整份文件。
func Parse(data []byte, c *Counter) (*File, error) {
	if len(data) < 8 {
		return nil, errors.New("bad-header")
	}
	var order binary.ByteOrder
	file := &File{}
	switch string(data[:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
		file.BigEndian = true
	default:
		return nil, errors.New("bad-header")
	}
	if order.Uint16(data[2:4]) != 42 {
		return nil, errors.New("bad-header")
	}
	offset := int(order.Uint32(data[4:8]))
	for offset > 0 && offset+2 <= len(data) {
		c.Scanned++
		count := int(order.Uint16(data[offset:]))
		ifd := map[int][]string{}
		for i := 0; i < count; i++ {
			base := offset + 2 + i*12
			if base+12 > len(data) {
				break
			}
			tag := int(order.Uint16(data[base:]))
			typ := int(order.Uint16(data[base+2:]))
			n := int(order.Uint32(data[base+4:]))
			field := data[base+8 : base+12]
			ifd[tag] = readValues(data, order, typ, n, field, c)
		}
		file.IFDs = append(file.IFDs, ifd)
		next := offset + 2 + count*12
		if next+4 > len(data) {
			break
		}
		offset = int(order.Uint32(data[next:]))
	}
	return file, nil
}

func readValues(data []byte, order binary.ByteOrder, typ, count int, field []byte, c *Counter) []string {
	size := typeSize(typ)
	total := size * count
	var buf []byte
	if total <= 4 {
		buf = field[:total]
	} else {
		start := int(order.Uint32(field))
		if start+total > len(data) {
			return nil
		}
		buf = data[start : start+total]
	}
	var out []string
	for i := 0; i < count; i++ {
		c.Scanned++
		off := i * size
		switch typ {
		case 3:
			out = append(out, strconv.Itoa(int(order.Uint16(buf[off:]))))
		case 4:
			out = append(out, strconv.Itoa(int(order.Uint32(buf[off:]))))
		case 5:
			out = append(out, strconv.Itoa(int(order.Uint32(buf[off:])))+"/"+strconv.Itoa(int(order.Uint32(buf[off+4:]))))
		default:
			out = append(out, string(buf[off:off+size]))
		}
	}
	return out
}

// Get 取某个 IFD 的某个标签。
func (f *File) Get(ifd, tag int, c *Counter) []string {
	if ifd < 0 || ifd >= len(f.IFDs) {
		return nil
	}
	c.Scanned++
	return f.IFDs[ifd][tag]
}
