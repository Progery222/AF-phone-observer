package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("adbserial", flag.ContinueOnError)
	mode := fs.String("mode", "first", "mode: first or list")
	adbBin := fs.String("adb", envDefault("ADB", "adb"), "adb binary")
	if err := fs.Parse(args); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, *adbBin, "devices").Output()
	if err != nil {
		return err
	}
	devices := parseADBDevices(out)
	switch *mode {
	case "first":
		if len(devices) == 0 {
			return errors.New("ADB не видит ни одного устройства со статусом device")
		}
		fmt.Println(devices[0])
	case "list":
		if len(devices) == 0 {
			return errors.New("ADB не видит ни одного устройства со статусом device")
		}
		for _, serial := range devices {
			fmt.Println(serial)
		}
	default:
		return fmt.Errorf("неизвестный mode: %s", *mode)
	}
	return nil
}

func parseADBDevices(out []byte) []string {
	lines := bytes.Split(out, []byte{'\n'})
	devices := make([]string, 0, len(lines))
	for _, line := range lines {
		fields := strings.Fields(string(line))
		if len(fields) >= 2 && fields[1] == "device" {
			devices = append(devices, fields[0])
		}
	}
	return devices
}

func envDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
