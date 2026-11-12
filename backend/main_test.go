package main

import "testing"

func TestSimulatePayment(t *testing.T) {
	cases := []struct {
		name, number, expiry, cvc string
		wantErr                   bool
	}{
		{"valid test card", "4242 4242 4242 4242", "12/30", "123", false},
		{"declined card", "4000000000000002", "12/30", "123", true},
		{"too short", "4242", "12/30", "123", true},
		{"letters", "abcd efgh ijkl mnop", "12/30", "123", true},
		{"missing cvc", "4242424242424242", "12/30", "", true},
	}
	for _, c := range cases {
		ref, msg := simulatePayment(c.number, c.expiry, c.cvc)
		if (msg != "") != c.wantErr {
			t.Errorf("%s: got error %q, wantErr=%v", c.name, msg, c.wantErr)
		}
		if !c.wantErr && ref == "" {
			t.Errorf("%s: expected a payment reference", c.name)
		}
	}
}

func TestSlugify(t *testing.T) {
	if got := slugify("Baldur's Gate 3 Steam"); got != "baldur-s-gate-3-steam" {
		t.Errorf("slugify = %q", got)
	}
}

func TestNewKeyFormat(t *testing.T) {
	k := newKey()
	if len(k) != 17 || k[5] != '-' || k[11] != '-' {
		t.Errorf("unexpected key format %q", k)
	}
}
