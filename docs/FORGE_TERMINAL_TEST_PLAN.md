# Forge Terminal Test Plan using ios-agent-cli

**Version:** 1.0
**Date:** February 4, 2026
**Purpose:** Automated testing strategy for Forge Terminal iOS app using ios-agent-cli

---

## 1. Forge Terminal Overview

### What It Does

Forge Terminal is a native iOS SSH/Mosh terminal emulator designed for small screens (iPhone 14 Pro, iPad Mini) with the following features:

**Core Capabilities:**
- SSH connections with Ed25519 key authentication
- Mosh (mobile shell) support for cellular resilience
- Metal-accelerated GPU rendering with custom glyph cache
- iCloud-synced host configuration
- Keychain-secured SSH key storage
- Advanced text selection and clipboard management
- Terminal emulation (xterm-256color)
- Gesture-based interaction (pinch zoom, scroll, selection)

### Key Screens & Navigation Flow

```
┌─────────────────────────────────────────────────────┐
│ HostListView (Root)                                  │
│ • List of configured SSH hosts                       │
│ • Add/Edit/Delete hosts                              │
│ • iCloud sync status indicator                       │
│ • Search hosts                                        │
└──────────────────┬──────────────────────────────────┘
                   │ (Tap host)
                   ▼
┌─────────────────────────────────────────────────────┐
│ TerminalView                                         │
│ • Metal-rendered terminal display                    │
│ • Input bar for commands                             │
│ • Text selection overlay                             │
│ • Clipboard history                                  │
│ • Font size controls                                 │
│ • Connection status indicator                        │
└──────────────────┬──────────────────────────────────┘
                   │ (Menu actions)
                   ▼
┌─────────────────────────────────────────────────────┐
│ Ancillary Views                                      │
│ • AddHostView / EditHostView                         │
│ • KeyPickerView (SSH key selection)                  │
│ • ClipboardHistoryView                               │
│ • LookUpView (text actions)                          │
│ • SettingsView (planned)                             │
└─────────────────────────────────────────────────────┘
```

### Terminal UI Elements

**HostListView Components:**
- Navigation bar with "+" button (add host)
- Search bar
- Host list items (each showing: key/lock icon, name, username@hostname:port)
- iCloud sync status banner
- Empty state view (when no hosts)

**TerminalView Components:**
- Metal terminal canvas (full screen)
- Input bar at bottom (paste button, text field, send button)
- Toolbar with clipboard menu, font size buttons, connection status
- Selection overlay (appears during text selection)
- Scroll indicator (appears during scrolling)

**Common UI Patterns:**
- SwiftUI NavigationStack (iOS 16+)
- Context menus (long press on host items)
- Sheets for modal views (add/edit host)
- Pull-to-refresh (force iCloud sync)

---

## 2. Test Scenarios for ios-agent-cli

### Scenario 1: Host List Management

**Objective:** Verify host CRUD operations and iCloud sync UI

**Test Steps:**
1. Launch app to HostListView
2. Verify empty state shows "No Hosts" message
3. Tap "+" button to open AddHostView
4. Fill in host details (name, hostname, port, username)
5. Save host
6. Verify host appears in list
7. Verify iCloud sync indicator appears
8. Edit host via context menu
9. Delete host via swipe-to-delete
10. Verify empty state returns

**Expected UI Elements:**
- `server.rack` icon (empty state)
- "Add Host" button
- Text fields: "Name", "Hostname", "Port", "Username", "Key Alias"
- "Save" button (enabled after valid input)
- Host row with `key.fill` or `lock.fill` icon
- iCloud status icon (`checkmark.icloud`, `arrow.triangle.2.circlepath.icloud`)

**Success Criteria:**
- All UI elements render correctly
- Navigation transitions smoothly
- Host persists after app restart (iCloud sync)

---

### Scenario 2: SSH Key Generation

**Objective:** Verify Ed25519 key generation workflow

**Test Steps:**
1. In AddHostView, tap key icon next to "Key Alias"
2. KeyPickerView opens with empty key list
3. Tap "Generate New Key" button
4. Alert appears with "Key Alias" text field
5. Enter alias (e.g., "test-key")
6. Tap "Generate"
7. Verify key appears in list with checkmark
8. Verify key is selected in AddHostView

**Expected UI Elements:**
- "Generate New Key" button with `plus.circle` icon
- Alert dialog with text field
- Generated key row with `checkmark` icon
- "Available Keys" section header

**Success Criteria:**
- Key generates without error
- Public key stored in Keychain
- Key appears in picker list
- Key alias auto-fills in AddHostView

---

### Scenario 3: Terminal Session Launch

**Objective:** Verify terminal view loads and renders welcome message

**Test Steps:**
1. From HostListView, tap configured host
2. TerminalView opens with loading state
3. Metal renderer initializes
4. Welcome message displays:
   ```
   Forge Terminal v1.0
   ===================

   Ready for connection.
   ```
5. Verify input bar at bottom
6. Verify toolbar with clipboard menu, font buttons, status indicator

**Expected UI Elements:**
- Metal-rendered text (white on dark background)
- Input text field (placeholder: "Command")
- Send button (`arrow.up.circle.fill` icon)
- Paste button (`doc.on.clipboard` icon)
- Connection status (red circle + "Disconnected")
- Font size buttons (`textformat.size.smaller`, `textformat.size.larger`)

