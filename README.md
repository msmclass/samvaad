# Samvaad: Real-time Communications Server

**Samvaad** is a scalable, modern WebRTC-based Selective Forwarding Unit (SFU) server designed for building real-time video, audio, and data communication applications.

Built for the MSM Education ecosystem and hosted at [MSM Class](https://msmclass.in), Samvaad provides a production-ready foundation for interactive virtual classrooms, conferencing, and collaborative applications. It serves as our dedicated SFU for the Samvaad video conference system.

Samvaad's server is written in Go using the awesome [Pion WebRTC](https://github.com/pion/webrtc) implementation.

[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/msmclass/samvaad)](https://github.com/msmclass/samvaad/releases/latest)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/msmclass/samvaad/buildtest.yaml?branch=master)](https://github.com/msmclass/samvaad/actions/workflows/buildtest.yaml)
[![License](https://img.shields.io/github/license/msmclass/samvaad)](https://github.com/msmclass/samvaad/blob/main/LICENSE)

## Features

-   Scalable, distributed WebRTC SFU (Selective Forwarding Unit)
-   Production-ready with JWT authentication
-   Robust networking and connectivity, UDP/TCP/TURN
-   Easy to deploy: single binary, Docker, or Kubernetes
-   Advanced features including:
    -   Speaker detection and automatic layout management
    -   Simulcast for adaptive bitrate streaming
    -   Selective subscription for bandwidth optimization
    -   Moderation and participant management APIs
    -   End-to-end encryption support
    -   SVC codecs (VP9, AV1)
    -   Webhook integration for room events
    -   Distributed and multi-region support via Redis

## Documentation & Getting Started

Visit the [Samvaad documentation](https://github.com/msmclass/samvaad/wiki) for detailed guides and API documentation.

## Quick Links

- **GitHub**: https://github.com/msmclass/samvaad
- **Issues**: https://github.com/msmclass/samvaad/issues
- **Configuration**: See [config-sample.yaml](config-sample.yaml)

## Building from Source

### Prerequisites

- Go 1.26 or later
- `git`

### Build

```bash
# Clone repository
git clone https://github.com/msmclass/samvaad.git
cd samvaad

# Setup build tools
./bootstrap.sh

# Build server
mage build
```

The binary will be available at `./bin/samvaad-server`

## Automated Releases & Docker Builds

### Automated Release Pipeline
Our GitHub Actions pipeline automatically detects pushed tags and creates production releases based on the `<version>.<release>` or `<version>.<release>.<patch>` format:
- **Triggers**: Any Git tag starting with a `v` prefix (e.g. `v1.0`, `v1.0.1`, `v2.0`) will automatically initiate a compile-and-release workflow via GoReleaser.
- **Artifacts**: Binaries are automatically compiled, zipped, and attached to the GitHub Release for multiple operating systems and architectures.

### Multi-Platform Docker Images (AMD64 & ARM64)
Production Docker images are automatically compiled for all major server environments using Docker Buildx and published to **GitHub Container Registry (GHCR)**:
- **Base Registry**: `ghcr.io/msmclass/samvaad-server`
- **Supported Architectures**:
  - `linux/amd64` (Standard x86/Intel/AMD 64-bit platforms)
  - `linux/arm64` (ARM 64-bit platforms, Apple Silicon, AWS Graviton, etc.)
- **Tags**: On each release, the image is automatically tagged with `latest`, the exact version tag (e.g. `v1.0.1`), and branch tags.

