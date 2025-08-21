# MTG Card Bot

<p align="center">
  <img src="https://github.com/dunamismax/images/blob/main/golang/discord-bots/mtg.png" alt="MTG Card Bot" width="400" />
</p>

<p align="center">
  <img src="https://readme-typing-svg.demolab.com/?font=Fira+Code&size=22&pause=1000&color=5865F2&center=true&vCenter=true&width=900&lines=Advanced+Magic+Card+Lookup+with+Live+Pricing;Smart+Fuzzy+Search+and+Advanced+Filtering;Real-Time+Market+Data+and+Format+Legality;Multi-Card+Grid+Display+with+Rich+Embeds;Scryfall+API+Integration+with+Rate+Limiting;Official+Rulings+and+Card+Image+Display;No+Caching+-+Always+Fresh+Card+Data;Rate+Limited+Anti-Spam+Protection;Modern+Python+3.13+Architecture;Built+with+discord.py+and+uv+Package+Manager" alt="Typing SVG" />
</p>

<p align="center">
  <a href="https://python.org/"><img src="https://img.shields.io/badge/Python-3.13+-3776AB.svg?logo=python&logoColor=white" alt="Python Version"></a>
  <a href="https://github.com/Rapptz/discord.py"><img src="https://img.shields.io/badge/Discord-discord.py-5865F2.svg?logo=discord&logoColor=white" alt="discord.py"></a>
  <a href="https://scryfall.com/docs/api"><img src="https://img.shields.io/badge/API-Scryfall-FF6B35.svg" alt="Scryfall API"></a>
  <a href="https://docs.astral.sh/uv/"><img src="https://img.shields.io/badge/Package%20Manager-uv-DE5FE9.svg" alt="uv"></a>
  <a href="https://github.com/structlog/structlog"><img src="https://img.shields.io/badge/Logging-structlog-blue.svg" alt="structlog"></a>
  <a href="https://opensource.org/licenses/Apache-2.0"><img src="https://img.shields.io/badge/License-Apache%202.0-green.svg" alt="Apache 2.0 License"></a>
</p>

A dedicated Magic: The Gathering card lookup Discord bot built in modern Python. Features fuzzy search, advanced filtering, random card discovery, and rich embeds powered by the Scryfall API with real-time pricing and format legality information.

## Quick Start

### Prerequisites