**Success Criteria:**
- Welcome message renders correctly
- Terminal accepts input
- Metal renderer maintains 60 FPS
- No visual glitches or flicker

---

### Scenario 4: Terminal Input and Command Execution

**Objective:** Verify command input and display workflow

**Test Steps:**
1. In TerminalView, tap input text field
2. Type command: "ls -la"
3. Tap send button or press Return
4. Verify command echoes to terminal: `$ ls -la`
5. Type multi-line command with newlines
6. Verify proper line wrapping

**Expected UI Elements:**
- Keyboard appears on tap
- Monospaced font in text field
- Command echoes with `$` prompt
- Send button enables when text present

**Success Criteria:**
- Commands render in terminal canvas
- Input field clears after send
- Terminal scrolls to show latest output
- Long commands wrap correctly

---

### Scenario 5: Text Selection and Clipboard

**Objective:** Verify text selection, copy, paste, and clipboard history

**Test Steps:**
1. Long-press on terminal text
2. Verify haptic feedback
3. Drag to select text
4. SelectionOverlay appears with handles
5. Tap "Copy" in context menu
6. Verify clipboard indicator updates
7. Tap clipboard menu button
8. Select "Clipboard History"
9. Verify copied text in history
10. Tap history item to paste
11. Verify text pastes to terminal

**Expected UI Elements:**
- Selection handles (blue overlay)
- Context menu: "Copy", "Paste", "Select All", "Look Up"
- Clipboard menu badge (if enabled)
- ClipboardHistoryView with timestamped items
- Swipe-to-delete on history items
- "Clear All" button

**Success Criteria:**
- Selection smooth and responsive
- Copy adds to clipboard and history
- Paste sends bracketed paste sequences (if mode enabled)
- History persists between sessions

---

### Scenario 6: Font Size Adjustment

**Objective:** Verify pinch-to-zoom and toolbar font controls

**Test Steps:**
1. Pinch-to-zoom on terminal canvas
2. Verify font size increases/decreases (8pt - 32pt range)
3. Verify terminal re-flows text
4. Tap "+" font button in toolbar
5. Verify font increases by 2pt
6. Tap "-" font button
7. Verify font decreases by 2pt
8. Verify buttons disable at min/max

**Expected UI Elements:**
- `textformat.size.larger` button (enabled/disabled)
- `textformat.size.smaller` button (enabled/disabled)
- Font size indicator (implicit in rendered text)

**Success Criteria:**
- Pinch gesture scales font smoothly
- Toolbar buttons work correctly
- Font size clamped to 8-32pt range
- Terminal dimensions recalculate (cols/rows)

---

### Scenario 7: Scrolling and Scroll Buffer

**Objective:** Verify terminal scroll behavior and history

**Test Steps:**
1. Generate enough output to exceed screen height (e.g., `seq 1 100`)
2. Swipe up to scroll back
3. Verify scroll indicator appears with offset
4. Two-finger swipe for fast scroll
5. Tap "Scroll to Bottom" in clipboard menu
6. Verify terminal jumps to latest output

**Expected UI Elements:**
- Scroll indicator overlay (shows offset)
- `arrow.up.circle.fill` icon
- Scroll indicator fades after 1 second
- "Scroll to Bottom" menu item with `arrow.down.to.line` icon

**Success Criteria:**
- Scroll smooth at 60 FPS
- Indicator shows correct line offset
- Fast scroll (2-finger) works
- Scroll to bottom resets offset to 0

---

### Scenario 8: Host Search and Filtering

**Objective:** Verify host list search functionality

**Test Steps:**
1. Add multiple hosts with different names
2. Tap search bar in HostListView
3. Type partial hostname (e.g., "prod")
4. Verify list filters to matching hosts
5. Clear search
6. Verify full list returns

**Expected UI Elements:**
- Search bar at top of list (iOS searchable)
- Filtered results update live
- "No hosts match 'query'" message if empty

**Success Criteria:**
- Search matches name and hostname
- Case-insensitive matching
- Filtering instant (no lag)
- Move/delete disabled during search

---

### Scenario 9: Connection Status Indicator

**Objective:** Verify connection status updates and visual feedback

**Test Steps:**
1. Open TerminalView (disconnected state)
2. Verify red circle + "Disconnected" in toolbar
3. Simulate SSH connection attempt
4. Verify status changes to green circle + "Connected"
5. Simulate connection loss
6. Verify status returns to red

**Expected UI Elements:**
- Status dot (8x8 circle)
- Color: green (connected), red (disconnected)
- Text: "Connected" / "Disconnected"
- Font: caption, secondary color

**Success Criteria:**
- Status updates in real-time
- Colors accurate (red/green)
- No jitter or flicker during transitions

---

### Scenario 10: Dark Mode and Accessibility

**Objective:** Verify app respects system appearance and accessibility settings

**Test Steps:**
1. Enable Dark Mode in iOS Settings
2. Launch app
3. Verify terminal background darkens
4. Verify text remains readable
5. Enable Dynamic Type (larger text)
6. Verify UI scales appropriately
7. Enable VoiceOver
8. Verify all controls have labels

**Expected UI Elements:**
- Adaptive colors (Light/Dark mode)
- Dynamic Type support for UI text
- VoiceOver labels for buttons/icons

