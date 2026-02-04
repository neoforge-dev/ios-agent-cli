# FORGE Harness Integration for ios-agent-cli

**Version:** 1.0
**Date:** 2026-02-04
**Status:** Implementation Guide

## Executive Summary

This document defines how to use the FORGE harness/flywheel for autonomous development of ios-agent-cli, a Go CLI tool for AI-agent iOS automation. The harness provides autonomous feature implementation through the Ralph Loop pattern, with meta-learning that compounds efficiency over time.

**Key Benefits:**
- Autonomous feature implementation via Claude Code SDK
- Quality-driven feature prioritization from tech debt scans
- Pattern learning for faster similar features
- Human approval gates for risky operations
- Self-improving decision thresholds

---

## 1. Harness Architecture Overview

### Core Components

```
┌─────────────────────────────────────────────────┐
│         FORGE Harness (Python)                  │
├─────────────────────────────────────────────────┤
│                                                 │
│  ┌──────────────┐    ┌──────────────┐          │
│  │  Flywheel    │───▶│ Ralph Loop   │          │
│  │  Orchestrator│    │  Harness     │          │
│  └──────────────┘    └──────────────┘          │
│         │                    │                  │
│         ▼                    ▼                  │
│  ┌──────────────┐    ┌──────────────┐          │
│  │   Quality    │    │  Decision    │          │
│  │   Scanner    │    │   Engine     │          │
│  └──────────────┘    └──────────────┘          │
│         │                    │                  │
│         ▼                    ▼                  │
│  ┌──────────────┐    ┌──────────────┐          │
│  │  Tech Debt   │    │  Approval    │          │
│  │  Findings    │    │   Queue      │          │
│  └──────────────┘    └──────────────┘          │
└─────────────────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────┐
│         ios-agent-cli (Go)                      │
├─────────────────────────────────────────────────┤
│  ✓ Project scaffold                             │
│  ✓ Device discovery                             │
│  ✓ Simulator lifecycle                          │
│  ✓ App management                               │
│  ✓ UI interactions                              │
│  ✓ Screenshot capture                           │
│  ✓ Error handling                               │
│  ✓ Tests (unit + integration)                   │
└─────────────────────────────────────────────────┘
```

### How It Works

1. **Quality Scanner** → Scans Go codebase for tech debt
2. **Feature Generator** → Converts findings to features.json entries
3. **Ralph Loop** → Iteratively implements features via Claude Code SDK
4. **Decision Engine** → Routes complex features to human review
5. **Approval Queue** → Collects risky operations for approval
6. **Pattern Learning** → Successful implementations become reusable patterns

---

## 2. Using Flywheel for iOS Agent CLI

### Setup

```bash
# Navigate to project
cd /Users/bogdan/work/FORGE/neoforge-dev/ios-agent-cli

# Ensure features.json exists (already present)
ls -la features.json

# Set environment (optional - graceful degradation)
export FORGE_ROOT=/Users/bogdan/work/FORGE
export FORGE_CODE_ATLAS_URL=http://localhost:8000    # Optional: RAG context
export FORGE_TECH_DILIGENCE_URL=http://localhost:8001  # Optional: Quality scan
```

### Basic Usage

```bash
# Full flywheel: scan → generate → implement
forge-harness flywheel run -d neoforge-dev -p ios-agent-cli

# Ralph Loop only (use existing features.json)
forge-harness flywheel loop -d neoforge-dev -p ios-agent-cli

# Dry run (plan without implementing)
forge-harness flywheel run -d neoforge-dev -p ios-agent-cli --dry-run

# Scan for tech debt → update features.json
forge-harness flywheel scan --domain neoforge-dev --project ios-agent-cli
```

### Programmatic Usage

```python
from pathlib import Path
from forge_harness import create_flywheel_loop, FlywheelConfig

# Create fully-wired Ralph loop
loop = create_flywheel_loop(
    domain="neoforge-dev",
    project="ios-agent-cli",
    features_path=Path("features.json"),
    config=FlywheelConfig(
        max_iterations=100,
        test_command="make test",  # Go test command
        priority_threshold="high",  # Only P0/P1 features
    ),
)

# Run autonomous implementation
result = await loop.run()

print(f"Features implemented: {result.features_completed}")
print(f"Features blocked: {result.features_blocked}")
```

---

## 3. Feature Mapping: features.json → Flywheel

### Current Feature Status

From `features.json`:
- **P0 Features (Critical)**: 9 features (IOS-001 done, 8 pending)
- **P1 Features (High)**: 5 features (all pending)
- **P2 Features (Medium)**: 2 features (remote support)

