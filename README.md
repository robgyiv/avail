## Calendar availability tool

A reliable way to generate human-friendly availability text from a real calendar, without forcing the recipient into a tool.

Key properties:
- Host-centric
- Read-only calendar integration
- Zero commitment required from the guest
- Works in email, Slack, WhatsApp, etc.

Implemented first as a TUI to validate functionality and core library, then extended into a SaaS web application.

---

## Setup

### Prerequisites

- Go 1.21 or later

### Building from Source

```bash
# Clone the repository
git clone https://github.com/robgyiv/availability.git
cd availability

# Build the binary
go build -o bin/avail ./cmd/avail

# Or install directly
go install ./cmd/avail
```

### Development

```bash
# Run tests
go test ./...

# Build for multiple platforms
GOOS=darwin GOARCH=arm64 go build -o bin/avail-darwin-arm64 ./cmd/avail
GOOS=linux GOARCH=amd64 go build -o bin/avail-linux-amd64 ./cmd/avail
GOOS=windows GOARCH=amd64 go build -o bin/avail-windows-amd64.exe ./cmd/avail
```

---

## Usage

Currently, the following commands are implemented:

### Configuration

On first run, `avail` will create a default configuration file at `~/.config/avail/config.toml`. You can customize:

- `timezone` - Your timezone (default: "UTC")
- `meeting_duration` - Default meeting duration (default: 30 minutes)
- `work_hours_start` - Start of work day (default: "09:00")
- `work_hours_end` - End of work day (default: "17:00")
- `calendar_provider` - Calendar provider (default: "google")

Example config:

```toml
timezone = "America/New_York"
meeting_duration = "30m"
work_hours_start = "09:00"
work_hours_end = "17:00"
calendar_provider = "google"
```

### Authentication

Before using `avail`, you need to authenticate with your calendar provider. The tool supports:

- **Google Calendar** (OAuth2)
- **Apple/iCloud Calendar** (public calendar URL - privacy-first, read-only)

#### Google Calendar Setup

1. **Create OAuth credentials:**
   - Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
   - Create an OAuth 2.0 Client ID (Application type: Desktop app)
   - Note your Client ID and Client Secret

2. **Set environment variables:**
   ```bash
   export GOOGLE_CLIENT_ID="your-client-id"
   export GOOGLE_CLIENT_SECRET="your-client-secret"
   ```

3. **Authenticate:**
   ```bash
   avail auth --provider google
   ```

   This will open a browser window for OAuth authentication. The token will be stored securely in your system keyring.

#### Apple/iCloud Calendar Setup (Public Calendar - Privacy-First)

Apple/iCloud calendars use public calendar URLs for privacy-first, read-only access. No app-specific passwords needed!

1. **Get your public calendar URL:**
   - Open the Calendar app on your iPhone or Mac
   - Tap/click the "Calendars" button at the bottom
   - Tap/click the info icon (ℹ️) next to the calendar you want to use
   - Toggle on "Public Calendar"
   - Tap/click "Share Link" to copy the public calendar URL
   - The URL will look like: `webcal://p[numbers]-calendars.icloud.com/published/2/[long-string]`

2. **Authenticate:**
   ```bash
   avail auth --provider apple
   ```
   
   When prompted, paste your public calendar URL. Or use:
   ```bash
   avail auth --provider apple --url "your-public-calendar-url"
   ```

   The URL will be stored securely in your system keyring.

**Privacy Note:** Public calendars are read-only and don't require any credentials. Anyone with the URL can view your calendar, so only share it with people you trust. You can revoke access anytime by turning off "Public Calendar" in the Calendar app.

#### Switching Providers

To switch between providers, update your config file:

```toml
calendar_provider = "apple"  # or "google"
```

Then authenticate with the new provider using `avail auth --provider <provider>`.

### Show Availability

Display your availability for the next 5 days:

```bash
$ avail show
```

Output example:

```
Your availability (next 5 days):

Tue 12 Mar
  • 14:00–16:00

Wed 13 Mar
  • 10:00–11:30

Fri 15 Mar
  • after 13:00

Time zone: UTC
```

**Note:** Requires calendar authentication. See the [Authentication](#authentication) section below.

### Copy to Clipboard

Copy formatted availability text to your clipboard:

```bash
$ avail copy
```

This copies text in the following format:

```
I'm free:
• Tue 12 Mar 14:00–16:00
• Wed 13 Mar 10:00–11:30
• Fri 15 Mar after 13:00
```

Perfect for pasting into emails, Slack, or other messaging apps.

### Help

View available commands and options:

```bash
$ avail --help
$ avail show --help
$ avail copy --help
```

---

## Current Status

**Implemented:**
- ✅ Core availability engine
- ✅ Configuration system
- ✅ `avail show` command
- ✅ `avail copy` command
- ✅ `avail auth` command (OAuth flow for Google, public URL for Apple)
- ✅ Google Calendar API integration
- ✅ Apple/iCloud Calendar integration (public calendar URLs - privacy-first)

**In Progress / Planned:**
- ⏳ `avail link` command (shareable links)
- ⏳ `avail propose` command (interactive TUI for time selection)

---

## Reframed Capability

From the guest’s perspective:

> “I pick a time I like, then I get something useful I can send or add — nothing happens automatically.”

This reduces friction **without**:

* reserving the slot
* writing to the host’s calendar
* requiring accounts or permissions

---

## Updated Guest Flow (Availability Link)

1. Guest opens availability link
2. Sees days → time blocks
3. Clicks a block (e.g. Tue 14:00–16:00)
4. Chooses a specific start time (e.g. 14:30)
5. App generates:

   * a **proposed meeting payload**

Guest then chooses:

* 📅 “Add to my calendar”
* ✉️ “Email this time to host”
* 📋 “Copy details”

Nothing is confirmed yet.

---

## What the App Actually Generates

### A. Generic calendar event (ICS)

The event is **tentative**, not authoritative.

Fields:

* Title: `Proposed meeting with Alex`
* Start / end time
* Time zone
* Description:

  ```
  Proposed via Alex’s availability link.
  Please confirm before considering this final.
  ```

This works across:

* Google Calendar
* Apple Calendar
* Outlook

No integrations required.

---

### B. Structured email content

Auto-generated, human-readable:

```
Hi Alex,

I’m free at:
Tuesday 12 March, 14:30–15:00 (GMT)

Let me know if that works for you.

Best,
Sam
```

You’re not sending the email — just generating it.

---

### C. Copyable payload

For Slack / WhatsApp / Teams:

```
How about Tue 12 Mar at 14:30–15:00 (GMT)?
```

---

## Critical Constraint: No Slot Locking

This must be explicit in the UX.

* Selecting a time does **not** reserve it
* Multiple guests can propose the same time
* Host confirmation remains the source of truth

You should surface this with:

* copy (“This doesn’t book the time yet”)
* subtle UI affordances (e.g. dotted outlines, “proposed” labels)

This avoids legal and emotional liability.

---

## Updated Feature List (Tight)

### Feature 1: Read-only calendar integration

(unchanged)

### Feature 2: Availability engine

(unchanged, core)

### Feature 3: Shareable availability link

* Live view
* Mobile-friendly
* No auth

### Feature 4: Time selection → proposal generator

* Click block → pick start time
* Generate:

  * ICS
  * email text
  * copyable message

Still:

* ❌ no booking
* ❌ no confirmation
* ❌ no calendar writes

---

## Edge Cases to Handle (Now Explicit)

You’ve introduced a few — manageable, but must be defined:

* **Time disappears between view and click**

  * Show warning: “This time may no longer be available”
* **Guest calendar conflicts**

  * That’s their responsibility; you’re generating, not validating
* **Host changes calendar after proposal**

  * Host simply declines — same as email today
* **Time zones**

  * Always show:

    * guest-local time
    * host time zone in parentheses

---

## 1. Why a TUI / CLI MVP is a good idea

### This works because:

* Developers already live in terminals
* Copy-paste is the primary interaction
* Text output *is the product*
* You avoid premature UI bikeshedding
* You validate the **availability engine**, not your CSS

You’re effectively building a **calendar query tool** first.

This is closer to:

* `gh`
* `kubectl`
* `pass`
  than to Calendly.

That’s a *good* thing.

---

## 2. What the TUI *is* and *is not*

### The TUI *is*:

* A personal availability generator
* A formatter
* A link creator
* A proposal generator

### The TUI is *not*:

* A scheduler
* A daemon
* A calendar writer
* A background sync service

No long-running processes. No event listeners.

---

## 3. MVP CLI / TUI Flow (Planned)

> **Note:** This section describes the planned full MVP feature set. See the [Usage](#usage) section above for currently implemented commands.

### Installation

```
$ brew install avail
# or
$ npm install -g avail
```

---

### Auth (one-time) - *Planned*

```
$ avail auth
✔ Opening browser to connect Google Calendar…
✔ Calendar connected (read-only)
```

Tokens stored securely (keychain if possible).

---

### Basic availability - *Implemented*

```
$ avail show
```

Output:

```
Your availability (next 5 days):

Tue 12 Mar
  • 14:00–16:00

Wed 13 Mar
  • 10:00–11:30

Fri 15 Mar
  • after 13:00

Time zone: GMT
```

This alone is a useful tool.

---

### Copyable output - *Implemented*

```
$ avail copy
```

Copies to clipboard:

```
I'm free:
• Tue 12 Mar 14:00–16:00
• Wed 13 Mar 10:00–11:30
• Fri 15 Mar after 13:00
```

---

### Generate shareable link - *Planned*

```
$ avail link
```

Output:

```
Live availability:
https://avail.app/alex/abc123
(valid for 7 days)
```

---

### Propose a time (structured intent) - *Planned*

```
$ avail propose
```

Interactive TUI:

```
Select a day:
> Tue 12 Mar
  Wed 13 Mar
  Fri 15 Mar

Select a time:
> 14:30–15:00
  15:00–15:30
```

Result:

```
✔ Proposal created

Options:
[1] Copy message
[2] Generate .ics
[3] Email text
```

No network dependency beyond the proposal itself.

---

## 4. Minimal Web UI (Only Where Needed)

The web UI exists for **two reasons only**:

1. OAuth callback
2. Guest availability viewing

That’s it.

No dashboards.
No settings pages initially.

Host configuration can live in:

```
~/.config/avail/config.toml
```

Example:

```toml
timezone = "Europe/London"
meeting_duration = 30
work_hours = "09:00-17:00"
```

This is extremely dev-friendly.

---

## 5. Architecture Benefits (This Is the Hidden Win)

By going CLI-first:

* Your **availability engine** becomes a pure function
* Your web UI becomes a thin renderer
* Your API is naturally composable
* You avoid UI-driven design mistakes

Core shape:

```
calendar → availability engine → text / link / proposal
```

Same engine powers:

* CLI
* TUI
* web guest page
* future API integrations

That’s very clean.

---

## 6. MVP Scope (Now Very Clear)

### Phase 0 (Private alpha)

* CLI only
* Google Calendar
* Copyable availability
* Shareable link
* Proposal generator

### Phase 1

* Guest web view
* Time selection UI
* ICS download

### Phase 2 (if demanded)

* Lightweight web host UI
* Non-developer onboarding

---

## 7. Positioning (This Will Attract the Right Users)

You’re no longer competing with Calendly.

You’re competing with:

* writing emails manually
* thinking too hard about time zones
* calendar anxiety

Possible tagline:

> “Generate availability from your terminal.”

Or:

> “Scheduling, without scheduling.”

---

## 8. Where does config live long-term?

* Local-first (CLI owns config)
* Then server-stored and synced(CLI is a client)

---