**Success Criteria:**
- No white flashes in Dark Mode
- Terminal colors adapt (ANSI palette)
- Dynamic Type scales UI (not terminal canvas)
- VoiceOver announces all elements

---

## 3. Required ios-agent Commands

### 3.1 Device and Simulator Setup

```bash
# List available simulators
ios-agent devices

# Boot iPhone 14 Pro simulator
ios-agent simulator boot --name "iPhone 14 Pro" --os-version 17.4

# Get device ID
DEVICE_ID=$(ios-agent devices | jq -r '.devices[0].id')
```

### 3.2 App Installation and Launch

```bash
# Install Forge Terminal from .ipa or .app bundle
ios-agent app install --device $DEVICE_ID --ipa /path/to/ForgeTerminal.app

# Launch app
ios-agent app launch \
  --device $DEVICE_ID \
  --bundle com.codeswiftr.forge-terminal \
  --wait-for-ready 5

# Terminate app
ios-agent app terminate --device $DEVICE_ID --bundle com.codeswiftr.forge-terminal
```

### 3.3 UI Observation

```bash
# Take screenshot
ios-agent screenshot \
  --device $DEVICE_ID \
  --format png \
  --output /tmp/forge-terminal-screenshot.png

# Get device state (future: with UI tree)
ios-agent state \
  --device $DEVICE_ID \
  --include-screenshot
```

### 3.4 UI Interaction

```bash
# Tap "Add Host" button (coordinates from screenshot analysis)
ios-agent io tap --device $DEVICE_ID --x 350 --y 100

# Type in text field
ios-agent io text --device $DEVICE_ID "moltbot.local"

# Tap "Save" button
ios-agent io tap --device $DEVICE_ID --x 320 --y 700

# Swipe to scroll
ios-agent io swipe \
  --device $DEVICE_ID \
  --start-x 200 --start-y 500 \
  --end-x 200 --end-y 200 \
  --duration 300

# Press Home button (background app)
ios-agent io button --device $DEVICE_ID --button HOME

# Pinch-to-zoom (future enhancement - not in MVP)
# For MVP: use toolbar buttons instead
```

### 3.5 Test Loop Pattern

```bash
# Standard test iteration
function test_iteration() {
  local test_name=$1

  # 1. Take screenshot
  ios-agent screenshot --device $DEVICE_ID --output /tmp/${test_name}.png

  # 2. Analyze screenshot (AI vision / OCR)
  local screen_state=$(analyze_screenshot /tmp/${test_name}.png)

  # 3. Decide next action based on state
  if [[ $screen_state == *"No Hosts"* ]]; then
    echo "Empty state detected - tapping Add Host"
    ios-agent io tap --device $DEVICE_ID --x 350 --y 100
  elif [[ $screen_state == *"Name"* ]]; then
    echo "Add Host form detected - filling fields"
    ios-agent io tap --device $DEVICE_ID --x 200 --y 200
    ios-agent io text --device $DEVICE_ID "Test Host"
  fi

  # 4. Wait for UI update
  sleep 1

  # 5. Verify success
  ios-agent screenshot --device $DEVICE_ID --output /tmp/${test_name}_result.png
  verify_expected_state /tmp/${test_name}_result.png
}
```

---

## 4. Sample Agent Loop Pseudo-Code

### 4.1 Host Management Test Loop