### Feature → Agent Mapping

The flywheel routes features to appropriate agents based on complexity and risk:

| Feature Epic | Feature IDs | Agent Strategy | Complexity |
|--------------|-------------|----------------|------------|
| **foundation** | IOS-001, IOS-015 | Backend Engineer | Low |
| **device-management** | IOS-002, IOS-003, IOS-004 | Backend Engineer | Medium |
| **app-management** | IOS-005, IOS-006, IOS-007, IOS-008 | Backend Engineer | Medium |
| **observation** | IOS-009, IOS-014 | Backend Engineer + Human Review | Medium-High |
| **ui-interactions** | IOS-010, IOS-011, IOS-012, IOS-013 | Backend Engineer + Human Review | High |
| **quality** | IOS-016 | Debug Detective + Backend Engineer | High |
| **remote** | IOS-017, IOS-018 | Backend Engineer + Human Review | Very High |

### Flywheel Decision Logic

```python
# The Decision Engine uses these signals:
{
    "feature_category": "device-management",  # From features.json
    "priority": "P0",                         # Critical → autonomous
    "estimated_tokens": 2000,                 # Low → autonomous
    "has_tests": True,                        # Yes → higher confidence
    "depends_on": [],                         # No blockers → proceed
    "similar_patterns": 3,                    # Found in Code Atlas → proceed
    "quality_score": 85,                      # High → autonomous
}

# Decision outcomes:
# - PROCEED: Autonomous implementation
# - HUMAN_REVIEW: Send to approval queue
# - BLOCK: Dependencies not met
```

---

## 4. Test Integration

### Go Test Command Integration

The harness runs Go tests after each feature implementation:

```bash
# Configure in FlywheelConfig
test_command = "make test"

# Harness executes (after feature implementation):
cd /path/to/ios-agent-cli && make test

# Parses output for pass/fail
# Updates features.json status:
# - PASSING → mark feature complete
# - FAILING → increment attempts, add error message
```

### Test Structure

```
ios-agent-cli/
├── Makefile
│   └── test: uv run pytest tests/ -v
├── tests/
│   ├── unit/
│   │   ├── test_device_manager.py
│   │   ├── test_xcrun_bridge.py
│   │   └── test_output_formatter.py
│   ├── integration/
│   │   ├── test_simulator_lifecycle.py
│   │   ├── test_app_management.py
│   │   └── test_ui_interactions.py
│   └── fixtures/
│       └── sample_device_list.json
```

### Test-Driven Feature Flow

```
1. Ralph Loop picks feature IOS-002 (Device discovery)
   └─ Status: pending

2. Orchestrator implements feature
   └─ Creates: pkg/device/manager.go, cmd/devices.go
   └─ Creates: tests/unit/test_device_manager.py
   └─ Updates: cmd/root.go

3. Harness runs: make test
   └─ Output: PASSED (all tests pass)

4. Harness updates features.json
   └─ IOS-002: status="passing", attempts=1

5. Ralph Loop picks next feature IOS-003
   └─ Blocked check: depends_on=["IOS-002"] ✓ (IOS-002 passing)
   └─ Proceed to implementation
```

---

## 5. Iteration Sequence for MVP

### Phase 1: Foundation + Device Management (Autonomous)

**Priority:** P0
**Estimated Iterations:** 20-30
**Approval Gates:** None (safe operations)

```bash
# Features: IOS-001 (done), IOS-002, IOS-003, IOS-004, IOS-015
forge-harness flywheel loop -d neoforge-dev -p ios-agent-cli \
  --max-iterations 30
```

**Expected Flow:**
1. IOS-002: Device discovery (local simulators) → 3-5 iterations
2. IOS-015: Error handling framework → 2-3 iterations
3. IOS-003: Simulator boot → 4-6 iterations (polls xcrun simctl)
4. IOS-004: Simulator shutdown → 2-3 iterations

**Success Criteria:**
- All P0 device management features passing
- 70%+ test coverage on pkg/device/, pkg/xcrun/
- Error codes standardized in pkg/output/error.go

---

### Phase 2: App Management (Semi-Autonomous)

**Priority:** P0 (IOS-005, IOS-006), P1 (IOS-007, IOS-008)
**Estimated Iterations:** 20-30
**Approval Gates:** IOS-007 (app install - filesystem write)

```bash
# Features: IOS-005, IOS-006, IOS-007, IOS-008
forge-harness flywheel loop -d neoforge-dev -p ios-agent-cli \
  --priority high \
  --max-iterations 30
```

