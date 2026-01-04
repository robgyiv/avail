# avail

Generate up-to-date availability you can share — without booking links.

## What avail is
- A local-first availability generator
- Privacy-first by design
- Works in your terminal
- Produces text, links, and proposals

## What avail is not
- A scheduler
- A booking system
- A calendar writer
- A background sync service

## How it works
1. Read your calendar (locally where possible)
2. Derive free/busy availability
3. Generate human-friendly output
4. Optionally publish a live link

## Privacy model
- Event details never leave your machine
- Only derived availability is shared
- Read-only access only

## Paid features
- Shareable links
- Longer expiry
- Multiple active links

## Philosophy
Avail assists conversations — it does not automate them.

---

## Setup

### Prerequisites

- Go 1.21 or later

### Building from Source

```bash
# Clone the repository
git clone https://github.com/robgyiv/avail.git
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
- **Public Calendar URLs** (any service serving iCalendar format - privacy-first, read-only)
- **Local Calendar Files** (.ics files)

#### Google Calendar Setup (OAuth2)

Avail requires you to create your own OAuth application for privacy. We don't provide shared credentials to ensure your calendar data never passes through third-party servers when using this library.

1. **Create OAuth credentials:**
   - Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
   - Create a new project (or select existing)
   - Enable the Google Calendar API
   - Create OAuth 2.0 Client ID
     - Application type: **Desktop app**
     - Name: "Avail CLI" (or any name you prefer)
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
   
   This opens a browser for OAuth authentication. The token is stored securely in your system keyring.

**Privacy:** Your OAuth credentials are used only by your local CLI. Calendar data is processed locally and never sent to any third-party servers.

#### Public Calendar URL Setup

Avail can fetch events from any publicly accessible calendar URL that serves iCalendar (.ics) format.

**Supported sources:**
- Apple/iCloud public calendars
- Google Calendar public feeds
- CalDAV server public calendars
- Any HTTP/HTTPS endpoint serving `.ics` format

**Setup:**

1. **Get your public calendar URL:**
   - **Apple/iCloud**: Open Calendar app → Calendars → Info icon → Toggle "Public Calendar" → "Share Link"
   - **Google Calendar**: Calendar Settings → Integrate calendar → Public URL
   - **Other services**: Check your calendar provider's documentation for public feed URLs

2. **Authenticate:**
   ```bash
   avail auth --provider network --url "https://calendar.example.com/public.ics"
   ```
   
   Or configure in `config.toml`:
   ```toml
   calendar_provider = "network"
   calendar_url = "https://calendar.example.com/public.ics"
   ```

**Privacy Note:** Only works with calendars explicitly made public. Private calendars require OAuth authentication (see Google Calendar setup). Public calendar URLs are readable by anyone with the URL.

**URL formats:**
- `https://` URLs (recommended)
- `http://` URLs (if server doesn't support HTTPS)
- `webcal://` URLs (automatically converted to `https://`)

#### Switching Providers

To switch between providers, update your config file:

```toml
calendar_provider = "network"  # or "google" or "local"
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
