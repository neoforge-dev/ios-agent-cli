# E2E Test Manifest

## Test Suite: forge_terminal_test.sh

**Target Application:** ForgeTerminal.app
**Bundle ID:** com.codeswiftr.forge-terminal
**Last Updated:** 2026-02-04

### Test Coverage Matrix

| Test ID | Test Name | ios-agent Command | Fallback | Status | Priority |
|---------|-----------|-------------------|----------|--------|----------|
| T1 | Device Discovery | `ios-agent devices` | - | ✅ Implemented | P0 |
| T2 | Simulator Boot | `ios-agent simulator boot` | `xcrun simctl boot` | ✅ Implemented | P0 |
| T3 | App Installation | Future: `ios-agent app install` | `xcrun simctl install` | ✅ Implemented | P0 |
| T4 | App Launch | Future: `ios-agent app launch` | `xcrun simctl launch` | ✅ Implemented | P1 |
| T5 | Screenshot Capture | `ios-agent screenshot` | - | ✅ Implemented | P0 |
| T6 | App Termination | Future: `ios-agent app terminate` | `xcrun simctl terminate` | ✅ Implemented | P1 |

### Test Details

#### T1: Device Discovery
**Purpose:** Verify device enumeration and JSON response format
**Command:** `ios-agent devices`
**Success Criteria:**
- JSON response with `success: true`
- At least one available device
- Valid device schema (see `fixtures/expected-device-schema.json`)
- Device has all required fields: id, name, state, type, os_version, udid, available

**Validation:**
```bash
ios-agent devices | jq -e '.success == true and (.result.devices | length > 0)'
```

---

#### T2: Simulator Boot
**Purpose:** Ensure simulator is running before tests
**Command:** `ios-agent simulator boot` (future) or `xcrun simctl boot`
**Success Criteria:**
- Simulator reaches "Booted" state within 120 seconds
- No boot errors or crashes

**Validation:**
```bash
xcrun simctl list devices -j | jq -r --arg id "$DEVICE_ID" \
  '.devices[] | .[] | select(.udid == $id) | .state' | grep -q "Booted"
```

---

#### T3: App Installation
**Purpose:** Verify app can be installed on simulator
**Command:** `xcrun simctl install` (ios-agent app install in future)
**Success Criteria:**
- App installs without errors
- App container exists and is accessible
- Bundle ID is registered

**Validation:**
```bash
xcrun simctl get_app_container "$DEVICE_ID" "$BUNDLE_ID" >/dev/null
```

---

#### T4: App Launch
**Purpose:** Verify app can be launched and starts successfully
**Command:** `xcrun simctl launch` (ios-agent app launch in future)
**Success Criteria:**
- Launch command succeeds
- App process starts
- No immediate crashes

**Validation:**
```bash
xcrun simctl launch "$DEVICE_ID" "$BUNDLE_ID"
```

---

#### T5: Screenshot Capture
**Purpose:** Test screenshot functionality - core feature for AI agents
**Command:** `ios-agent screenshot --device <id> --output <path> --format png`
**Success Criteria:**
- JSON response with `success: true`
- Screenshot file created at specified path
- File size > 1KB
- Valid PNG/JPEG format (verified via magic bytes)

**Validation:**
```bash
ios-agent screenshot --device "$DEVICE_ID" --output test.png | \
  jq -e '.success == true' && \
  [ -f test.png ] && [ $(stat -f%z test.png) -gt 1024 ]
```

---

#### T6: App Termination
**Purpose:** Clean teardown of app process
**Command:** `xcrun simctl terminate` (ios-agent app terminate in future)
**Success Criteria:**
- Termination command succeeds
- App process stops
- No zombie processes

**Validation:**
```bash
xcrun simctl terminate "$DEVICE_ID" "$BUNDLE_ID"
```

---

## Future Test Scenarios

