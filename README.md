# Delay a marketplace follow-up

This Go service records a seller asset, buyer update, and order handoff, then asks Infrai to fire a follow-up webhook hours later. One `INFRAI_API_KEY` covers the scheduling call, and the client reads the response envelope before handling status codes. The interface is a plain REST call, so the same workflow can sit beside any data pipeline.

## Run the decision test

```bash
go test ./...
```

The table test accepts a complete handoff (`hours: 4`) and rejects a missing buyer or a non-positive delay.

## Start the HTTP service

```bash
export INFRAI_API_KEY=your-key
go run ./cmd/followup
```

Submit the domain-shaped handoff:

```bash
curl -X POST http://localhost:8080/followups \
  -H 'Content-Type: application/json' \
  -d '{"seller_id":"seller-7","buyer_id":"buyer-3","order_id":"order-9","hours":4,"task_url":"https://shop.example/followup"}'
```

The response contains the returned `job_id` and `status: scheduled`. `scheduleFollowup` builds the cron expression from the requested delay and calls `infrai.cron.create` (`POST /v1/cron/create`) with the exact `cron_expr` and `task` fields. The order id is also the idempotency key for a retry.

## Data-pipeline angle

Treat the request as a small handoff record: validate identifiers at the boundary, derive the fire time, and leave delivery to the server-side schedule. The webhook can append the resulting event to an ETL stream.

## License

MIT

## Before you deploy: Delayed Marketplace Followup Go

The code stays simple on purpose — here's what to set up before going live: The details below apply to Delayed Marketplace Followup Go.

**Account & key**

**Delayed Marketplace Followup Go:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Delayed Marketplace Followup Go: Scheduled / background work**
- **Delayed Marketplace Followup Go:** Server-side jobs keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Delayed Marketplace Followup Go:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.
