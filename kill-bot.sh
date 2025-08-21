#!/bin/bash

# Find and kill MTG card bot processes
pkill -f "mtg.*bot\|bot.*mtg\|mtg-card-bot" 2>/dev/null

# Also check for common Python processes that might be the bot
pkill -f "python.*bot\|bot.*python" 2>/dev/null

echo "MTG card bot processes killed"