- **[uv](https://docs.astral.sh/uv/)** - Fast Python package manager
- **Discord Bot Token** - From [Discord Developer Portal](https://discord.com/developers/applications)

### Installation

```bash
# 1. Install uv (Python package manager)
curl -LsSf https://astral.sh/uv/install.sh | sh
source ~/.bashrc  # or restart your terminal

# 2. Install Python 3.13 and set as global
uv python install 3.13
uv python pin 3.13

# 3. Clone and setup project
git clone https://github.com/dunamismax/mtg-card-bot.git
cd mtg-card-bot

# 4. Configure environment
cp .env.example .env
# Edit .env with your Discord bot token

# 5. Install dependencies
uv sync

# 6. Run the bot (choose one method)
uv run python manage_bot.py start    # Using bot manager (recommended)
# OR
uv run python manage_bot.py          # Interactive management mode
```

### Environment Configuration

```bash
# Required Discord token
MTG_DISCORD_TOKEN=your_discord_bot_token_here

# Optional settings
MTG_COMMAND_PREFIX=!               # Command prefix (default: !)
MTG_LOG_LEVEL=INFO                 # Log level: DEBUG, INFO, WARNING, ERROR
MTG_JSON_LOGGING=false             # Use JSON structured logging
```

### Bot Management

The `manage_bot.py` script provides comprehensive bot management with both interactive and command-line modes:

```bash
# Interactive management mode (recommended for beginners)
uv run python manage_bot.py

# Direct commands
uv run python manage_bot.py start     # Start the bot with live logs
uv run python manage_bot.py stop      # Stop the bot gracefully
uv run python manage_bot.py restart   # Restart the bot
uv run python manage_bot.py status    # Check bot status
uv run python manage_bot.py kill      # Force kill all bot processes
uv run python manage_bot.py logs      # Monitor running bot logs
```

**Interactive Mode Features:**

- Menu-driven interface with numbered options
- Real-time process monitoring and status checking
- Graceful shutdown with fallback to force termination
- Environment variable validation
- Live log streaming during bot operation

## Bot Features

### Advanced Card Lookup

Comprehensive Magic: The Gathering card search with **live pricing**, format legality, and official rulings. **No caching** - always fresh data from Scryfall.

```bash
# Basic card lookup with pricing and legality
!lightning bolt
!the one ring
!jac bele                         # Fuzzy search: finds "Jace Beleren"
[[Lightning Bolt]]                # Alternative bracket syntax

# Official card rulings
!rules counterspell               # Get official rulings and errata
!rules lightning bolt

# Random card discovery
!random                           # Get any random Magic card
!random rarity:mythic             # Random mythic rare card
!random e:mh3                     # Random card from Modern Horizons 3
!random rarity:rare e:who         # Random rare from Doctor Who set

# Multi-card grid display
!black lotus; sol ring; time walk # Multiple cards in one command
!sol ring e:lea; mox ruby e:lea   # Filtered multi-card lookup

# Command aliases
!r, !rand, !h, !help, !?          # Shorter command variants
```

### Advanced Filtering

Support for all Scryfall filter syntax with smart fallback when filtered searches fail:

```bash
# Set filtering
!lightning bolt e:mh3             # From Modern Horizons 3
!sol ring e:lea                   # From Limited Edition Alpha
!brainstorm e:ice                 # From Ice Age

# Rarity and treatment filtering
!brainstorm is:foil               # Foil version only
!sol ring is:showcase             # Showcase treatment
!lightning bolt frame:borderless   # Borderless frame style
!force of will rarity:mythic      # Mythic rare versions only

# Advanced combinations
!force of will e:all is:foil frame:1997    # Specific set, foil, old frame
!lightning bolt is:fullart e:sta           # Full-art from Strixhaven Archives
!the one ring border:borderless e:ltr      # Borderless from Lord of the Rings
```

## Bot Commands Reference

### Basic Commands

- `!<card name>` - Look up any Magic card by name
- `!rules <card name>` - Get official rulings for a card
- `!random` - Get a random Magic card
- `!random <filters>` - Get a filtered random card
- `!help` - Show command help and examples

### Multi-Card Lookup

- `!card1; card2; card3` - Look up multiple cards (semicolon-separated)
- `!card1 filter; card2 filter` - Multi-card lookup with individual filters

### Filter Examples

- **Sets**: `e:mh3`, `e:ltr`, `e:who`, `e:lea`
- **Rarity**: `rarity:mythic`, `rarity:rare`, `rarity:uncommon`
- **Treatments**: `is:foil`, `is:showcase`, `frame:borderless`, `is:fullart`
- **Colors**: `c:red`, `c:blue`, `c:wubrg`

## Architecture

```sh
mtg-card-bot/
├── mtg_card_bot/           # Main bot package
│   ├── __init__.py
│   ├── __main__.py         # Entry point with main() function
│   ├── bot.py              # Core Discord bot logic
│   ├── config.py           # Configuration management
│   ├── errors.py           # Custom error types
│   ├── logging.py          # Structured logging setup
│   └── scryfall.py         # Scryfall API client
├── manage_bot.py           # Unified bot management script
├── pyproject.toml          # Project configuration & dependencies
├── .env.example            # Environment template
└── README.md               # Documentation
```

## Development

```bash
# Install dependencies
uv sync

# Run the bot (recommended methods)
uv run python manage_bot.py start       # Using management script
uv run python manage_bot.py             # Interactive mode

# Alternative direct execution
uv run python -m mtg_card_bot           # Direct module execution

# Development tools
uv run ruff format .                     # Code formatting
uv run ruff check .                      # Linting
uv run mypy mtg_card_bot/                # Type checking
uv run pytest                            # Run tests (when available)

# Bot management during development
uv run python manage_bot.py status      # Check bot status
uv run python manage_bot.py restart     # Quick restart during development
uv run python manage_bot.py logs        # Monitor logs
```

## Key Features

- **Smart Fuzzy Search** - Find cards with partial names and typos
- **Advanced Filtering** - Full Scryfall filter syntax support with fallback
- **Live Data Integration** - Real-time pricing from TCGPlayer and market data
- **Rich Discord Embeds** - High-quality card images with rarity-based colors
- **Multi-Card Display** - Grid layout for multiple card lookups
- **Official Rulings** - Access to comprehensive card rulings and errata
- **Rate Limiting** - Built-in API respect and anti-spam protection
- **Modern Python** - Type hints, async/await, structured logging
- **Zero Caching** - Always fresh card data and pricing information
- **Bracket Syntax** - Support for `[[card name]]` Magic community standard

## Bot in Action

The bot provides rich Discord embeds featuring:

- High-quality card images from Scryfall
- Rarity-based color coding for visual distinction
- Live pricing information from multiple markets
- Format legality across all major Magic formats
- Set information with artist credits
- Clickable Scryfall links for additional details
- Mana cost display with converted mana cost
- Multi-card grid layouts for batch lookups

## Deployment

### Local Deployment

```bash
# Install and run
uv sync
uv run python manage_bot.py start
```

### Systemd Service

Create `/etc/systemd/system/mtg-card-bot.service`:

```ini
[Unit]
Description=MTG Card Bot
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/path/to/mtg-card-bot
Environment=MTG_DISCORD_TOKEN=your_token_here
ExecStart=/home/your-user/.local/bin/uv run python manage_bot.py start
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

**Systemd Management:**

```bash
# Enable and start service
sudo systemctl enable mtg-card-bot
sudo systemctl start mtg-card-bot

# Check status and logs
sudo systemctl status mtg-card-bot
sudo journalctl -u mtg-card-bot -f
```

### Docker Deployment

```dockerfile
FROM python:3.13-slim
WORKDIR /app
COPY . .
RUN pip install uv && uv sync --frozen
CMD ["uv", "run", "python", "manage_bot.py", "start"]
```

**Docker Commands:**

```bash
# Build and run
docker build -t mtg-card-bot .
docker run -e MTG_DISCORD_TOKEN=your_token_here mtg-card-bot

# Docker Compose (create docker-compose.yml)
services:
  mtg-card-bot:
    build: .
    environment:
      - MTG_DISCORD_TOKEN=your_token_here
    restart: unless-stopped
```

## API Usage

The bot respects Scryfall's API guidelines with:

- Built-in rate limiting (max 10 requests/second)
- Proper error handling and retries
- User-agent identification for API tracking
- Duplicate request suppression
- Graceful fallback for failed filtered searches

## License

Apache 2.0 - see [LICENSE](LICENSE) for details.

---

<p align="center">
  <strong>Magic: The Gathering Discord Bot</strong><br>
  <sub>Built with Python 3.13+ • discord.py • Scryfall API • uv • Modern Architecture</sub>
</p>