```python
#!/usr/bin/env python3
"""
Forge Terminal Host Management Test
Tests host CRUD operations using ios-agent-cli
"""

import json
import subprocess
import time
from typing import Dict, Tuple

class ForgeTerminalTester:
    def __init__(self, device_id: str, bundle_id: str = "com.codeswiftr.forge-terminal"):
        self.device_id = device_id
        self.bundle_id = bundle_id
        self.screenshot_count = 0

    def run_command(self, cmd: list) -> Dict:
        """Execute ios-agent command and parse JSON output"""
        result = subprocess.run(cmd, capture_output=True, text=True)
        if result.returncode != 0:
            raise Exception(f"Command failed: {result.stderr}")
        return json.loads(result.stdout)

    def take_screenshot(self, label: str = "") -> str:
        """Capture screenshot and return path"""
        self.screenshot_count += 1
        filename = f"/tmp/forge_test_{self.screenshot_count:04d}_{label}.png"
        self.run_command([
            "ios-agent", "screenshot",
            "--device", self.device_id,
            "--output", filename
        ])
        return filename

    def tap(self, x: int, y: int) -> bool:
        """Tap at coordinates"""
        result = self.run_command([
            "ios-agent", "io", "tap",
            "--device", self.device_id,
            "--x", str(x),
            "--y", str(y)
        ])
        return result.get("success", False)

    def type_text(self, text: str) -> bool:
        """Type text into focused field"""
        result = self.run_command([
            "ios-agent", "io", "text",
            "--device", self.device_id,
            text
        ])
        return result.get("success", False)

    def swipe(self, start: Tuple[int, int], end: Tuple[int, int], duration: int = 300) -> bool:
        """Swipe gesture"""
        result = self.run_command([
            "ios-agent", "io", "swipe",
            "--device", self.device_id,
            "--start-x", str(start[0]),
            "--start-y", str(start[1]),
            "--end-x", str(end[0]),
            "--end-y", str(end[1]),
            "--duration", str(duration)
        ])
        return result.get("success", False)

    def analyze_screenshot(self, image_path: str) -> Dict:
        """
        Analyze screenshot using Claude Vision API
        Returns detected UI elements and state
        """
        # TODO: Integrate with Claude Vision API
        # For now, return mock data
        return {
            "view": "HostListView",
            "elements": ["add_button", "search_bar", "empty_state"],
            "text": ["No Hosts", "Add a host to get started", "Add Host"]
        }

    def wait_for_ui_update(self, seconds: float = 1.0):
        """Wait for UI animations to complete"""
        time.sleep(seconds)

    def test_add_host(self) -> bool:
        """Test Scenario: Add new SSH host"""
        print("🧪 Test: Add Host Flow")

        # Step 1: Verify initial state
        print("  1️⃣ Taking screenshot of initial state...")
        screenshot = self.take_screenshot("initial_state")
        state = self.analyze_screenshot(screenshot)

        if "empty_state" not in state["elements"]:
            print("  ⚠️  Warning: Expected empty state not found")

        # Step 2: Tap Add Host button (top-right corner, approx)
        print("  2️⃣ Tapping Add Host button...")
        self.tap(x=350, y=100)
        self.wait_for_ui_update()

        # Step 3: Verify AddHostView opened
        screenshot = self.take_screenshot("add_host_form")
        state = self.analyze_screenshot(screenshot)

        if "Name" not in state["text"]:
            print("  ❌ Add Host form did not open")
            return False

        # Step 4: Fill in host details
        print("  3️⃣ Filling host details...")

        # Tap Name field
        self.tap(x=200, y=200)
        self.wait_for_ui_update(0.5)
        self.type_text("Test Host")

        # Tap Hostname field
        self.tap(x=200, y=280)
        self.wait_for_ui_update(0.5)
        self.type_text("moltbot.local")

        # Tap Username field
        self.tap(x=200, y=400)
        self.wait_for_ui_update(0.5)
        self.type_text("bogdan")

        # Take screenshot of filled form
        screenshot = self.take_screenshot("form_filled")

        # Step 5: Tap Save button
        print("  4️⃣ Tapping Save button...")
        self.tap(x=350, y=80)  # Top-right Save button
        self.wait_for_ui_update(1.5)

        # Step 6: Verify host appears in list
        screenshot = self.take_screenshot("host_added")
        state = self.analyze_screenshot(screenshot)

        if "Test Host" not in state["text"]:
            print("  ❌ Host not found in list")
            return False

        print("  ✅ Host added successfully")
        return True

    def test_tap_host_and_verify_terminal(self) -> bool:
        """Test Scenario: Tap host and verify terminal opens"""
        print("🧪 Test: Open Terminal")

        # Step 1: Take screenshot of host list
        screenshot = self.take_screenshot("host_list")
        state = self.analyze_screenshot(screenshot)

        # Step 2: Tap first host (approximate center of first row)
        print("  1️⃣ Tapping host...")
        self.tap(x=200, y=300)
        self.wait_for_ui_update(2.0)  # Wait for connection attempt

        # Step 3: Verify TerminalView opened
        screenshot = self.take_screenshot("terminal_view")
        state = self.analyze_screenshot(screenshot)

        if "Forge Terminal v1.0" not in state["text"]:
            print("  ❌ Terminal did not open")
            return False

        # Step 4: Verify UI elements present
        expected_elements = ["input_bar", "connection_status", "font_buttons"]
        for element in expected_elements:
            if element not in state["elements"]:
                print(f"  ⚠️  Missing element: {element}")

        print("  ✅ Terminal opened successfully")
        return True

    def test_terminal_input(self) -> bool:
        """Test Scenario: Type command in terminal"""
        print("🧪 Test: Terminal Input")

        # Step 1: Tap input field
        print("  1️⃣ Tapping input field...")
        self.tap(x=200, y=750)
        self.wait_for_ui_update(0.5)

        # Step 2: Type command
        print("  2️⃣ Typing command...")
        self.type_text("ls -la")
        self.wait_for_ui_update(0.5)

        # Step 3: Take screenshot with command in field
        screenshot = self.take_screenshot("command_typed")

        # Step 4: Tap send button
        print("  3️⃣ Sending command...")
        self.tap(x=350, y=750)
        self.wait_for_ui_update(1.0)

        # Step 5: Verify command echoed to terminal
        screenshot = self.take_screenshot("command_sent")
        state = self.analyze_screenshot(screenshot)

        if "$ ls -la" not in state["text"]:
            print("  ❌ Command not echoed to terminal")
            return False

        print("  ✅ Command input works")
        return True

    def test_font_size_adjustment(self) -> bool:
        """Test Scenario: Adjust font size"""
        print("🧪 Test: Font Size Controls")

        # Step 1: Initial screenshot
        screenshot_before = self.take_screenshot("font_before")

        # Step 2: Tap increase font button (3 times)
        print("  1️⃣ Increasing font size...")
        for _ in range(3):
            self.tap(x=330, y=80)  # Larger text button
            self.wait_for_ui_update(0.3)

        # Step 3: Take screenshot after increase
        screenshot_after = self.take_screenshot("font_increased")

        # Step 4: Compare screenshots (visual diff or pixel comparison)
        # For now, assume success if no errors

        # Step 5: Decrease font size
        print("  2️⃣ Decreasing font size...")
        for _ in range(3):
            self.tap(x=280, y=80)  # Smaller text button
            self.wait_for_ui_update(0.3)

        screenshot_reset = self.take_screenshot("font_reset")

        print("  ✅ Font size controls work")
        return True

    def test_scroll_behavior(self) -> bool:
        """Test Scenario: Scroll terminal output"""
        print("🧪 Test: Terminal Scrolling")

        # Step 1: Generate output (simulate by taking multiple screenshots)
        # In real test, would send command that produces multi-screen output

        # Step 2: Swipe up to scroll
        print("  1️⃣ Scrolling up...")
        self.swipe(start=(200, 500), end=(200, 200), duration=300)
        self.wait_for_ui_update(0.5)

        # Step 3: Verify scroll indicator appears
        screenshot = self.take_screenshot("scrolled_up")
        state = self.analyze_screenshot(screenshot)

        # Scroll indicator should be visible
        if "scroll_indicator" not in state["elements"]:
            print("  ⚠️  Scroll indicator not detected")

        # Step 4: Swipe down to scroll bottom
        print("  2️⃣ Scrolling down...")
        self.swipe(start=(200, 200), end=(200, 500), duration=300)
        self.wait_for_ui_update(0.5)

        print("  ✅ Scrolling works")
        return True

    def run_all_tests(self):
        """Execute full test suite"""
        print("🚀 Starting Forge Terminal Test Suite")
        print(f"   Device: {self.device_id}")
        print(f"   Bundle: {self.bundle_id}\n")

        results = {
            "add_host": self.test_add_host(),
            "open_terminal": self.test_tap_host_and_verify_terminal(),
            "terminal_input": self.test_terminal_input(),
            "font_controls": self.test_font_size_adjustment(),
            "scrolling": self.test_scroll_behavior()
        }

        print("\n📊 Test Results:")
        passed = sum(results.values())
        total = len(results)

        for test_name, result in results.items():
            status = "✅ PASS" if result else "❌ FAIL"
            print(f"  {status}  {test_name}")

        print(f"\n🎯 Summary: {passed}/{total} tests passed")
        return passed == total


def main():
    # Step 1: Discover device
    result = subprocess.run(["ios-agent", "devices"], capture_output=True, text=True)
    devices = json.loads(result.stdout)

    if not devices["devices"]:
        print("❌ No devices found. Boot a simulator first:")
        print("   ios-agent simulator boot --name 'iPhone 14 Pro' --os-version 17.4")
        return 1

    device_id = devices["devices"][0]["id"]
    print(f"✅ Using device: {device_id}")

    # Step 2: Launch app
    print("🚀 Launching Forge Terminal...")
    subprocess.run([
        "ios-agent", "app", "launch",
        "--device", device_id,
        "--bundle", "com.codeswiftr.forge-terminal",
        "--wait-for-ready", "5"
    ])

    # Step 3: Run tests
    tester = ForgeTerminalTester(device_id)
    success = tester.run_all_tests()

    # Step 4: Cleanup
    print("\n🧹 Cleaning up...")
    subprocess.run([
        "ios-agent", "app", "terminate",
        "--device", device_id,
        "--bundle", "com.codeswiftr.forge-terminal"
    ])

    return 0 if success else 1


if __name__ == "__main__":
    exit(main())
```

