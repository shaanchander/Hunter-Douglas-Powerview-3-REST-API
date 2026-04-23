# Motorized Blinds — Reverse Engineered Protocol

## Overview

`SET_POSITION` packets are 13 bytes. Each packet controls both the **shade** (top, privacy layer) and the **blind** (blackout layer) simultaneously — the two share a rail, so their positions are geometrically coupled.

Packet framing: 2-byte LE command opcode · 1-byte sequence counter · 1-byte payload length · payload.

---

## Packet Structure

```
F7 01 | 01 | 09 | SS SS | GG GG | VV | 80 00 | 80 00
[cmd 2B][seq][len] [shade] [gap] [vel] [pos2 ] [pos3 ]
```

| Bytes | Field | Type | Description |
|-------|-------|------|-------------|
| 0–1 | Command | LE uint16 | `0x01F7` = `SET_POSITION` |
| 2 | Sequence ID | uint8 | Incremented per command by caller |
| 3 | Data length | uint8 | `0x09` — payload byte count |
| 4–5 | Shade | LE uint16 | % of window covered by shade from top × 100 |
| 6–7 | Gap | LE uint16 | % of window open at bottom × 100 |
| 8 | Velocity | uint8 | Motor speed; `0x00` = default |
| 9–10 | pos2 | LE uint16 | Secondary axis (e.g. tilt); `0x8000` = unused |
| 11–12 | pos3 | LE uint16 | Tertiary axis; `0x8000` = unused |

### Position encoding

All position fields are LE uint16 where `raw = pct × 100`. Range: `0x0000` (0%) – `0x2710` (100%). Example: 50% → raw 5000 → bytes `88 13`.

### Window geometry

```
┌─────────────────┐  ← top of window
│      SHADE      │  shade% — drops from top rail
├─────────────────┤  ← shade bottom / blind top (physically attached)
│      BLIND      │  100 − shade% − gap%  (derived, no direct field)
├─────────────────┤  ← blind bottom edge
│       GAP       │  gap% — open, light passes through
└─────────────────┘  ← bottom of window
```

Constraint: `shade% + gap% ≤ 100%` (physically enforced by the shared rail).

---

## Captured Samples

All samples captured at seq=1, velocity=0, pos2/pos3=`0x8000`.

| # | Description | Packet | Shade | Gap | Blind |
|---|-------------|--------|-------|-----|-------|
| 1 | All closed | `F7010109 0000 0000 00 8000 8000` | 0% | 0% | 100% |
| 2 | Blind fully open | `F7010109 0000 1027 00 8000 8000` | 0% | 100% | 0% |
| 3 | Shade fully closed | `F7010109 1027 0000 00 8000 8000` | 100% | 0% | 0% |
| 4 | Shade 35%, blind closed | `F7010109 AC0D 0000 00 8000 8000` | 35% | 0% | 65% |
| 5 | Shade 20%, gap ~8% | `F7010109 D007 1C03 00 8000 8000` | 20% | 7.96% | 72.04% |
| 6 | No shade, blind 70% closed | `F7010109 0000 DD0B 00 8000 8000` | 0% | 30.37% | 69.63% |
| 7 | Shade ~68%, thin blind strip | `F7010109 6D1A DD0B 00 8000 8000` | 67.65% | 30.37% | 1.98% |
| 8 | Shade top half, blind bottom half | `F7010109 8813 0000 00 8000 8000` | 50% | 0% | 50% |

### Byte-level breakdown

```
        cmd   s   l  shade  gap   vel  pos2   pos3
1  F7 01  01  09  00 00  00 00  00  80 00  80 00   shade=0%     gap=0%
2  F7 01  01  09  00 00  10 27  00  80 00  80 00   shade=0%     gap=100%
3  F7 01  01  09  10 27  00 00  00  80 00  80 00   shade=100%   gap=0%
4  F7 01  01  09  AC 0D  00 00  00  80 00  80 00   shade=35%    gap=0%
5  F7 01  01  09  D0 07  1C 03  00  80 00  80 00   shade=20%    gap=7.96%
6  F7 01  01  09  00 00  DD 0B  00  80 00  80 00   shade=0%     gap=30.37%
7  F7 01  01  09  6D 1A  DD 0B  00  80 00  80 00   shade=67.65% gap=30.37%
8  F7 01  01  09  88 13  00 00  00  80 00  80 00   shade=50%    gap=0%
```
