package main

import "testing"

func TestParseADBDevicesReturnsOnlyReadyDevices(t *testing.T) {
	got := parseADBDevices([]byte("List of devices attached\nR5GL2218DMR\tdevice\nbad\toffline\nunauth\tunauthorized\n\n"))
	if len(got) != 1 || got[0] != "R5GL2218DMR" {
		t.Fatalf("unexpected devices: %#v", got)
	}
}
