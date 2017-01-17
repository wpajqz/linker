package version

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestFull(t *testing.T) {
	version := Full()
	arr := strings.Split(version, ".")
	if len(arr) != 3 {
		t.Fatalf("Version string error: %s", version)
	}

	proto, err := strconv.ParseInt(arr[0], 10, 64)
	if err != nil || proto < 0 {
		t.Fatalf("Version proto error")
	}

	major, err := strconv.ParseInt(arr[1], 10, 64)
	if err != nil || major < 0 {
		t.Fatalf("Version major error")
	}

	minor, err := strconv.ParseInt(arr[2], 10, 64)
	if err != nil || minor < 0 {
		t.Fatalf("Version minor error")
	}
}

func TestVersion(t *testing.T) {
	proto := Proto(Full())
	major := Major(Full())
	minor := Minor(Full())
	parseVerion := fmt.Sprintf("%d.%d.%d", proto, major, minor)
	version := Full()
	if parseVerion != version {
		t.Fatalf("Get version incorrect, version [%s], parseVerion [%s]", version, parseVerion)
	}
}
