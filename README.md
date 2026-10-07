# tiff-ifd

解析 TIFF 目录：字节序、条目类型与计数、内联与偏移值、IFD 链与标签索引。只用标准库。

```
go test ./...
go vet ./...
go run ./cmd/tiffifd --sample=endian
go run ./cmd/tiffifd --sample=inline
go run ./cmd/tiffifd --sample=rational
go run ./cmd/tiffifd --sample=next
go run ./cmd/tiffifd --sample=work
```

## 口径

- **字节序**：前两字节 `II` 是小端、`MM` 是大端，之后所有多字节字段都按它读；再接 42 与 IFD0 偏移。
- **条目**：12 字节：标签(2)、类型(2)、计数(4)、值或偏移(4)。
- **类型**：BYTE=1、ASCII=2、SHORT=3、LONG=4、RATIONAL=5，单个元素分别是 1/1/2/4/8 字节。
- **内联**：值总字节数不超过 4 时直接放在值字段里，否则那 4 字节是从文件头算起的偏移。
- **链**：IFD 末尾 4 字节是下一个 IFD 的偏移，为 0 结束。
- **查找**：按标签走索引。

## 不变量

- 多值字段按文件字节序解出。
- 内联值不当作偏移读。
- `scanned` 不随标签数乘查询次数放大：八百个标签查八百次的 `scanned` 不超过 6000。

## 输出契约

`IFD` 是标签到值列表的映射；`File` 含字节序与 IFD 列表。场景打印一行 JSON。
`--sample=work` 打印 `{"tags": N, "scanned": N}`。
