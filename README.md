# avail

Remove manual checking and generate up-to-date availability you can share.

## What avail is
- A local-first availability generator
- Privacy-first by design
- Works in your terminal
- Produces text you can share with guests easily

## What avail is not
- A scheduler
- A booking system
- A calendar writer

## How it works
1. Read your calendar (locally where possible)
2. Derive free/busy availability
3. Generate human-friendly output
4. Optionally publish a live link

## Privacy model
- Event details never leave your machine
- Only derived availability is shared
- Read-only access only

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
- `calendar_provider` - Calendar provider: `"google"`, `"network"`, or `"local"` (default: "google")
- `calendar_url` - Public calendar URL (required when `calendar_provider = "network"`)
- `local_calendar_path` - Path to local .ics file (required when `calendar_provider = "local"`)

Example config for Google Calendar:

```toml
timezone = "America/New_York"
meeting_duration = "30m"
work_hours_start = "09:00"
work_hours_end = "17:00"
calendar_provider = "google"
```

Example config for public calendar URL:

```toml
calendar_provider = "network"
calendar_url = "https://calendar.example.com/public.ics"
```

Example config for local file:

```toml
calendar_provider = "local"
local_calendar_path = "~/Desktop/calendar.ics"
```

### Authentication

Before using `avail`, you need to configure and authenticate with your calendar provider. The tool supports:

- **Google Calendar** (OAuth2)
- **Public Calendar URLs** (any service serving iCalendar format - privacy-first, read-only)
- **Local Calendar Files** (.ics files)

**Configuration-first approach:** Set your provider in the config file first, then run `avail auth` to authenticate.

#### Google Calendar Setup (OAuth2)

Avail requires you to create your own OAuth application for privacy. We don't provide shared credentials to ensure your calendar data never passes through third-party servers when using this library.

1. **Configure provider:**
   Edit `~/.config/avail/config.toml`:
   ```toml
   calendar_provider = "google"
   ```

2. **Create OAuth credentials:**
   - Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
   - Create a new project (or select existing)
   - Enable the Google Calendar API
   - Create OAuth 2.0 Client ID
     - Application type: **Desktop app**
     - Name: "Avail CLI" (or any name you prefer)
   - Note your Client ID and Client Secret

3. **Set environment variables:**
   ```bash
   export GOOGLE_CLIENT_ID="your-client-id"
   export GOOGLE_CLIENT_SECRET="your-client-secret"
   ```

4. **Authenticate:**
   ```bash
   avail auth
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

1. **Configure provider:**
   Edit `~/.config/avail/config.toml`:
   ```toml
   calendar_provider = "network"
   calendar_url = "https://calendar.example.com/public.ics"
   ```
   
   To get your public calendar URL:
   - **Apple/iCloud**: Open Calendar app → Calendars → Info icon → Toggle "Public Calendar" → "Share Link"
   - **Google Calendar**: Calendar Settings → Integrate calendar → Public URL
   - **Other services**: Check your calendar provider's documentation for public feed URLs

2. **Authenticate:**
   ```bash
   avail auth
   ```
   
   The URL is validated and stored securely in your system keyring.

**Privacy Note:** Only works with calendars explicitly made public. Private calendars require OAuth authentication (see Google Calendar setup). Public calendar URLs are readable by anyone with the URL.

**URL formats:**
- `https://` URLs (recommended)
- `http://` URLs (if server doesn't support HTTPS)
- `webcal://` URLs (automatically converted to `https://`)

#### Local Calendar File Setup

Use a local `.ics` file on your filesystem.

To export your calendar:

1. Open your calendar application:
   - macOS: Calendar.app (File > Export > Export...)
   - Google Calendar: Settings > Export calendar
   - Other: Check your calendar app's export options

2. Export as .ics format and save to: /Users/robbie/.config/avail/calendar.ics

3. **Configure provider:**
   Edit `~/.config/avail/config.toml`:
   ```toml
   calendar_provider = "local"
   local_calendar_path = "/path/to/calendar.ics"
   ```
   
   You can use `~` to refer to your home directory:
   ```toml
   local_calendar_path = "~/Desktop/calendar.ics"
   ```

4. **Use directly:**
   ```bash
   avail show
   ```
   
   No authentication needed - the file is read directly from the filesystem.

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

### Push to API

Push your availability to avail.website to create a shareable link:

```bash
$ avail push
```

By default, this pushes availability for the next 5 days. You can customize the number of days:

```bash
$ avail push --days 7
```

**Setup:**

1. **Sign up and generate a token:**
   - Visit [https://avail.website/](https://avail.website/)
   - Sign up for an account
   - Generate an API token

2. **Store your API token:**
   Create a file at `~/.config/avail/credentials` with your token:
   ```bash
   echo "avail_your_token_here" > ~/.config/avail/credentials
   chmod 600 ~/.config/avail/credentials
   ```

   The token should start with `avail_`. The credentials file is stored with restricted permissions (read/write for owner only).

3. **Push your availability:**
   ```bash
   avail push
   ```

The command will:
- Calculate your availability using the same logic as `avail show`
- Transform it to the API format
- POST it to `https://api.avail.website/v1/availability`
- Display a success message

**Note:** Your calendar event details never leave your machine. Only the derived availability time slots are sent to the API.

### Help

View available commands and options:

```bash
$ avail --help
$ avail show --help
$ avail copy --help
$ avail push --help
```

---

## Current Status

**Implemented:**
- ✅ Core availability engine
- ✅ Configuration system
- ✅ `avail show` command
- ✅ `avail copy` command
- ✅ `avail push` command (sync availability to avail.website API)
- ✅ `avail auth` command (OAuth flow for Google, public URL for Apple)
- ✅ Google Calendar API integration
- ✅ Apple/iCloud Calendar integration (public calendar URLs - privacy-first)

**In Progress / Planned:**
- ⏳ `avail link` command (shareable links)
- ⏳ `avail propose` command (interactive TUI for time selection)

---
