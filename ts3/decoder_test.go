package ts3

import "testing"

func TestDecoderDecodeStruct(t *testing.T) {
	raw := "id=7 name=hello\\sworld active=1 groups=1,2,3"

	var out struct {
		ID     int    `ts3:"id"`
		Name   string `ts3:"name"`
		Active bool   `ts3:"active"`
		Groups []int  `ts3:"groups"`
	}

	if err := NewDecoder().Decode(raw, &out); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if out.ID != 7 || out.Name != "hello world" || !out.Active {
		t.Fatalf("decoded struct mismatch: %+v", out)
	}
	if len(out.Groups) != 3 || out.Groups[0] != 1 || out.Groups[2] != 3 {
		t.Fatalf("decoded groups mismatch: %+v", out.Groups)
	}
}

func TestDecoderDecodeSlice(t *testing.T) {
	raw := "id=1 name=one|id=2 name=two"

	var out []struct {
		ID   int    `ts3:"id"`
		Name string `ts3:"name"`
	}
	if err := NewDecoder().Decode(raw, &out); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if len(out) != 2 || out[0].ID != 1 || out[1].Name != "two" {
		t.Fatalf("unexpected decode result: %+v", out)
	}
}

func TestDecoderRejectsInvalidTarget(t *testing.T) {
	var out struct{}
	if err := NewDecoder().Decode("id=1", out); err == nil {
		t.Fatalf("expected error for non-pointer target")
	}
}
