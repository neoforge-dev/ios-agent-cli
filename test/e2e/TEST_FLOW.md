# E2E Test Execution Flow

## Visual Test Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    E2E Test Suite Start                         │
│                   forge_terminal_test.sh                        │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
            ┌────────────────────────┐
            │  Prerequisites Check   │
            │  ────────────────────  │
            │  ✓ xcrun installed     │
            │  ✓ ios-agent built     │
            │  ✓ ForgeTerminal.app   │
            │  ✓ jq (optional)       │
            └────────┬───────────────┘
                     │ Pass
                     ▼
            ┌────────────────────────┐
            │   Setup Environment    │
            │  ────────────────────  │
            │  • Create temp dir     │
            │  • Setup trap          │
            │  • Initialize state    │
            └────────┬───────────────┘
                     │
                     ▼
        ┌────────────────────────────┐
        │   TEST 1: Device Discovery │
        │  ────────────────────────   │
        │  $ ios-agent devices        │
        │                             │
        │  Validates:                 │
        │  • JSON response            │
        │  • success = true           │
        │  • devices array exists     │
        │  • device_count > 0         │
        │                             │
        │  Exports: TEST_DEVICE_ID    │
        └────────┬───────────────────┘
                 │ Pass
                 ▼
        ┌────────────────────────────┐
        │   TEST 2: Simulator Boot   │
        │  ────────────────────────   │
        │  $ xcrun simctl boot        │
        │                             │
        │  Validates:                 │
        │  • State = "Booted"         │
        │  • Boot time < 120s         │
        │  • No errors                │
        └────────┬───────────────────┘
                 │ Pass
                 ▼
        ┌────────────────────────────┐
        │   TEST 3: App Installation │
        │  ────────────────────────   │
        │  $ xcrun simctl install     │
        │                             │
        │  Validates:                 │
        │  • Install succeeds         │
        │  • App container exists     │
        │  • Bundle ID registered     │
        └────────┬───────────────────┘
                 │ Pass
                 ▼
        ┌────────────────────────────┐
        │    TEST 4: App Launch      │
        │  ────────────────────────   │
        │  $ xcrun simctl launch      │
        │                             │
        │  Validates:                 │
        │  • Launch succeeds          │
        │  • App process starts       │
        │  • Wait 3s for startup      │
        └────────┬───────────────────┘
                 │ Pass
                 ▼
        ┌────────────────────────────┐
        │  TEST 5: Screenshot Capture│
        │  ────────────────────────   │
        │  $ ios-agent screenshot     │
        │    --device <id>            │
        │    --output <path>          │
        │                             │
        │  Validates:                 │
        │  • JSON success = true      │
        │  • File exists              │
        │  • File size > 1KB          │
        │  • Valid PNG format         │
        └────────┬───────────────────┘
                 │ Pass
                 ▼
        ┌────────────────────────────┐
        │   TEST 6: App Termination  │
        │  ────────────────────────   │
        │  $ xcrun simctl terminate   │
        │                             │
        │  Validates:                 │
        │  • Terminate succeeds       │
        │  • App process stops        │
        └────────┬───────────────────┘
                 │ Pass
                 ▼
            ┌────────────────────────┐
            │  Cleanup & Reporting   │
            │  ────────────────────   │
            │  • Terminate app       │
            │  • Archive artifacts   │
            │  • Output JSON report  │
            └────────┬───────────────┘
                     │
                     ▼
            ┌────────────────────────┐
            │    Test Summary        │
            │  ────────────────────   │
            │  Total: 6              │
            │  Passed: 6             │
            │  Failed: 0             │
            │  Success Rate: 100%    │
            └────────────────────────┘
```

## Error Flow

```
Any Test Failure
       │
       ▼
┌──────────────────┐
│  Log Error       │
│  • Capture output│
│  • Mark test     │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Continue Tests  │
│  (no abort)      │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Cleanup Phase   │
│  • Archive logs  │
│  • Save artifacts│
│  • Exit code 1   │
└──────────────────┘
```

## Data Flow

```
┌─────────────────┐
│  Test Script    │
│  Start          │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────┐
│  Environment Variables          │
│  ─────────────────────────────  │
│  • TEST_DEVICE_ID               │
│  • TEST_OUTPUT_DIR              │
│  • TEST_FAILED (boolean)        │
└────────┬────────────────────────┘
         │
         ├────────────────────────────────┐
         │                                │
         ▼                                ▼
┌─────────────────┐          ┌───────────────────┐
│  Test Functions │          │  Output Files     │
│  ─────────────  │          │  ───────────────  │
│  • Setup        │          │  • devices.json   │
│  • Test 1-6     │──────────▶  • screenshot.png │
│  • Cleanup      │          │  • responses.json │
└────────┬────────┘          │  • test.log       │
         │                   └───────────────────┘
         ▼
┌─────────────────┐
│  JSON Results   │
│  ─────────────  │
│  stdout         │
└─────────────────┘
```

## State Management

```
Test Execution Timeline:

T0  ├─ Initialize
    │  • Create temp directory
    │  • Set trap for cleanup
    │
T1  ├─ Device Discovery
    │  • Query devices
    │  • Export TEST_DEVICE_ID ─────────────┐
    │                                        │
