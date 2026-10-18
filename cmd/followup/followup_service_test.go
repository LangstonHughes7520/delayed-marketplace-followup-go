package main

import (
	"testing"
	"time"
)

func TestScheduleFollowupValidatesHandoff(t *testing.T) {
	tests := []struct {
		name    string
		input   FollowupRequest
		wantErr bool
	}{
		{"complete handoff", FollowupRequest{SellerID: "seller-7", BuyerID: "buyer-3", OrderID: "order-9", Hours: 4, TaskURL: "https://shop.test/followup"}, false},
		{"missing buyer update", FollowupRequest{SellerID: "seller-7", OrderID: "order-9", Hours: 4, TaskURL: "https://shop.test/followup"}, true},
		{"zero delay", FollowupRequest{SellerID: "seller-7", BuyerID: "buyer-3", OrderID: "order-9", Hours: 0, TaskURL: "https://shop.test/followup"}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateFollowup(tc.input)
			if tc.wantErr != (err != nil) {
				t.Fatalf("wantErr=%v, err=%v", tc.wantErr, err)
			}
		})
	}
}

func TestCronCreatePayloadSchedulesOneShot(t *testing.T) {
	in := FollowupRequest{SellerID: "seller-7", BuyerID: "buyer-3", OrderID: "order-9", Hours: 4, TaskURL: "https://shop.test/followup"}
	now := time.Date(2026, time.October, 4, 8, 30, 0, 0, time.FixedZone("CST", 8*60*60))

	payload := cronCreatePayload(in, now)
	if got, want := payload["run_at"], "2026-10-04T04:30:00Z"; got != want {
		t.Fatalf("run_at=%q, want %q", got, want)
	}
	if _, exists := payload["cron_expr"]; exists {
		t.Fatal("one-shot follow-up must not contain cron_expr")
	}
	if payload["task"] == "" {
		t.Fatal("task is required")
	}
}
