# 🇮🇳 Samvaad SFU — Product Roadmap

> **Mission:** *Made in Bharat* — A sovereign, military-grade WebRTC SFU delivering uncompromising End-to-End Encryption, complete data localization, and flawless operation on constrained Bharatiya infrastructure.

---

## Current Release: v1.0.0 ✅

The foundation of Samvaad SFU is production-ready with the following capabilities:

- Scalable, distributed WebRTC Selective Forwarding Unit (Go + Pion WebRTC v4)
- JWT-authenticated RoomService API (Twirp/gRPC)
- Multi-node distribution via Redis pub/sub (`pkg/routing/redisrouter.go`)
- Simulcast with adaptive layer selection (`pkg/sfu/streamallocator/`)
- Embedded TURN server with TLS (`pkg/service/turn.go`)
- WHIP ingest endpoint (`pkg/service/whipservice.go`)
- SIP gateway support (`pkg/service/sip.go`)
- Prometheus metrics exposure (`pkg/telemetry/`)
- Single binary + Docker + Kubernetes deployments
- SVC codec support: VP9, AV1, H.264, Opus/RED

---

## Upcoming Release: v1.0.1 🚧

### Priority 1 — Bharat SFU: Military-Grade E2EE 🔐

**Goal:** The SFU becomes a zero-knowledge routing fabric. It forwards encrypted media and key-exchange signals but is cryptographically incapable of decoding either.

#### 1.1 Native In-Band Key Exchange via Reliable Data Channels

**Current state:** `pkg/sfu/datachannel/datachannel_writer.go` handles data channel I/O for `samvaad.DataPacket_User` messages. Key exchange currently requires an external signalling path.

**Target state:** Repurpose the existing reliable, ordered WebRTC Data Channel as a zero-knowledge key-exchange broker — no external key server required.

- Implement a new `pkg/sfu/e2ee/` sub-package:
  - `keymgr.go` — Per-participant ephemeral ECDH key-pair generation and lifecycle
  - `keyexchange.go` — X25519 ECDH handshake framing over `samvaad.DataPacket_Kind_USER` reliable data channels
  - `ratchet.go` — Optional Double Ratchet session upgrade for forward secrecy after initial ECDH
- The SFU routing layer (`pkg/routing/signal.go`) serializes and forwards `KeyExchangePacket` protobuf frames; it never inspects key material
- Session keys are derived client-side using HKDF-SHA256; the SFU only sees opaque ciphertext

**Key design invariant:** The SFU node (`pkg/service/roommanager.go`) participates in zero keying operations. The `PacketFactory` pool in `pkg/sfu/sfu.go` allocates buffers for encrypted RTP payloads only.

#### 1.2 Zero-Knowledge Architecture

- Enforce a strict data-plane contract: `pkg/sfu/downtrack.go` and `pkg/sfu/receiver_base.go` MUST NOT log, copy, or inspect RTP payload bytes beyond header parsing
- Add a build-time `//go:build e2ee` constraint to compile the E2EE key manager exclusively when the feature flag is enabled
- SFU-side RTCP (PLI/NACK/RR) continues to function normally — these operate on headers only and do not compromise E2EE

#### 1.3 Strict Data Localization — Telemetry Lockdown

- **Disable all external telemetry endpoints** at compile time for sovereign deployments via `SAMVAAD_SOVEREIGN_BUILD=1`
- Audit and gate all OpenTelemetry exporters (`go.opentelemetry.io/otel/exporters/otlp`) — disable OTLP HTTP trace exports unless explicitly opted in via config
- Webhooks (`config-sample.yaml` → `webhook.urls`) restricted to RFC-1918 / RFC-4193 private address space when `sovereign_mode: true` is set in config
- Prometheus metrics (`prometheus_port`) remain on-premise only; outbound scrape endpoints are blocked

#### 1.4 FIPS 140-2 Compliance (`boringcrypto`)

- Modify `Dockerfile` builder stage: `FROM golang:1.26-alpine` → `FROM golang:1.26` + `GOEXPERIMENT=boringcrypto go build`
- Update `magefile.go` `Build()` target to inject `GOEXPERIMENT=boringcrypto` for FIPS release builds
- Compile targets: all TLS sessions (gRPC, Twirp, WebSocket signal) and DTLS/SRTP sessions (Pion `pion/dtls/v3`, `pion/srtp/v3`) use BoringCrypto-backed primitives
- Add a CI job in `.github/workflows/buildtest.yaml`: FIPS build smoke-test with `go tool nm` to verify BoringCrypto symbol presence

---

### Priority 2 — High Efficiency: Low CPU & Low RAM Mode ⚡

**Goal:** Samvaad SFU runs flawlessly on a 512 MB RAM, 1 vCPU sovereign edge node — a ₹299/month VPS — without OOM kills or dropped audio.

#### 2.1 `SAMVAAD_LOW_RESOURCE_MODE` Runtime Flag

Introduce a new top-level config key and matching environment variable:

```yaml
# config-sample.yaml addition
low_resource_mode: false   # set true for edge / constrained nodes
```

```bash
export SAMVAAD_LOW_RESOURCE_MODE=true
```

Parsed in `pkg/config/` and propagated via dependency injection (`pkg/service/wire.go`) to all subsystems that honor it.

#### 2.2 Go GC Memory Tuning

Applied automatically when `low_resource_mode: true`:

| Parameter | Standard | Low Resource |
|-----------|----------|--------------|
| `GOMEMLIMIT` | unlimited | `400MiB` |
| `GOGC` | 100 | `20` |
| `GOMAXPROCS` | NumCPU | `1` (single-CPU nodes) |