T2  ├─ Simulator Boot                       │
    │  • Use TEST_DEVICE_ID ◀───────────────┤
    │  • Wait for boot                      │
    │                                        │
T3  ├─ App Install                          │
    │  • Use TEST_DEVICE_ID ◀───────────────┤
    │  • Install ForgeTerminal              │
    │                                        │
T4  ├─ App Launch                           │
    │  • Use TEST_DEVICE_ID ◀───────────────┤
    │  • Start app                          │
    │                                        │
T5  ├─ Screenshot                           │
    │  • Use TEST_DEVICE_ID ◀───────────────┤
    │  • Capture screen                     │
    │                                        │
T6  ├─ App Terminate                        │
    │  • Use TEST_DEVICE_ID ◀───────────────┤
    │  • Stop app                           │
    │                                        │
T7  └─ Cleanup
       • Terminate app (if running)
       • Archive artifacts (if failed)
       • Output final report
```

## JSON Response Flow

```
ios-agent devices
       │
       ▼
┌─────────────────────────────────┐
│  {                              │
│    "success": true,             │
│    "action": "devices.list",    │
│    "result": {                  │
│      "devices": [...]           │
│    },                           │
│    "timestamp": "..."           │
│  }                              │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│  Validate JSON                  │
│  • Check success = true         │
│  • Extract devices array        │
│  • Count devices                │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│  Select Device                  │
│  • Prefer "Booted" state        │
│  • Fall back to available       │
│  • Export DEVICE_ID             │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│  Return Test Result             │
│  {                              │
│    "test": "device_discovery",  │
│    "passed": true,              │
│    "message": "...",            │
│    "details": {...}             │
│  }                              │
└─────────────────────────────────┘
```

## File System Layout During Test

```
/Users/bogdan/work/FORGE/neoforge-dev/ios-agent-cli/
├── ios-agent                           ← Binary used by tests
├── test/
│   ├── e2e/
│   │   └── forge_terminal_test.sh     ← Main test script
│   └── fixtures/
│       └── expected-device-schema.json
│
/tmp/ios-agent-test-12345/              ← Created during test
├── start_time.txt
├── devices.json                        ← Device discovery output
├── screenshot-response.json
├── forge-terminal-screenshot.png       ← Captured screenshot
└── test.log                           ← Test execution log

/Users/bogdan/Library/Developer/Xcode/DerivedData/
└── ForgeTerminal-.../
    └── Build/Products/Debug-iphonesimulator/
        └── ForgeTerminal.app          ← Target app
```

## Parallel Test Execution (Future)

```
┌─────────────┐
│  Test Suite │
└──────┬──────┘
       │
       ├───────────────┬───────────────┬───────────────┐
       │               │               │               │
       ▼               ▼               ▼               ▼
  ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
  │ Device 1│    │ Device 2│    │ Device 3│    │ Device 4│
  ├─────────┤    ├─────────┤    ├─────────┤    ├─────────┤
  │ Test 1-6│    │ Test 1-6│    │ Test 1-6│    │ Test 1-6│
  └────┬────┘    └────┬────┘    └────┬────┘    └────┬────┘
       │              │              │              │
       └──────────────┴──────────────┴──────────────┘
                      │
                      ▼
              ┌───────────────┐
              │ Merge Results │
              └───────────────┘
```

## Integration Points

```
┌──────────────────┐
│   Developer      │
└────────┬─────────┘
         │
         │ $ make e2e-test
         │
         ▼
┌──────────────────┐
│   Makefile       │
│   e2e-test target│
└────────┬─────────┘
         │
         │ 1. Build ios-agent
         │ 2. Copy to project root
         │ 3. Execute test script
         │
         ▼
┌──────────────────┐
│  Test Script     │
│  forge_terminal_ │
│  test.sh         │
└────────┬─────────┘
         │
         │ Use ios-agent & xcrun
         │
         ▼
┌──────────────────┐
│  iOS Simulator   │
│  + ForgeTerminal │
└────────┬─────────┘
         │
         │ Return results
         │
         ▼
┌──────────────────┐
│  JSON Output     │
│  (stdout)        │
└──────────────────┘
```

## Test Expansion Pattern

```
New Feature Implemented in ios-agent
         │
         ▼
┌────────────────────────────────┐
│  1. Add Test Function          │
│     test_new_feature()         │
└────────┬───────────────────────┘
         │
         ▼
┌────────────────────────────────┐
│  2. Follow Pattern             │
│     • Check prerequisites      │
│     • Execute command          │
│     • Validate response        │
│     • Return JSON result       │
└────────┬───────────────────────┘
         │
         ▼
┌────────────────────────────────┐
│  3. Add to Test Array          │
│     tests+=("test_new_feature")│
└────────┬───────────────────────┘
         │
         ▼
┌────────────────────────────────┐
│  4. Update Documentation       │
│     • README.md                │
│     • TEST_MANIFEST.md         │
└────────────────────────────────┘
```

## Summary

This test flow ensures:
- ✅ Sequential execution with state sharing
- ✅ Comprehensive validation at each step
- ✅ Graceful error handling
- ✅ Complete artifact capture
- ✅ Structured JSON reporting
- ✅ Easy expansion for new features
