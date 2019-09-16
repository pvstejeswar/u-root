// Copyright 2016-2019 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	flag "github.com/spf13/pflag"
)

const (
	testDataDir = "testdata"
)

func resetFlags() {
	*flagFromDump = ""
	*flagType = nil
}

func testOutput(t *testing.T, dumpFile string, args []string, expectedOutFile string) {
	dumpFile = filepath.Join(testDataDir, dumpFile)
	expectedOutFile = filepath.Join(testDataDir, expectedOutFile)
	actualOutFile := fmt.Sprintf("%s.actual", expectedOutFile)
	os.Remove(actualOutFile)
	os.Args = []string{os.Args[0], "--from-dump", dumpFile}
	os.Args = append(os.Args, args...)
	flag.Parse()
	defer resetFlags()
	out := bytes.NewBuffer(nil)
	if err := dmiDecode(out); err != nil {
		t.Errorf("%+v %+v %+v: error: %s", dumpFile, args, expectedOutFile, err)
		return
	}
	actualOut := out.Bytes()
	expectedOut, err := ioutil.ReadFile(expectedOutFile)
	if err != nil {
		t.Errorf("%+v %+v %+v: failed to load %s: %s", dumpFile, args, expectedOutFile, expectedOutFile, err)
		return
	}
	if bytes.Compare(actualOut, expectedOut) != 0 {
		ioutil.WriteFile(actualOutFile, actualOut, 0644)
		t.Errorf("%+v %+v %+v: output mismatch, see %s", dumpFile, args, expectedOutFile, actualOutFile)
		diffOut, _ := exec.Command("diff", "-u", expectedOutFile, actualOutFile).CombinedOutput()
		t.Errorf("%+v %+v %+v: diff:\n%s", dumpFile, args, expectedOutFile, string(diffOut))
	}
}

func TestDMIDecode(t *testing.T) {
	testOutput(t, "UX307LA.bin", nil, "UX307LA.txt")
	testOutput(t, "UX307LA.bin", []string{"-t", "system"}, "UX307LA.system.txt")
	testOutput(t, "UX307LA.bin", []string{"-t", "1,131"}, "UX307LA.1_131.txt")
	testOutput(t, "T480S.bin", nil, "T480S.txt")
}

func testDumpBin(t *testing.T, entryData, expectedOutData []byte) {
	tmpfile, err := ioutil.TempFile("", "dmidecode")
	if err != nil {
		t.Fatalf("error creating temp file: %s", err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())
	textOut := bytes.NewBuffer(nil)
	if err := dumpBin(
		textOut,
		entryData,
		[]byte{0xaa, 0xbb}, // dummy
		tmpfile.Name(),
	); err != nil {
		t.Fatalf("failed to dump bin: %s", err)
	}
	outData, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to read output: %s", err)
	}
	if bytes.Compare(outData, expectedOutData) != 0 {
		t.Fatalf("binary data mismatch,\nexpected:\n  %s\ngot:\n  %s", hex.EncodeToString(expectedOutData), hex.EncodeToString(outData))
	}
}

func TestDMIDecodeDumpBin32(t *testing.T) {
	// We expect entry point address to be rewritten and chcecksum adjusted.
	testDumpBin(
		t,
		[]byte{
			0x5f, 0x53, 0x4d, 0x5f, 0x64, 0x1f, 0x02, 0x08, 0x14, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x5f, 0x44, 0x4d, 0x49, 0x5f, 0x37, 0x6e, 0x08, 0x00, 0x50, 0x7c, 0xac, 0x1b, 0x00, 0x28,
		},
		[]byte{
			0x5f, 0x53, 0x4d, 0x5f, 0x64, 0x1f, 0x02, 0x08, 0x14, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x5f, 0x44, 0x4d, 0x49, 0x5f, 0x8f, 0x6e, 0x08, 0x20, 0x00, 0x00, 0x00, 0x1b, 0x00, 0x28, 0x00,
			0xaa, 0xbb,
		},
	)
}

func TestDMIDecodeDumpBin64(t *testing.T) {
	// We expect entry point address to be rewritten and chcecksum adjusted.
	testDumpBin(
		t,
		[]byte{
			0x5f, 0x53, 0x4d, 0x33, 0x5f, 0xe6, 0x18, 0x03, 0x00, 0x00, 0x01, 0x00, 0xe3, 0x0b, 0x00, 0x00,
			0x00, 0xe0, 0x10, 0x8f, 0x00, 0x00, 0x00, 0x00,
		},
		[]byte{
			0x5f, 0x53, 0x4d, 0x33, 0x5f, 0x45, 0x18, 0x03, 0x00, 0x00, 0x01, 0x00, 0xe3, 0x0b, 0x00, 0x00,
			0x20, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0xaa, 0xbb,
		},
	)
}