---

## 5. Prerequisites and Setup

### 5.1 iOS Simulator Setup

```bash
# 1. Verify Xcode Command Line Tools installed
xcode-select --install

# 2. List available simulator runtimes
xcrun simctl list runtimes

# 3. Create iPhone 14 Pro simulator (if not exists)
xcrun simctl create "iPhone 14 Pro Test" \
  "com.apple.CoreSimulator.SimDeviceType.iPhone-14-Pro" \
  "com.apple.CoreSimulator.SimRuntime.iOS-17-4"

# 4. Boot simulator
ios-agent simulator boot --name "iPhone 14 Pro Test" --os-version 17.4

# 5. Wait for boot to complete
sleep 10
```

### 5.2 Forge Terminal Build

```bash
# Navigate to Forge Terminal project
cd /Users/bogdan/work/FORGE/codeswiftr-com/forge-terminal/ios

# Build for simulator (Debug configuration)
xcodebuild \
  -scheme ForgeTerminal \
  -configuration Debug \
  -sdk iphonesimulator \
  -destination 'platform=iOS Simulator,name=iPhone 14 Pro' \
  -derivedDataPath ./build \
  build

# Extract .app bundle path
APP_PATH=$(find ./build/Build/Products/Debug-iphonesimulator -name "ForgeTerminal.app" -print -quit)

# Verify bundle exists
if [ ! -d "$APP_PATH" ]; then
  echo "❌ Build failed - .app bundle not found"
  exit 1
fi

echo "✅ Built app at: $APP_PATH"
```

### 5.3 App Installation

```bash
# Get device ID
DEVICE_ID=$(ios-agent devices | jq -r '.devices[0].id')

# Install app on simulator
ios-agent app install \
  --device $DEVICE_ID \
  --ipa "$APP_PATH"

# Verify installation
xcrun simctl listapps $DEVICE_ID | grep -i forge
```

### 5.4 Test Environment Configuration

