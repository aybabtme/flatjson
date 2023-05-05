package flatjson

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"

	"github.com/buger/jsonparser"
	"github.com/valyala/fastjson"
)

func BenchmarkFlatJSON(b *testing.B) {
	lines := loadObjects(b, "dump.json")

	b.ResetTimer()
	for i, line := range lines {
		b.SetBytes(int64(len(line)))
		_, found, err := ScanObject(line, 0, &Callbacks{
			OnRaw: func(prefixes Prefixes, name Prefix, value Pos) {
				if !name.IsArrayIndex() && !name.IsObjectKey() {
					panic("what")
				}
			},
		})
		if err != nil {
			b.Errorf("line %d: %v", i, err)
		}
		if !found {
			b.Errorf("should have found an object")
		}
	}
}

func BenchmarkEncodingJSON(b *testing.B) {
	lines := loadObjects(b, "dump.json")
	q := struct{}{}
	b.ResetTimer()
	for i, line := range lines {
		b.SetBytes(int64(len(line)))

		err := json.Unmarshal(line, &q)
		if err != nil {
			b.Errorf("line %d: %v", i, err)
		}
	}
}

func Benchmark_buger_jsonparse(b *testing.B) {
	lines := loadObjects(b, "dump.json")

	b.ResetTimer()
	for i, line := range lines {
		b.SetBytes(int64(len(line)))
		err := jsonparser.ObjectEach(line, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
			if len(key) == 0 {
				panic("what")
			}
			return nil
		})
		if err != nil {
			b.Errorf("line %d (%q): %v", i, string(line), err)
		}
	}
}

func Benchmark_valyala_fastjson(b *testing.B) {
	lines := loadObjects(b, "dump.json")

	b.ResetTimer()
	for i, line := range lines {
		b.SetBytes(int64(len(line)))
		v, err := fastjson.ParseBytes(line)
		if err != nil {
			b.Errorf("line %d (%q): %v", i, string(line), err)
		}
		if v.Type() != fastjson.TypeObject {
			b.Error("not an object")
		}
	}
}

func loadObjects(b *testing.B, filename string) [][]byte {
	var objects [][]byte

	f, err := os.Open(filename)
	if err != nil {
		b.Error(err)
	}
	defer f.Close()

	scan := bufio.NewScanner(f)
	scan.Split(bufio.ScanLines)

	for scan.Scan() {
		if len(objects) == b.N {
			return objects
		}
		// Text() to bytes, to force a copy of the memory,
		// otherwise Bytes() will recycle the bytes
		objects = append(objects, []byte(scan.Text()))
	}

	if scan.Err() != nil {
		b.Error(scan.Err())
	}

	for i := 0; len(objects) < b.N; i++ {
		objects = append(objects, []byte(string(objects[i])))
	}

	if len(objects) < b.N {
		b.Errorf("only %d lines in %q", len(objects), filename)
	}

	return objects
}
