package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type FollowupRequest struct {
	SellerID string `json:"seller_id"`
	BuyerID  string `json:"buyer_id"`
	OrderID  string `json:"order_id"`
	Hours    int    `json:"hours"`
	TaskURL  string `json:"task_url"`
}

type followupResponse struct {
	JobID string `json:"job_id"`
}

func scheduleFollowup(c *infraiClient, in FollowupRequest) (string, error) {
	if err := validateFollowup(in); err != nil {
		return "", err
	}
	if c == nil {
		return "", fmt.Errorf("client is required")
	}
	var out followupResponse
	// infrai.cron.create is the domain operation represented by this REST call.
	err := c.post("/v1/cron/create", cronCreatePayload(in, time.Now()), in.OrderID, &out)
	return out.JobID, err
}

func cronCreatePayload(in FollowupRequest, now time.Time) map[string]string {
	when := now.Add(time.Duration(in.Hours) * time.Hour).UTC()
	return map[string]string{
		"run_at": when.Format(time.RFC3339),
		"task":   in.TaskURL + "?seller_id=" + in.SellerID + "&buyer_id=" + in.BuyerID + "&order_id=" + in.OrderID,
	}
}

func validateFollowup(in FollowupRequest) error {
	if in.SellerID == "" || in.BuyerID == "" || in.OrderID == "" || in.TaskURL == "" || in.Hours < 1 {
		return fmt.Errorf("seller_id, buyer_id, order_id, task_url and positive hours are required")
	}
	return nil
}

func handler(c *infraiClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/followups" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		var in FollowupRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		jobID, err := scheduleFollowup(c, in)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"job_id": jobID, "status": "scheduled"})
	})
}

func main() {
	c, err := newInfraiClient()
	if err != nil {
		panic(err)
	}
	port := "8080"
	if value := getenv("PORT"); value != "" {
		port = value
	}
	fmt.Println("follow-up service listening on :" + port)
	if err := http.ListenAndServe(":"+port, handler(c)); err != nil {
		panic(err)
	}
}

func getenv(key string) string { return os.Getenv(key) }