- Set programmatically via `runtime/debug.SetMemoryLimit()` and `debug.SetGCPercent()` at startup in `cmd/server/main.go`
- Emit a structured log warning when GC pressure exceeds 15% CPU time (tracked via `runtime.ReadMemStats`)

#### 2.3 Adaptive Buffer & NACK Window Reduction

Modify `pkg/sfu/buffer/` and `pkg/sfu/sequencer.go`:

| Buffer | Standard | Low Resource |
|--------|----------|--------------|
| `packet_buffer_size_video` | 500 packets | 100 packets |
| `packet_buffer_size_audio` | 200 packets | 50 packets |
| NACK window | 1000 ms | 300 ms |
| `PacketFactory` pool max | unbounded | 256 slots |

- Halt background Prometheus label-cardinality aggregation loops in `pkg/telemetry/` when low-resource mode is active
- Reduce `pkg/sfu/rtpstats/` histogram bucket counts from 32 → 8 in low-resource mode

#### 2.4 Smart Simulcast Layer Governor

Extends the existing `pkg/sfu/streamallocator/streamallocator.go` with a CPU-aware layer enforcement policy:

```
CPU usage < 60%  →  allow High (1080p) layer
CPU usage 60–80% →  cap at Mid (720p) layer, warn subscriber
CPU usage > 80%  →  enforce Low (360p) layer; audio NEVER dropped
CPU usage > 95%  →  pause all video forwarding; audio-only fallback
```

- Uses `github.com/mackerelio/go-osstat` (already in `go.mod`) for real-time CPU sampling — no new dependency
- The governor runs in `pkg/sfu/streamallocator/streamallocator.go` as a new `cpuGovernorLoop()` goroutine, polling at 2-second intervals
- Governor state is exposed as a Prometheus gauge: `samvaad_sfu_cpu_layer_cap{node="...", cap="high|mid|low|audio_only"}`
- Layer enforcement signals existing `pkg/sfu/forwarder.go` `SetMaxSpatialLayer()` — no new RTP path

#### 2.5 Goroutine Budget & Worker Scaling

- Cap `workerpool` (`github.com/gammazero/workerpool`) worker counts relative to `GOMAXPROCS` in low-resource mode
- In single-CPU mode, collapse separate signal and RTC goroutine pools into a single cooperative scheduler
- Idle room goroutines (`pkg/service/roommanager.go`) yield CPU after 5 s of silence instead of the default 30 s

---

### Priority 3 — Any-Topology Support 🌐

**Goal:** Samvaad SFU supports every deployment topology from a single Raspberry Pi to a geo-distributed sovereign cloud.

#### 3.1 Topology Modes

| Mode | Router | Use Case |
|------|--------|----------|
| `standalone` | `pkg/routing/localrouter.go` | Single node, air-gapped, classroom LAN |
| `clustered` | `pkg/routing/redisrouter.go` | Multi-node, state via Redis |
| `mesh` | New: `pkg/routing/meshrouter.go` | Peer-to-peer node federation, no central Redis |
| `edge` | `standalone` + `low_resource_mode` | Sub-1GB RAM sovereign edge node |
| `hybrid` | `clustered` with regional routing | Geo-distributed Bharat CDN fabric |

#### 3.2 Mesh Router (`pkg/routing/meshrouter.go`) — New

For deployments where Redis is unavailable or undesirable:

- Nodes discover peers via mDNS (LAN) or a static `peers:` config list
- Room routing table replicated using a gossip protocol over the existing PSRPC transport (`github.com/msmclass/samvaad/pkg/samvaad/psrpc`)
- Node health tracked via `pkg/routing/nodestats.go` heartbeat extended with gossip epoch counters
- Graceful fallback: if gossip quorum lost, node operates as `standalone` and rejects cross-node joins

#### 3.3 Regional-Aware Load Balancing

Enhance `pkg/routing/selector/` with:

- `LatencyAwareSelector`: prefer nodes with lowest RTT to participant (using STUN binding latency measurement at ICE time)
- `SovereignRegionSelector`: hard-pin participants to nodes within a declared `sovereign_zone` — guarantees data never crosses defined geographic or legal boundaries

#### 3.4 ICE / NAT Traversal Hardening for Bharatiya Networks

Indian network conditions (CGNAT, ISP-level NAT, mobile carriers) require aggressive ICE handling:

- Enable ICE TCP fallback by default (`allow_tcp_fallback: true`)
- Promote `tcp_port: 443` as the default TURN/TLS port for maximum firewall penetration
- Add `ice_restart_on_dtls_failure: true` — automatic ICE restart on DTLS handshake timeout (common on JIO/BSNL networks)
- Document and test TURN relay path for Airtel, JIO, BSNL, Vi network profiles

---

## Release Timeline

```
v1.0.0  ████████████████████ RELEASED
v1.0.1  ████████░░░░░░░░░░░░ Q3 2026  (E2EE + Low Resource Mode)
v1.1.0  ░░░░░░░░░░░░░░░░░░░░ Q4 2026  (Mesh Topology + FIPS CI)
v1.2.0  ░░░░░░░░░░░░░░░░░░░░ Q1 2027  (Geo-sovereign routing + Regional CDN)
```

---

## Guiding Principles

| Principle | Implementation |
|-----------|----------------|
| **Zero-Knowledge SFU** | SFU routes bytes; never holds keys |
| **Data Localization** | All data bounded by sovereign network perimeter |
| **Bharatiya Infrastructure First** | Default configs tuned for ₹299 VPS, JIO/BSNL networks |
| **No External Dependencies at Runtime** | Single binary; Redis optional; no SaaS telemetry |
| **FIPS by Default for GOI Deployments** | `boringcrypto` build for sensitive sectors |

---

*Jai Hind 🇮🇳 — Built with Go, Pion WebRTC, and Bharatiya engineering excellence.*
