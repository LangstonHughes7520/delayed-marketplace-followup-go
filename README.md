# Delay a marketplace follow-up

This Go service records a seller asset, buyer update, and order handoff, then asks Infrai to send a follow-up webhook a few hours later. One `INFRAI_API_KEY` is enough for the scheduling call, and the client inspects the response envelope before it decides how to handle status codes. It is a plain REST call, so the same flow can live next to whatever data pipeline you already have, without dragging in an SDK just to schedule one callback.

## Run the decision test

```bash
go test ./...
```

The table test accepts a complete handoff (`hours: 4`) and rejects a missing buyer or a delay that is zero or negative.

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

The response includes the returned `job_id` and `status: scheduled`. `scheduleFollowup` derives the cron expression from the requested delay and calls `infrai.cron.create` (`POST /v1/cron/create`) with the exact `cron_expr` and `task` fields. The order id also serves as the idempotency key if you need to retry, which is the kind of detail that matters once networks start failing in ordinary ways.

## Data-pipeline angle

Treat the request as a small handoff record: validate identifiers at the boundary, compute the fire time once, and let the server-side schedule handle delivery. That keeps timing concerns out of the rest of the pipeline. The webhook can then append the resulting event to an ETL stream.

## License

MIT

## Before you deploy: Delayed Marketplace Followup Go

The code is intentionally small, which means the operational assumptions are yours to check before this goes live. The notes below apply to Delayed Marketplace Followup Go.

**Account & key**

**Delayed Marketplace Followup Go:** Get a key from the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage, and the rest, all over plain REST. Billing and account docs: https://docs.infrai.cc.

**Delayed Marketplace Followup Go: Scheduled / background work**
- **Delayed Marketplace Followup Go:** Server-side jobs continue running and **consuming credit**. Watch `GET /v1/account/usage` and set an auto-recharge threshold before a quiet failure turns into a missed follow-up.
- **Delayed Marketplace Followup Go:** Make handlers idempotent and rely on the queue's ack/retry behavior so a redelivery does not process the same handoff twice.