# Discord Setup Guide - MTG Card Bot

Complete step-by-step guide to deploying your MTG Card Bot on Discord servers.

## Table of Contents

1. [Discord Developer Portal Setup](#1-discord-developer-portal-setup)
2. [Bot Configuration](#2-bot-configuration)
3. [Server Permissions Setup](#3-server-permissions-setup)
4. [Bot Deployment](#4-bot-deployment)
5. [Testing & Verification](#5-testing--verification)
6. [Troubleshooting](#6-troubleshooting)

---

## 1. Discord Developer Portal Setup

### Step 1.1: Create Discord Application

1. **Navigate to Discord Developer Portal**
   - Go to [https://discord.com/developers/applications](https://discord.com/developers/applications)
   - Log in with your Discord account

2. **Create New Application**
   - Click **"New Application"** button (top right)
   - Enter application name: `MTG Card Bot` (or your preferred name)
   - Click **"Create"**

3. **Configure Application Settings**
   - **Name**: `MTG Card Bot`
   - **Description**: `Magic: The Gathering card lookup bot with fuzzy search and high-quality images`
   - **Tags**: `Magic`, `Cards`, `MTG`, `Lookup`, `Games`
   - **App Icon**: Upload an MTG or bot-related image (optional)

### Step 1.2: Create Bot User

1. **Navigate to Bot Section**
   - In your application, click **"Bot"** in the left sidebar

2. **Create Bot**
   - Click **"Add Bot"** button
   - Confirm by clicking **"Yes, do it!"**

3. **Configure Bot Settings**

   **Basic Information:**
   - **Username**: `MTG-Card-Bot` (or your preferred name)
   - **Avatar**: Upload bot avatar image (optional)

   **Authorization Flow:**
   - ✅ **Public Bot**: `ON` (allows others to invite your bot)
   - ❌ **Require OAuth2 Code Grant**: `OFF` (not needed for this bot)

   **Privileged Gateway Intents:**
   - ❌ **Presence Intent**: `OFF` (not needed)
   - ❌ **Server Members Intent**: `OFF` (not needed)
   - ✅ **Message Content Intent**: `ON` (required to read message content)

4. **Copy Bot Token**
   - Click **"Reset Token"** to generate a new token
   - **⚠️ IMPORTANT**: Copy the token immediately and store it securely
   - **Never share this token publicly or commit it to version control**

---

## 2. Bot Configuration

### Step 2.1: Environment Setup

1. **Clone Repository**

   ```bash
   git clone https://github.com/dunamismax/go-discord-bots.git
   cd go-discord-bots
   ```

2. **Install Dependencies**

   ```bash
   go mod tidy
   go install github.com/magefile/mage@latest
   ```

3. **Create Environment File**

   ```bash
   cp .env.example .env
   ```

4. **Configure Environment Variables**

   Edit `.env` file with your settings:

   ```bash
   # Required - Your Discord bot token
   DISCORD_TOKEN=your_bot_token_here_from_step_1.2
   
   # Optional - Bot configuration
   COMMAND_PREFIX=!
   LOG_LEVEL=info
   BOT_NAME=mtg-card-bot
   ```

### Step 2.2: Development Setup

1. **Install Development Tools**

   ```bash
   mage setup
   ```

2. **Verify Configuration**

   ```bash
   mage status
   ```

   Expected output:

   ```
   Discord Bot Development Environment Status
   =========================================
   Go: go version go1.24+ darwin/arm64
   Environment: .env file found ✓
   Bots: 1 found (mtg-card-bot)
   Built binaries: None found
   ```

---

## 3. Server Permissions Setup

### Step 3.1: Generate Invite Link

1. **Navigate to OAuth2 Section**
   - In Discord Developer Portal, go to **"OAuth2"** → **"URL Generator"**

2. **Select Scopes**
   - ✅ **bot**: Required for bot functionality
   - ✅ **applications.commands**: Required for slash commands (future feature)

3. **Select Bot Permissions**

   **Essential Permissions:**
   - ✅ **Send Messages**: Post card information
   - ✅ **Embed Links**: Display rich card embeds
   - ✅ **Attach Files**: Send card images (if needed)
   - ✅ **Read Message History**: Process commands
   - ✅ **Use External Emojis**: Enhanced embed formatting (optional)

   **Recommended Additional Permissions:**
   - ✅ **Add Reactions**: React to messages (optional)
   - ✅ **Manage Messages**: Delete invalid commands (optional)

4. **Copy Generated URL**
   - Copy the generated URL at the bottom
   - Example: `https://discord.com/api/oauth2/authorize?client_id=YOUR_CLIENT_ID&permissions=412384301120&scope=bot%20applications.commands`

### Step 3.2: Add Bot to Server

1. **Visit Invite URL**
   - Paste the generated URL into your browser
   - Log in to Discord if prompted

2. **Select Server**
   - Choose the Discord server where you want to add the bot
   - You must have **"Manage Server"** permission

3. **Confirm Permissions**
   - Review the permissions list
   - Click **"Authorize"**
   - Complete any CAPTCHA if presented

4. **Verify Bot Addition**
   - Check your Discord server's member list
   - The bot should appear offline (until you start it)

---

## 4. Bot Deployment

### Step 4.1: Local Development

**Start Development Server:**

```bash
mage dev mtg-card-bot
```

Expected output:

```
Starting mtg-card-bot in development mode with auto-restart...
Press Ctrl+C to stop.
2024/08/08 12:00:00 Starting MTG Card Bot...
2024/08/08 12:00:00 Bot Name: mtg-card-bot
2024/08/08 12:00:00 Command Prefix: !
2024/08/08 12:00:00 Log Level: info
2024/08/08 12:00:00 Scryfall client initialized
2024/08/08 12:00:00 Bot is now running as MTG-Card-Bot
```

### Step 4.2: Production Deployment

**Option A: Direct Binary Deployment**

1. **Build Production Binary**

   ```bash
   mage build
   ```

2. **Deploy Binary**

   ```bash
   # Copy binary to production server
   scp bin/mtg-card-bot user@your-server:/opt/mtg-card-bot/
   
   # Run with environment variables
   DISCORD_TOKEN=your_token LOG_LEVEL=info ./mtg-card-bot
   ```

**Option B: Systemd Service (Ubuntu/Linux)**

1. **Create Service User**

   ```bash
   sudo useradd -r -s /bin/false mtgbot
   sudo mkdir -p /opt/mtg-card-bot
   sudo chown mtgbot:mtgbot /opt/mtg-card-bot
   ```

2. **Deploy Binary**

   ```bash
   sudo cp bin/mtg-card-bot /opt/mtg-card-bot/
   sudo chmod +x /opt/mtg-card-bot/mtg-card-bot
   sudo chown mtgbot:mtgbot /opt/mtg-card-bot/mtg-card-bot
   ```

3. **Create Systemd Service**

   ```bash
   sudo tee /etc/systemd/system/mtg-card-bot.service > /dev/null << EOF
   [Unit]
   Description=MTG Card Discord Bot
   After=network.target
   Wants=network.target
   
   [Service]
   Type=simple
   User=mtgbot
   Group=mtgbot
   WorkingDirectory=/opt/mtg-card-bot
   ExecStart=/opt/mtg-card-bot/mtg-card-bot
   Restart=always
   RestartSec=10
   
   # Environment variables
   Environment=DISCORD_TOKEN=your_bot_token_here
   Environment=LOG_LEVEL=info
   Environment=COMMAND_PREFIX=!
   Environment=BOT_NAME=mtg-card-bot
   
   # Security settings
   NoNewPrivileges=true
   PrivateTmp=true
   ProtectSystem=strict
   ProtectHome=true
   ReadWritePaths=/opt/mtg-card-bot
   
   [Install]
   WantedBy=multi-user.target
   EOF
   ```

4. **Start and Enable Service**

   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable mtg-card-bot
   sudo systemctl start mtg-card-bot
   ```

5. **Check Service Status**

   ```bash
   sudo systemctl status mtg-card-bot
   sudo journalctl -u mtg-card-bot -f
   ```

**Option C: Docker Deployment**

1. **Create Dockerfile** (if not exists)

   ```dockerfile
   FROM golang:1.24-alpine AS builder
   WORKDIR /app
   COPY . .
   RUN go mod download
   RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o mtg-card-bot ./bots/mtg-card-bot/main.go
   
   FROM alpine:latest
   RUN apk --no-cache add ca-certificates tzdata
   WORKDIR /root/
   COPY --from=builder /app/mtg-card-bot .
   
   ENV DISCORD_TOKEN=""
   ENV LOG_LEVEL=info
   ENV COMMAND_PREFIX=!
   ENV BOT_NAME=mtg-card-bot
   
   CMD ["./mtg-card-bot"]
   ```

2. **Build and Run**

   ```bash
   docker build -t mtg-card-bot .
   docker run -e DISCORD_TOKEN=your_token_here mtg-card-bot
   ```

---

## 5. Testing & Verification

### Step 5.1: Basic Functionality Tests

1. **Verify Bot is Online**
   - Check Discord server member list
   - Bot should show as "Online" with green status

2. **Test Card Lookup Commands**

   ```
   # Test exact card name
   !lightning bolt
   
   # Test fuzzy matching  
   !jac bele
   
   # Test random card
   !random
   
   # Test complex card name
   !the-one-ring
   ```

3. **Expected Response Format**
   - Rich embed with card image
   - Card name as title
   - Mana cost in description
   - Set name and rarity in footer
   - Clickable link to Scryfall page
   - Color-coded by rarity

### Step 5.2: Error Handling Tests

1. **Test Invalid Card Name**

   ```
   !nonexistentcard12345
   ```

   Expected: Error message with suggestions

2. **Test Empty Command**

   ```
   !
   ```

   Expected: No response (ignored)

3. **Test Rate Limiting**
   - Send multiple commands rapidly
   - Bot should handle gracefully without errors

### Step 5.3: Performance Verification

1. **Check Response Time**
   - Commands should respond within 1-3 seconds
   - Image loading may take slightly longer

2. **Monitor Resource Usage**

   ```bash
   # Check memory usage
   ps aux | grep mtg-card-bot
   
   # Check logs for errors
   sudo journalctl -u mtg-card-bot -n 50
   ```

---

## 6. Troubleshooting

### Common Issues

#### Bot Appears Offline

**Symptoms**: Bot shows as offline in Discord

**Solutions**:

1. **Check Token**: Verify `DISCORD_TOKEN` is correct in `.env`
2. **Check Intents**: Ensure "Message Content Intent" is enabled
3. **Check Logs**: Look for authentication errors

   ```bash
   mage dev mtg-card-bot
   # or
   sudo journalctl -u mtg-card-bot -f
   ```

#### Bot Doesn't Respond to Commands

**Symptoms**: Bot is online but ignores commands

**Possible Causes & Solutions**:

1. **Missing Message Content Intent**
   - Go to Discord Developer Portal → Bot → Privileged Gateway Intents
   - Enable "Message Content Intent"

2. **Wrong Command Prefix**
   - Check `COMMAND_PREFIX` in `.env` (default: `!`)
   - Test with correct prefix: `!lightning bolt`

3. **Insufficient Permissions**
   - Check bot has "Send Messages" and "Embed Links" permissions
   - Re-invite bot with correct permissions

4. **Bot Not in Channel**
   - Ensure bot has access to the channel you're testing in
   - Check channel permissions

#### API Errors

**Symptoms**: "Sorry, something went wrong" messages

**Solutions**:

1. **Rate Limiting**: Wait a moment between commands
2. **Network Issues**: Check internet connectivity
3. **Scryfall API Status**: Check [status.scryfall.com](https://status.scryfall.com)

#### Performance Issues

**Symptoms**: Slow responses or high memory usage

**Solutions**:

1. **Check System Resources**

   ```bash
   top -p $(pgrep mtg-card-bot)
   ```

2. **Restart Bot**

   ```bash
   sudo systemctl restart mtg-card-bot
   ```

3. **Check Logs for Memory Leaks**

   ```bash
   sudo journalctl -u mtg-card-bot | grep -i "memory\|leak\|panic"
   ```

### Log Analysis

**Viewing Logs**:

```bash
# Development mode
mage dev mtg-card-bot

# Production systemd
sudo journalctl -u mtg-card-bot -f

# Docker
docker logs -f container_name
```

**Important Log Messages**:

- ✅ `Bot is now running as MTG-Card-Bot` - Successful startup
- ✅ `Scryfall client initialized` - API client ready
- ❌ `invalid authentication token` - Wrong Discord token
- ❌ `missing access` - Insufficient permissions
- ⚠️ `rate limit` - Temporary API limitation

### Getting Help

1. **Check Repository Issues**: [GitHub Issues](https://github.com/dunamismax/go-discord-bots/issues)
2. **Discord Developer Documentation**: [Discord API Docs](https://discord.com/developers/docs)
3. **Scryfall API Status**: [status.scryfall.com](https://status.scryfall.com)

### Security Best Practices

1. **Token Security**
   - Never commit tokens to version control
   - Use environment variables or secure secrets management
   - Regenerate tokens if compromised

2. **Server Security**
   - Run bot with minimal privileges
   - Use systemd security features (shown in deployment section)
   - Regularly update dependencies

3. **Monitoring**
   - Set up log monitoring
   - Monitor API usage and rate limits
   - Set up alerts for bot downtime

---

## Quick Reference Commands

```bash
# Development
mage dev mtg-card-bot     # Start with auto-restart
mage status               # Check environment
mage build               # Build production binary

# Production Management
sudo systemctl status mtg-card-bot
sudo systemctl restart mtg-card-bot
sudo journalctl -u mtg-card-bot -f

# Testing Commands
!lightning bolt           # Test card lookup
!random                  # Test random card
!jac bele                # Test fuzzy search
```

---

**Success!** Your MTG Card Bot should now be running and responding to commands in your Discord server. Users can look up Magic: The Gathering cards using `!<card-name>` or get random cards with `!random`.