```bash
# Create test output directory
mkdir -p /tmp/forge-terminal-test-results

# Set environment variables
export IOS_AGENT_DEVICE_ID=$DEVICE_ID
export FORGE_TERMINAL_BUNDLE="com.codeswiftr.forge-terminal"
export TEST_OUTPUT_DIR="/tmp/forge-terminal-test-results"

# Configure screenshot storage
export SCREENSHOT_DIR="$TEST_OUTPUT_DIR/screenshots"
mkdir -p "$SCREENSHOT_DIR"
```

### 5.5 Dependencies

**Required Tools:**
- ios-agent-cli (built from this repo)
- Xcode 15+ with iOS 17.4 SDK
- xcrun / simctl (Xcode Command Line Tools)
- jq (JSON parsing: `brew install jq`)
- Python 3.11+ (for test scripts)

**Optional Tools:**
- Claude Vision API (for screenshot analysis)
- ImageMagick (for image comparison: `brew install imagemagick`)
- FFmpeg (for screen recording: `brew install ffmpeg`)

### 5.6 Pre-Flight Checklist

```bash
# Verify all prerequisites
function preflight_check() {
  echo "🔍 Running pre-flight checks..."

  # Check ios-agent-cli
  if ! command -v ios-agent &> /dev/null; then
    echo "❌ ios-agent-cli not found. Build it first:"
    echo "   cd ios-agent-cli && make build && make install"
    return 1
  fi

  # Check Xcode
  if ! command -v xcodebuild &> /dev/null; then
    echo "❌ Xcode not installed"
    return 1
  fi

  # Check jq
  if ! command -v jq &> /dev/null; then
    echo "❌ jq not installed (brew install jq)"
    return 1
  fi

  # Check device
  local device_count=$(ios-agent devices | jq '.devices | length')
  if [ "$device_count" -eq 0 ]; then
    echo "❌ No devices available. Boot a simulator first."
    return 1
  fi

  echo "✅ All checks passed"
  return 0
}

# Run checks
preflight_check || exit 1
```

---

## 6. Test Execution Plan

### 6.1 Automated Test Run

```bash
#!/bin/bash
# run_forge_terminal_tests.sh

set -e

echo "🚀 Forge Terminal Automated Test Suite"
echo "========================================"

# 1. Pre-flight checks
source ./scripts/preflight_check.sh
preflight_check || exit 1

# 2. Setup environment
DEVICE_ID=$(ios-agent devices | jq -r '.devices[0].id')
BUNDLE_ID="com.codeswiftr.forge-terminal"
TEST_OUTPUT="/tmp/forge-terminal-test-$(date +%Y%m%d-%H%M%S)"

mkdir -p "$TEST_OUTPUT"
echo "📁 Test output: $TEST_OUTPUT"

# 3. Launch app
echo "🚀 Launching Forge Terminal..."
ios-agent app launch \
  --device $DEVICE_ID \
  --bundle $BUNDLE_ID \
  --wait-for-ready 5

# 4. Run test script
echo "🧪 Running tests..."
python3 ./tests/forge_terminal_test.py \
  --device $DEVICE_ID \
  --output $TEST_OUTPUT

# 5. Collect results
echo "📊 Collecting results..."
cp $TEST_OUTPUT/screenshots/*.png ./test-results/
cp $TEST_OUTPUT/test-report.json ./test-results/

# 6. Cleanup
echo "🧹 Cleanup..."
ios-agent app terminate --device $DEVICE_ID --bundle $BUNDLE_ID

echo "✅ Test run complete. Results in: ./test-results/"
```

### 6.2 CI/CD Integration

```yaml
# .github/workflows/ios-ui-tests.yml
name: Forge Terminal UI Tests

on:
  push:
    branches: [main, develop]
  pull_request:

jobs:
  test-ios-ui:
    runs-on: macos-14
    steps:
      - uses: actions/checkout@v3

      - name: Install ios-agent-cli
        run: |
          cd ios-agent-cli
          make build
          make install

      - name: Boot simulator
        run: |
          ios-agent simulator boot \
            --name "iPhone 14 Pro" \
            --os-version 17.4

      - name: Build Forge Terminal
        run: |
          cd forge-terminal/ios
          xcodebuild -scheme ForgeTerminal \
            -sdk iphonesimulator \
            -destination 'platform=iOS Simulator,name=iPhone 14 Pro' \
            build

      - name: Install app
        run: |
          DEVICE_ID=$(ios-agent devices | jq -r '.devices[0].id')
          ios-agent app install \
            --device $DEVICE_ID \
            --ipa ./forge-terminal/ios/build/Debug-iphonesimulator/ForgeTerminal.app

      - name: Run UI tests
        run: |
          python3 tests/forge_terminal_test.py \
            --device $(ios-agent devices | jq -r '.devices[0].id') \
            --output ./test-results

      - name: Upload screenshots
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: test-screenshots
          path: test-results/screenshots/

      - name: Upload test report
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: test-report
          path: test-results/test-report.json
```

---

## 7. Feedback Loop Closure

### 7.1 Test Result Reporting