**Expected Flow:**
1. IOS-005: App launch → 4-6 iterations (xcrun simctl launch + PID tracking)
2. IOS-006: App terminate → 2-3 iterations
3. **IOS-007: App install → HUMAN_REVIEW** (filesystem write risk)
   - Harness pauses, sends to approval queue
   - Human approves → implementation continues
4. IOS-008: App uninstall → 2-3 iterations

**Decision Engine Signals:**
```python
# IOS-007 triggers human review:
{
    "feature_id": "IOS-007",
    "risk_signals": ["filesystem_write", "ipa_processing"],
    "confidence": 0.62,  # Below threshold (0.70)
    "recommendation": "HUMAN_REVIEW",
    "reason": "Filesystem write operation in MVP - verify safety"
}
```

---

### Phase 3: UI Interactions + Observation (Semi-Autonomous)

**Priority:** P0 (IOS-009, IOS-010, IOS-011), P1 (IOS-012, IOS-013, IOS-014)
**Estimated Iterations:** 30-40
**Approval Gates:** IOS-010 (tap coordinates - validation critical)

```bash
# Features: IOS-009, IOS-010, IOS-011, IOS-012, IOS-013, IOS-014
forge-harness flywheel loop -d neoforge-dev -p ios-agent-cli \
  --priority medium \
  --max-iterations 40
```

**Expected Flow:**
1. IOS-009: Screenshot → 3-5 iterations (mobilecli wrapper + metadata)
2. **IOS-010: Tap interaction → HUMAN_REVIEW**
   - Critical: coordinate validation to prevent off-screen taps
   - Approval → implementation with bounds checking
3. IOS-011: Text input → 2-3 iterations (depends on IOS-010)
4. IOS-012: Swipe interaction → 3-4 iterations
5. IOS-013: Button press → 2-3 iterations
6. IOS-014: State command → 4-6 iterations (complex JSON assembly)

**Success Criteria:**
- All P0 UI interactions working
- Screenshot returns valid PNG with metadata
- Tap/swipe validates coordinates within screen bounds
- Error handling for UI_ACTION_FAILED

---

### Phase 4: Quality + Integration (Human-Guided)

**Priority:** P1 (IOS-016)
**Estimated Iterations:** 15-20
**Approval Gates:** Test strategy review

```bash
# Feature: IOS-016 (integration tests)
forge-harness flywheel loop -d neoforge-dev -p ios-agent-cli \
  --features features_quality.json \
  --max-iterations 20
```

**Expected Flow:**
1. **IOS-016: Integration tests → HUMAN_REVIEW**
   - Harness generates test plan
   - Human reviews test strategy
   - Approval → implement tests
2. Test coverage verification
3. End-to-end workflow tests (boot → launch → tap → screenshot → terminate)

**Test Categories:**
- Unit: Device manager, xcrun bridge, output formatter
- Integration: Full simulator lifecycle, app management workflows
- End-to-End: Agent simulation (screenshot → tap → verify)

---

### Phase 5: Remote Support (Post-MVP)

**Priority:** P2 (IOS-017, IOS-018)
**Estimated Iterations:** 20-30
**Approval Gates:** Multiple (Tailscale, SSH, remote execution)

```bash
# Features: IOS-017, IOS-018
forge-harness flywheel loop -d neoforge-dev -p ios-agent-cli \
  --features features_remote.json \
  --max-iterations 30
```

**Defer to Phase 2+ of project:**
- IOS-017: Manual remote host support (--remote-host flag)
- IOS-018: Tailscale auto-discovery (mDNS + SSH)

---

## 6. Agent Delegation Strategy

### Autonomous Features (No Human Gate)

**Characteristics:**
- Low risk (read-only operations, local simulator only)
- Well-defined patterns (xcrun simctl wrappers)
- Testable with unit tests
- No external dependencies

**Features:**
- IOS-002: Device discovery (xcrun simctl list)
- IOS-003: Simulator boot (deterministic polling)
- IOS-004: Simulator shutdown
- IOS-006: App terminate
- IOS-011: Text input (safe IO operation)
- IOS-015: Error handling framework

**Agent:** Backend Engineer (via Claude Code SDK)

---

### Semi-Autonomous Features (Human Approval Gate)

**Characteristics:**
- Medium risk (write operations, coordinate validation)
- New patterns (first implementation of category)
- Integration with external tools (mobilecli)
- Complex error handling