### UI Interactions (Planned)
| Test ID | Test Name | ios-agent Command | Status | Priority |
|---------|-----------|-------------------|--------|----------|
| T7 | Tap Gesture | `ios-agent io tap --device <id> --x X --y Y` | 🔜 Planned | P1 |
| T8 | Text Input | `ios-agent io text --device <id> "text"` | 🔜 Planned | P1 |
| T9 | Swipe Gesture | `ios-agent io swipe --device <id> --start-x X1 ...` | 🔜 Planned | P2 |
| T10 | Button Press | `ios-agent io button --device <id> --button HOME` | 🔜 Planned | P2 |

### Advanced Features (Planned)
| Test ID | Test Name | Description | Status | Priority |
|---------|-----------|-------------|--------|----------|
| T11 | Device State | Verify `ios-agent state --device <id>` returns app state | 🔜 Planned | P1 |
| T12 | Remote Device | Test Tailscale remote device support | 🔜 Planned | P3 |
| T13 | App Uninstall | Test `ios-agent app uninstall` | 🔜 Planned | P2 |
| T14 | Simulator Shutdown | Test `ios-agent simulator shutdown` | 🔜 Planned | P2 |

---

## Test Execution Flow

```
┌─────────────────────────┐
│  1. Prerequisites Check │
│  - Xcode tools          │
│  - Simulators available │
│  - ForgeTerminal built  │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│  2. Device Discovery    │
│  - Find available device│
│  - Export DEVICE_ID     │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│  3. Simulator Boot      │
│  - Check state          │
│  - Boot if needed       │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│  4. App Installation    │
│  - Uninstall old        │
│  - Install fresh        │
│  - Verify               │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│  5. App Launch          │
│  - Launch app           │
│  - Wait 3s for startup  │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│  6. Screenshot Capture  │
│  - Capture PNG          │
│  - Verify file & format │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│  7. App Termination     │
│  - Graceful shutdown    │
│  - Verify stopped       │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│  8. Cleanup & Report    │
│  - Archive artifacts    │
│  - Output JSON results  │
└─────────────────────────┘
```

---

## Test Data & Fixtures

### Expected Device Schema
Location: `test/fixtures/expected-device-schema.json`

Defines the JSON schema for device discovery response. Used for validation in future automated schema testing.

### Reference Screenshots (Future)
Location: `test/fixtures/reference-screenshots/`

Will contain reference screenshots for image comparison testing.

---

## CI/CD Integration Status

| Platform | Status | Notes |
|----------|--------|-------|
| GitHub Actions | 🔜 Planned | Requires macOS runner |
| Local Dev | ✅ Ready | Run via `make e2e-test` |
| Pre-commit Hook | 🔜 Planned | Optional local validation |

---

## Metrics & KPIs

### Test Performance Targets
- Total suite execution: < 120 seconds
- Device discovery: < 2 seconds
- Simulator boot (cold): < 60 seconds
- App install: < 10 seconds
- Screenshot capture: < 3 seconds

### Coverage Targets
- Command coverage: 100% of implemented commands
- Error scenario coverage: 80%
- Edge case coverage: 60%

---

## Maintenance Schedule

- **Weekly:** Run full E2E suite on main branch
- **Per PR:** Run affected test subset
- **Post-release:** Full regression suite
- **Monthly:** Review and update test scenarios

---

## Known Limitations

1. **Single Device:** Tests run on one device at a time
2. **No Parallel Execution:** Sequential test execution only
3. **No Visual Assertions:** Screenshot capture only, no image comparison yet
4. **Local Only:** No remote device testing yet
5. **Simulator Only:** Physical device testing not implemented

---

## Change Log

| Date | Version | Changes |
|------|---------|---------|
| 2026-02-04 | 1.0 | Initial E2E test suite with 6 core tests |

---

## Contributing

To add new test scenarios:

1. Update this manifest with test details
2. Implement test function in `forge_terminal_test.sh`
3. Add to `tests` array in main()
4. Update README.md with usage examples
5. Test locally before submitting PR
