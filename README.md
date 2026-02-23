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
- Read-only access wherever possible

## Philosophy

Avail assists conversations — it does not automate them.

**NOTE**: This project is under active development and breaking changes are likely.

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
- `include_weekends` - Include Saturday/Sunday availability (default: false)
- `[[calendars]]` - Array of calendar configurations (supports multiple calendars)

Each calendar entry requires:
- `provider` - Calendar provider: `"google"`, `"network"`, or `"local"`
- `calendar_id` - (Google only, optional) Calendar ID: `"primary"` or email address (default: "primary")
- `url` - (Network only) Public calendar URL
- `path` - (Local only) Path to .ics file

Example config with multiple calendars (Google + Network + Local):

```toml
timezone = "America/New_York"
meeting_duration = "30m"
work_hours_start = "09:00"
work_hours_end = "17:00"
include_weekends = false

[[calendars]]
provider = "google"
calendar_id = "primary"

[[calendars]]
provider = "google"
calendar_id = "work@example.com"

[[calendars]]
provider = "network"
url = "https://calendar.example.com/public.ics"

[[calendars]]
provider = "local"
path = "~/Desktop/calendar.ics"
```

Example config for single Google Calendar:

```toml
timezone = "America/New_York"
meeting_duration = "30m"
work_hours_start = "09:00"
work_hours_end = "17:00"
include_weekends = false

[[calendars]]
provider = "google"
calendar_id = "primary"
```

Example config for single public calendar URL:

```toml
timezone = "America/New_York"

[[calendars]]
provider = "network"
url = "https://calendar.example.com/public.ics"
```

Example config for single local file:

```toml
[[calendars]]
provider = "local"
path = "~/Desktop/calendar.ics"
```

### Authentication

Before using `avail`, you need to configure and authenticate with your calendar providers. The tool supports:

- **Google Calendar** (OAuth2) - supports multiple calendars per account
- **Public Calendar URLs** (any service serving iCalendar format - privacy-first, read-only)
- **Local Calendar Files** (.ics files)

You can mix and match providers - configure multiple calendars of different types and avail will fetch from all of them.

**Configuration-first approach:** Add all your calendars to the config file first, then run `avail auth` to authenticate.

#### Google Calendar Setup (OAuth2)

Avail requires you to create your own OAuth application for privacy. We don't provide shared credentials to ensure your calendar data never passes through third-party servers when using this library.

1. **Configure calendars:**
   Edit `~/.config/avail/config.toml` and add one or more Google Calendar entries:

   ```toml
   [[calendars]]
   provider = "google"
   calendar_id = "primary"

   [[calendars]]
   provider = "google"
   calendar_id = "work@example.com"
   ```

   - `calendar_id = "primary"` - Your main calendar
   - `calendar_id = "email@gmail.com"` - Other calendars you own (using their email address)

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

   This opens a browser for OAuth authentication once (the token covers all your Google Calendars). The token is stored securely in your system keyring.

**Privacy:** Your OAuth credentials are used only by your local CLI. Calendar data is processed locally and never sent to any third-party servers.

#### Public Calendar URL Setup

Avail can fetch events from any publicly accessible calendar URL that serves iCalendar (.ics) format. You can add multiple public calendars.

**Supported sources:**

- Apple/iCloud public calendars
- Google Calendar public feeds
- CalDAV server public calendars
- Any HTTP/HTTPS endpoint serving `.ics` format

**Setup:**

1. **Configure calendars:**
   Edit `~/.config/avail/config.toml` and add one or more public calendar entries:

   ```toml
   [[calendars]]
   provider = "network"
   url = "https://calendar.example.com/public.ics"

   [[calendars]]
   provider = "network"
   url = "https://another-calendar.example.com/feed.ics"
   ```

   To get your public calendar URL:

   - **Apple/iCloud**: Open Calendar app → Calendars → Info icon → Toggle "Public Calendar" → "Share Link"
   - **Google Calendar**: Calendar Settings → Integrate calendar → Public URL
   - **Other services**: Check your calendar provider's documentation for public feed URLs

2. **Use directly:**

   ```bash
   avail show
   ```

   No authentication needed - public calendars are read directly from the URLs.

**Privacy Note:** Only works with calendars explicitly made public. Private calendars require OAuth authentication (see Google Calendar setup). Public calendar URLs are readable by anyone with the URL.

**URL formats:**

- `https://` URLs (recommended)
- `http://` URLs (if server doesn't support HTTPS)
- `webcal://` URLs (automatically converted to `https://`)

#### Local Calendar File Setup

Use one or more local `.ics` files on your filesystem. You can mix different exported calendars together.

To export your calendar:

1. Open your calendar application:

   - macOS: Calendar.app (File > Export > Export...)
   - Google Calendar: Settings > Export calendar
   - Other: Check your calendar app's export options

2. Export as .ics format and note the filepath(s).

3. **Configure calendars:**
   Edit `~/.config/avail/config.toml` and add one or more local calendar entries:

   ```toml
   [[calendars]]
   provider = "local"
   path = "~/Desktop/personal.ics"

   [[calendars]]
   provider = "local"
   path = "~/Desktop/work.ics"
   ```

   You can use `~` to refer to your home directory.

4. **Use directly:**

   ```bash
   avail show
   ```

   No authentication needed - files are read directly from the filesystem.

### Show Availability

Display your availability for the next 5 days:

```bash
avail show
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
avail copy
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
avail push
```

By default, this pushes availability for the next 5 days. You can customize the number of days:

```bash
avail push --days 7
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
avail --help
avail show --help
avail copy --help
avail push --help
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

avail consists of an open-source client and availability computation engine, alongside a hosted service that provides persistence, sharing, rate limiting, and integrations.

The boundary between open and hosted components may evolve over time.

---