**Features:**
- IOS-005: App launch (PID tracking complexity)
- IOS-007: App install (filesystem write)
- IOS-009: Screenshot (mobilecli integration)
- IOS-010: Tap interaction (coordinate validation critical)
- IOS-012: Swipe interaction
- IOS-013: Button press
- IOS-014: State command (complex JSON assembly)

**Agent:** Backend Engineer (autonomous) + Human Review (approval)

**Approval Workflow:**
```
1. Ralph Loop attempts feature IOS-007
2. Decision Engine classifies: HUMAN_REVIEW
3. Harness creates approval request:
   - Title: "Implement IOS-007: App install command"
   - Risk: Medium (filesystem write)
   - Context: Feature description, acceptance criteria
4. Approval request sent to queue (.forge_approvals/ios-007.json)
5. Human reviews via CLI or webhook:
   forge-harness approvals list
   forge-harness approvals approve ios-007
6. Harness resumes implementation
```

---

### Human-Guided Features (Full Review)

**Characteristics:**
- High risk (security, remote execution, test strategy)
- Architectural decisions required
- Multiple implementation approaches
- Quality gates

**Features:**
- IOS-016: Integration tests (test strategy review)
- IOS-017: Remote host support (SSH/network security)
- IOS-018: Tailscale discovery (auto-discovery logic)

**Agent:** Backend Engineer (implementation) + Human (design review)

---

## 7. Harness Configuration

### Project-Specific Config

Create `.forge/config.yml` in ios-agent-cli:

```yaml
project:
  domain: neoforge-dev
  name: ios-agent-cli
  language: go
  test_command: make test
  lint_command: make lint
  build_command: make build

ralph_loop:
  max_iterations: 100
  max_failures_per_feature: 3
  checkpoint_interval: 5
  timeout_seconds: 300

decision_engine:
  confidence_thresholds:
    proceed: 0.70      # Autonomous if confidence >= 0.70
    review: 0.40       # Human review if 0.40 <= confidence < 0.70
    block: 0.0         # Block if confidence < 0.40

  risk_categories:
    filesystem_write: 0.60   # Reduce confidence for file writes
    network_call: 0.50       # Reduce confidence for network ops
    external_tool: 0.65      # Reduce confidence for tool integration

approval_queue:
  enabled: true
  channels:
    - slack          # Slack notifications with approve buttons
    - cli            # CLI approval via forge-harness approvals
  timeout_hours: 24  # Auto-reject after 24h

quality:
  min_coverage: 70
  security_scanners:
    - gosec         # Go security scanner
    - staticcheck   # Go static analysis

feedback_loops:
  code_atlas:
    enabled: true
    index_after_success: true  # Index successful implementations

  session_notes:
    enabled: true
    daily_log: .forge_learning/daily/{date}.md
```

---

## 8. Monitoring & Observability

### Dashboard

```bash
# Live TUI dashboard
forge-harness dashboard --live

# Panels:
# - Active Pipelines: Ralph Loop progress
# - Pending Approvals: Human gates
# - Ralph Loop: Current feature, iteration, failures
# - Recent Errors: Last 5 errors
```

### Daily Notes

```bash
# View today's session notes
forge-harness notes today

# Events logged:
# - Feature completions (IOS-002 ✓)
# - Feature failures (IOS-010 failed: coordinate validation)
# - Approval decisions (IOS-007 approved by human)
# - Threshold adjustments (confidence threshold raised to 0.75)
```

### Approval Queue

```bash
# List pending approvals
forge-harness approvals list

# Approve a feature
forge-harness approvals approve IOS-007 --comment "Reviewed implementation plan"

# Reject with feedback
forge-harness approvals reject IOS-010 \
  --reason "Add stricter coordinate validation"
```

---

## 9. Compounding Benefits

### Pattern Learning (Code Atlas)

```
Session 1: Implement IOS-002 (device discovery via xcrun simctl)
           → Pattern indexed: "xcrun simctl list --json parsing"

Session 2: Implement IOS-003 (simulator boot)
           → Query: "xcrun simctl boot polling pattern"
           → Found: IOS-002 pattern → Higher confidence → PROCEED
           → Implementation 2x faster
```

### Quality Improvement

```
Week 1: Quality score 65% → Generate 8 debt features
Week 2: Ralph Loop fixes debt → Quality score 78%
Week 3: Fewer approval gates (less debt) → Faster iteration
```

### Threshold Optimization

```
Initial: 70% autonomous, 30% human review
Week 2:  85% autonomous, 15% human review (thresholds optimized)
Week 4:  92% autonomous, 8% human review (learned patterns)
```