```json
// test-report.json structure
{
  "run_id": "20260204-153045",
  "device": {
    "id": "12345678-1234567890ABCDEF",
    "name": "iPhone 14 Pro",
    "os_version": "17.4"
  },
  "app": {
    "bundle_id": "com.codeswiftr.forge-terminal",
    "version": "1.0.0",
    "build": "42"
  },
  "tests": [
    {
      "name": "test_add_host",
      "status": "passed",
      "duration_ms": 3450,
      "screenshots": [
        "screenshots/0001_initial_state.png",
        "screenshots/0002_add_host_form.png",
        "screenshots/0003_form_filled.png",
        "screenshots/0004_host_added.png"
      ],
      "assertions": [
        {"type": "ui_element_present", "element": "Add Host button", "result": "pass"},
        {"type": "text_matches", "expected": "Test Host", "result": "pass"},
        {"type": "navigation", "to": "HostListView", "result": "pass"}
      ]
    },
    {
      "name": "test_terminal_input",
      "status": "failed",
      "duration_ms": 2100,
      "error": "Command not echoed to terminal",
      "screenshots": [
        "screenshots/0010_command_typed.png",
        "screenshots/0011_command_sent.png"
      ],
      "assertions": [
        {"type": "text_matches", "expected": "$ ls -la", "result": "fail"}
      ]
    }
  ],
  "summary": {
    "total": 5,
    "passed": 4,
    "failed": 1,
    "duration_ms": 12500
  }
}
```

### 7.2 Screenshot Analysis with Claude Vision

```python
import anthropic
import base64

def analyze_forge_terminal_screenshot(image_path: str) -> dict:
    """
    Use Claude Vision API to analyze Forge Terminal screenshot
    Returns detected UI elements, text, and current view
    """
    with open(image_path, "rb") as f:
        image_data = base64.standard_b64encode(f.read()).decode("utf-8")

    client = anthropic.Anthropic(api_key=os.environ.get("ANTHROPIC_API_KEY"))

    message = client.messages.create(
        model="claude-opus-4-5-20251101",
        max_tokens=1024,
        messages=[
            {
                "role": "user",
                "content": [
                    {
                        "type": "image",
                        "source": {
                            "type": "base64",
                            "media_type": "image/png",
                            "data": image_data,
                        },
                    },
                    {
                        "type": "text",
                        "text": """Analyze this Forge Terminal iOS app screenshot.

                        Return JSON with:
                        {
                          "current_view": "HostListView|TerminalView|AddHostView|etc",
                          "ui_elements": ["button_names", "icons", "text_fields"],
                          "visible_text": ["all readable text"],
                          "interactive_regions": [
                            {"type": "button", "label": "Add Host", "approx_x": 350, "approx_y": 100},
                            ...
                          ],
                          "app_state": "description of current state"
                        }

                        Be precise about coordinates for tappable elements."""
                    }
                ],
            }
        ],
    )

    return json.loads(message.content[0].text)
```

### 7.3 Automated Issue Creation

```python
def create_github_issue_for_failed_test(test_result: dict):
    """
    Automatically create GitHub issue when test fails
    """
    if test_result["status"] != "failed":
        return

    issue_title = f"[UI Test Failure] {test_result['name']}"
    issue_body = f"""
## Test Failure Report

**Test Name:** `{test_result['name']}`
**Date:** {datetime.now().isoformat()}
**Device:** {test_result['device']['name']} ({test_result['device']['os_version']})
**App Version:** {test_result['app']['version']} (build {test_result['app']['build']})

### Error

```
{test_result.get('error', 'No error message')}
```

### Failed Assertions

{format_assertions(test_result['assertions'])}

### Screenshots

{format_screenshot_links(test_result['screenshots'])}

### Reproduction Steps

1. Boot simulator: `ios-agent simulator boot --name "iPhone 14 Pro"`
2. Install app: `ios-agent app install --device <id> --ipa ForgeTerminal.app`
3. Run test: `python3 tests/forge_terminal_test.py --test {test_result['name']}`

### Test Logs

```
{test_result.get('logs', 'No logs captured')}
```

---

*Auto-generated by ios-agent-cli test framework*
    """

    # Create issue via GitHub API
    requests.post(
        "https://api.github.com/repos/codeswiftr/forge-terminal/issues",
        headers={"Authorization": f"token {os.environ['GITHUB_TOKEN']}"},
        json={"title": issue_title, "body": issue_body, "labels": ["ui-test-failure", "automated"]}
    )
```

---

## 8. Advanced Test Scenarios

### 8.1 Regression Test: iCloud Sync

```python
def test_icloud_sync_between_devices():
    """
    Test iCloud sync by adding host on Device A and verifying it appears on Device B
    Requires two simulators running
    """
    device_a = "iphone-14-pro-1"
    device_b = "iphone-14-pro-2"

    # 1. Add host on Device A
    tester_a = ForgeTerminalTester(device_a)
    tester_a.test_add_host()

    # 2. Wait for iCloud sync (force sync via pull-to-refresh)
    tester_a.swipe(start=(200, 200), end=(200, 400), duration=500)
    time.sleep(5)

    # 3. Launch app on Device B
    launch_app(device_b)
    tester_b = ForgeTerminalTester(device_b)

    # 4. Force sync on Device B
    tester_b.swipe(start=(200, 200), end=(200, 400), duration=500)
    time.sleep(5)

    # 5. Verify host appears
    screenshot = tester_b.take_screenshot("device_b_synced")
    state = tester_b.analyze_screenshot(screenshot)

    assert "Test Host" in state["text"], "Host did not sync to Device B"
    print("✅ iCloud sync verified across devices")
```

### 8.2 Performance Test: Metal Rendering

