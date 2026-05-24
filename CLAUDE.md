# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**geektime-downloader** is a Go CLI tool for downloading course content from Geek Time (极客时间), a Chinese tech education platform. It supports downloading columns (PDF/Markdown/audio), video courses, daily lessons, QCon Plus case studies, university training videos, and enterprise courses.

## Commands

### Build
```bash
# Quick build (current platform)
go build -o geektime-downloader

# Multi-platform build via script
./build.sh                     # All platforms
./build.sh -p darwin-arm64     # Specific platform
./build.sh -c                  # Clean dist/
./build.sh --list              # List supported platforms
```

### Run
```bash
# Interactive mode (default)
./geektime-downloader --gcid "xxx" --gcess "yyy"

# Non-interactive with course IDs
./geektime-downloader --gcid "xxx" --gcess "yyy" --course-ids 100056701 --product-type normal

# YAML config batch mode
./geektime-downloader --config courses.yaml
```

### Test & Lint
```bash
go test ./...                  # Run all tests (only filenamify_test.go exists)
golangci-lint run              # Lint (gosimple, staticcheck, stylecheck, unused)
```

### Release
```bash
goreleaser release --snapshot  # Local snapshot release
```

## Architecture

### Entry Point
`main.go` delegates to `cmd.Execute()` — there are no subcommands, all functionality is flag-driven via Cobra.

### Directory Structure
```
cmd/                          # CLI (Cobra root command, ~1156 lines in root.go)
internal/
  geektime/                   # API client layer (geektime.go, client.go, account.go, enterprise.go, university.go)
    response/                 # API response structs (12 files)
  pdf/                        # PDF generation via chromedp (headless Chrome)
  markdown/                   # HTML-to-Markdown conversion + image download
  audio/                      # Audio (.mp3) download
  video/                      # Video download (HLS/m3u8 + TS decryption)
    vod/                      # Alibaba Cloud VOD API integration
  config/                     # Cookie persistence to os.UserConfigDir()
  pkg/
    crypto/                   # AES, HMAC-SHA1, RSA utilities
    downloader/               # Concurrent chunked file download
    filenamify/               # Filename sanitization
    files/                    # File existence checks
    logger/                   # Logrus-based logging to file
    m3u8/                     # M3U8 playlist parsing, TS decryption
```

### Key Flows

**Authentication**: Cookies (`gcid`/`gcess`) are read from `--config` YAML, CLI flags, or a persistent file in `os.UserConfigDir()/geektime-downloader/`. If no cookies exist, prompts for password and calls `geektime.Login()`. Cookies are persisted after successful login.

**Two Operating Modes**:
1. **Interactive** (default): Uses `promptui` for terminal UI — select product type, enter course ID, choose articles.
2. **Non-interactive**: Activated when `--course-ids`, `--config`, or `--non-interactive` is provided.

**Video Download Pipeline**: Get article info → video ID → play auth → build Alibaba Cloud VOD API URL (RSA + HMAC-SHA1 signed) → fetch m3u8 → parse TS files → download segments concurrently → decrypt → merge.

**PDF Generation**: Uses `chromedp` to launch headless Chrome, navigate to article page, wait for content load, hide UI elements, then `page.PrintToPDF()`.

### Key Configuration
- **Default download folder**: `$HOME/geektime-downloader/` (Unix/macOS), `%USERPROFILE%\geektime-downloader\` (Windows)
- **Log file**: `os.UserConfigDir()/geektime-downloader/geektime-downloader.log`
- **YAML config format**: See `courses-example.yaml` — supports `courses` and `advanced_courses` sections with per-course overrides

### Important Patterns
- HTTP client wraps `resty` with cookie injection, 1 retry, 10s timeout, Geek Time-specific error codes (451=rate limit, 452=auth failure)
- Concurrency uses `golang.org/x/sync/errgroup`; default concurrency = half of CPU count
- `--output` is a bitmask: 1=PDF, 2=Markdown, 4=Audio (combine freely, e.g., 7=all)
- `--product-type` values: `normal`, `daily`, `openclass`, `qconplus`, `university`, `other`