### Self-Improvement

```bash
# Harness scans itself for improvements
forge-harness quality self-analyze

# Generates features to improve harness
# Ralph Loop implements improvements
# Next session: Better harness → Faster development
```

---

## 10. Quick Reference

### Common Commands

```bash
# Full flywheel (scan + implement)
forge-harness flywheel run -d neoforge-dev -p ios-agent-cli

# Ralph Loop only
forge-harness flywheel loop -d neoforge-dev -p ios-agent-cli

# Dry run (planning only)
forge-harness flywheel run -d neoforge-dev -p ios-agent-cli --dry-run

# Scan for tech debt
forge-harness flywheel scan --domain neoforge-dev --project ios-agent-cli

# Monitor progress
forge-harness dashboard --live

# Manage approvals
forge-harness approvals list
forge-harness approvals approve <REQUEST_ID>

# View session notes
forge-harness notes today
```

### File Locations

| Path | Purpose |
|------|---------|
| `features.json` | Feature backlog (managed by harness) |
| `.forge/config.yml` | Project-specific harness config |
| `.forge_approvals/` | Approval requests (JSON) |
| `.ralph_checkpoints/` | Ralph Loop state (resume on crash) |
| `.forge_learning/daily/` | Daily session notes |
| `.forge_sessions/` | Session state tracking |

### Environment Variables

| Variable | Purpose | Required |
|----------|---------|----------|
| `FORGE_ROOT` | FORGE repository root | No (auto-detect) |
| `FORGE_CODE_ATLAS_URL` | Code Atlas RAG service | No (graceful degradation) |
| `FORGE_TECH_DILIGENCE_URL` | Quality scan service | No (local fallback) |
| `SLACK_WEBHOOK_URL` | Slack notifications | No (CLI only) |

---

## 11. Success Metrics

### MVP Completion

| Metric | Target | Current |
|--------|--------|---------|
| P0 Features Complete | 100% | 11% (1/9) |
| P1 Features Complete | 80% | 0% (0/5) |
| Test Coverage | 70% | TBD |
| Quality Score | 85% | TBD |
| Human Review Rate | <30% | TBD |

### Compounding Indicators

| Indicator | Week 1 | Week 2 | Week 4 |
|-----------|--------|--------|--------|
| Features/session | 1-2 | 2-3 | 3-5 |
| Human review % | 30% | 20% | 10% |
| Pattern reuse | 0% | 15% | 35% |
| Decision accuracy | 80% | 85% | 90% |

---

## 12. Next Steps

### Immediate (This Session)

1. ✅ Create harness integration document (this file)
2. Run flywheel scan to baseline quality
3. Start Phase 1: Foundation + Device Management
4. Monitor Ralph Loop progress via dashboard

### Short-Term (Week 1)

1. Complete Phase 1 (P0 device management)
2. Review approval queue patterns
3. Adjust confidence thresholds based on outcomes
4. Index successful patterns to Code Atlas

### Medium-Term (Week 2-3)

1. Complete Phase 2 (App management)
2. Complete Phase 3 (UI interactions + observation)
3. Implement Phase 4 (Integration tests)
4. Document learned patterns

### Long-Term (Week 4+)

1. Defer Phase 5 (Remote support) to post-MVP
2. Self-improvement loop on harness
3. Portfolio-wide pattern sharing
4. Threshold optimization complete

---

## Appendix: Harness Commands Cheat Sheet

```bash
# === Flywheel Operations ===
forge-harness flywheel run -d DOMAIN -p PROJECT           # Full cycle
forge-harness flywheel loop -d DOMAIN -p PROJECT          # Ralph Loop only
forge-harness flywheel scan [--domain D] [--project P]    # Tech debt scan

# === Monitoring ===
forge-harness dashboard                                   # Portfolio view
forge-harness dashboard --live                            # Live TUI
forge-harness notes today                                 # Daily session notes

# === Approvals ===
forge-harness approvals list                              # Pending approvals
forge-harness approvals approve REQUEST_ID [--comment C]  # Approve
forge-harness approvals reject REQUEST_ID --reason R      # Reject

# === Quality ===
forge-harness quality generate-features -d D -p P         # Scan → features.json
forge-harness quality self-analyze                        # Harness self-improvement

# === Utilities ===
forge-harness pipeline status CHECKPOINT_ID               # Check pipeline
forge-harness pipeline resume CHECKPOINT_ID               # Resume paused
```

---

**End of Document**
