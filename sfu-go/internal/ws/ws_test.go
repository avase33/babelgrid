package ws

import (
	"bufio"
	"bytes"
	"testing"
)

func TestAcceptKeyRFCExample(t *testing.T) {
	if got := AcceptKey("dGhlIHNhbXBsZSBub25jZQ=="); got != "s3pPLMBiTxaQ9kYGzzhZRbK+xOo=" {
		t.Fatalf("AcceptKey = %q", got)
	}
}

func maskedFrame(op byte, payload []byte) []byte {
	mask := [4]byte{0x11, 0x22, 0x33, 0x44}
	var b bytes.Buffer
	b.WriteByte(0x80 | op)
	n := len(payload)
	if n < 126 {
		b.WriteByte(0x80 | byte(n))
	} else {
		b.WriteByte(0x80 | 126)
		b.WriteByte(byte(n >> 8))
		b.WriteByte(byte(n))
	}
	b.Write(mask[:])
	for i, c := range payload {
		b.WriteByte(c ^ mask[i&3])
	}
	return b.Bytes()
}

func TestReadMaskedFrame(t *testing.T) {
	f := maskedFrame(opText, []byte("audio-frame"))
	op, p, err := readFrame(bufio.NewReader(bytes.NewReader(f)))
	if err != nil || op != opText || string(p) != "audio-frame" {
		t.Fatalf("op=%d p=%q err=%v", op, p, err)
	}
}

func TestReadExtendedLength(t *testing.T) {
	big := bytes.Repeat([]byte("x"), 500)
	f := maskedFrame(opText, big)
	_, p, err := readFrame(bufio.NewReader(bytes.NewReader(f)))
	if err != nil || len(p) != 500 {
		t.Fatalf("len=%d err=%v", len(p), err)
	}
}

func TestWriteFrameUnmasked(t *testing.T) {
	var buf bytes.Buffer
	bw := bufio.NewWriter(&buf)
	if err := writeFrame(bw, opText, []byte("hi")); err != nil {
		t.Fatal(err)
	}
	out := buf.Bytes()
	if out[0] != 0x81 || out[1] != 0x02 || string(out[2:]) != "hi" {
		t.Fatalf("bad frame: %v", out)
	}
}