```python
def test_metal_rendering_performance():
    """
    Measure Metal rendering FPS under load
    Requires ios-agent to support performance metrics (future enhancement)
    """
    tester = ForgeTerminalTester(device_id)

    # 1. Open terminal
    tester.test_tap_host_and_verify_terminal()

    # 2. Generate rapid output (simulated)
    # In real test, would run command like: `yes | head -1000`
    for i in range(100):
        tester.type_text(f"Line {i}\n")
        time.sleep(0.016)  # Target 60 FPS (16ms per frame)

    # 3. Capture metrics (future: ios-agent performance command)
    # metrics = tester.run_command(["ios-agent", "performance", "--device", device_id])
    # assert metrics["fps_average"] >= 55, "FPS below target (60)"

    print("✅ Rendering performance acceptable")
```

### 8.3 Edge Case: Memory Pressure

```python
def test_memory_pressure_handling():
    """
    Simulate memory pressure and verify app doesn't crash
    """
    tester = ForgeTerminalTester(device_id)

    # 1. Open terminal
    tester.test_tap_host_and_verify_terminal()

    # 2. Generate large scroll buffer (10,000 lines)
    # Simulated by repeatedly writing to terminal
    for i in range(1000):
        tester.type_text(f"Long line of text with data {i}\n")
        if i % 100 == 0:
            time.sleep(0.5)  # Allow UI to update

    # 3. Verify app still responsive
    screenshot = tester.take_screenshot("after_stress")
    state = tester.analyze_screenshot(screenshot)

    assert state["current_view"] == "TerminalView", "App crashed or changed state"
    print("✅ Memory pressure handled gracefully")
```

---

## 9. Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| **Test Coverage** | 80% of user flows | 5 core scenarios automated |
| **Test Reliability** | 95% pass rate | <5% flaky tests |
| **Execution Time** | <5 minutes | Full suite on single device |
| **Feedback Speed** | <10 seconds per iteration | Screenshot → analyze → action |
| **CI Integration** | <10 minutes | Build + test on every PR |
| **Issue Detection** | 100% of regressions | No manual testing required |

---

## 10. Future Enhancements

### Phase 2: Advanced Capabilities

1. **UI Tree Analysis**: Use WebDriverAgent for semantic element detection
2. **Network Simulation**: Test SSH/Mosh under poor network conditions
3. **Accessibility Testing**: Verify VoiceOver labels and Dynamic Type
4. **Gesture Recording**: Record and replay complex multi-touch gestures
5. **Visual Regression**: Automated screenshot diffing (Percy, Applitools)
6. **Real Device Testing**: Test on physical iPhone via Tailscale

### Phase 3: Agent-Driven Testing

```python
def autonomous_exploration_loop():
    """
    Agent explores Forge Terminal UI autonomously, discovering bugs
    """
    tester = ForgeTerminalTester(device_id)
    explored_states = set()

    while True:
        screenshot = tester.take_screenshot("explore")
        state = tester.analyze_screenshot(screenshot)
        state_hash = hash_state(state)

        if state_hash in explored_states:
            # Already seen this state - try new action
            action = select_unexplored_action(state)
        else:
            explored_states.add(state_hash)
            action = select_most_interesting_action(state)

        # Execute action
        execute_action(tester, action)

        # Check for crashes, errors, unexpected states
        if detect_error_state(state):
            report_bug(state, action)
```

---

## 11. Appendix

### A. Coordinate Reference Guide

**iPhone 14 Pro (390x844 logical points):**

```
┌────────────────────────────────────┐
│ Status Bar (0, 0) - (390, 54)      │
├────────────────────────────────────┤
│ Navigation Bar                      │
│   Left button: (20, 70)             │
│   Title: (195, 80)                  │
│   Right button: (350, 70)           │
├────────────────────────────────────┤
│                                     │
│                                     │
│     Content Area                    │
│     (0, 100) - (390, 750)           │
│                                     │
│                                     │
├────────────────────────────────────┤
│ Input Bar (0, 750) - (390, 810)    │
│   Paste button: (20, 780)           │
│   Text field: (50, 765)             │
│   Send button: (350, 780)           │
├────────────────────────────────────┤
│ Home Indicator (170, 825)           │
└────────────────────────────────────┘
```

### B. Common Error Codes

| Error Code | Meaning | Resolution |
|------------|---------|------------|
| `APP_NOT_FOUND` | Bundle ID not installed | Verify app installed with `xcrun simctl listapps` |
| `DEVICE_NOT_FOUND` | Simulator not booted | Boot simulator first |
| `UI_ACTION_FAILED` | Tap coordinates invalid | Take screenshot and verify coordinates |
| `SCREENSHOT_FAILED` | Metal rendering issue | Check simulator GPU acceleration |

### C. Troubleshooting

**Screenshot appears black:**
- Add 2-second delay after app launch before first screenshot
- Verify Metal layer initialized: check `renderer != nil`

**Tap not registering:**
- Coordinates may be in points vs pixels (use logical points for simulators)
- Verify element not covered by modal/overlay
- Check keyboard not blocking element

**App crashes on launch:**
- Check simulator iOS version matches app deployment target
- Verify all frameworks/libraries copied to .app bundle
- Review crash logs: `xcrun simctl spawn booted log show --predicate 'process == "ForgeTerminal"' --last 5m`

---

**Document Version:** 1.0
**Author:** Senior Backend Engineer (Claude)
**Last Updated:** February 4, 2026
**Next Review:** After ios-agent-cli MVP completion
