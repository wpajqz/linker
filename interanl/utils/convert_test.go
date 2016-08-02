package utils

import "testing"

func TestConvert(t *testing.T) {
	var a uint32 = 199211
	bytes := Uint32ToBytes(a)
	var b uint32 = BytesToUint32(bytes)
	if a != b {
		t.Error("convert error")
	}

	a = 0xffffffff
	b = BytesToUint32(Uint32ToBytes(a))
	if a != b {
		t.Error("convert error")
	}

	var c uint16 = 0xfefe
	bytes2 := Uint16ToBytes(c)
	var d uint16 = BytesToUint16(bytes2)
	if c != d {
		t.Error("convert error")
	}

	c = 65535
	d = BytesToUint16(Uint16ToBytes(c))
	if c != d {
		t.Error("convert error")
	}

	var e int32 = 98765
	bytes3 := Int32ToBytes(e)
	var f int32 = BytesToInt32(bytes3)
	if e != f {
		t.Error("convert error")
	}

	var g int16 = 0x7FFF
	bytes4 := Int16ToBytes(g)
	var h int16 = BytesToInt16(bytes4)
	if g != h {
		t.Error("convert error")
	}

	var i int16 = 32767
	j := BytesToInt16(Int16ToBytes(i))
	if i != j {
		t.Error("convert error")
	}
}
