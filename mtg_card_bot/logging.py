"""Simple logging functionality for MTG Card bot."""

import logging
import sys
from datetime import UTC, datetime
from typing import Any


class Logger:
    """Simple logger for MTG Card bot."""

    def __init__(self, component: str = "mtg_card_bot"):
        self.component = component
        self._setup_logging()

    def _setup_logging(self):
        """Set up basic logging."""
        logging.basicConfig(
            level=logging.INFO,
            format="%(asctime)s [%(levelname)s] %(message)s",
            handlers=[logging.StreamHandler(sys.stdout)],
        )
        self.logger = logging.getLogger(self.component)

    def debug(self, message: str, **kwargs: Any) -> None:
        """Log a debug message."""
        self._log("DEBUG", message, **kwargs)

    def info(self, message: str, **kwargs: Any) -> None:
        """Log an info message."""
        self._log("INFO", message, **kwargs)

    def warning(self, message: str, **kwargs: Any) -> None:
        """Log a warning message."""
        self._log("WARNING", message, **kwargs)

    def error(self, message: str, **kwargs: Any) -> None:
        """Log an error message."""
        self._log("ERROR", message, **kwargs)

    def _log(self, level: str, message: str, **kwargs: Any) -> None:
        """Internal logging method."""
        timestamp = datetime.now(UTC).isoformat()

        # Build log message with context
        log_parts = ["[MTG]", f"{timestamp}", f"[{level}]", message]

        # Add context fields
        if kwargs:
            context_parts = []
            for key, value in kwargs.items():
                context_parts.append(f"{key}={value}")
            if context_parts:
                log_parts.append(" ".join(context_parts))

        full_message = " ".join(log_parts)

        # Use appropriate logging level
        if level == "DEBUG":
            self.logger.debug(full_message)
        elif level == "INFO":
            self.logger.info(full_message)
        elif level == "WARNING":
            self.logger.warning(full_message)
        elif level == "ERROR":
            self.logger.error(full_message)


def initialize_logger(level: str = "info", json_format: bool = False) -> None:
    """Initialize the global logger with the specified level and format."""
    level_map = {
        "debug": logging.DEBUG,
        "info": logging.INFO,
        "warn": logging.WARNING,
        "warning": logging.WARNING,
        "error": logging.ERROR,
    }
    log_level = level_map.get(level.lower(), logging.INFO)

    logging.basicConfig(
        level=log_level,
        format="%(asctime)s [%(levelname)s] %(message)s",
        handlers=[logging.StreamHandler(sys.stdout)],
    )


def with_component(component: str) -> Logger:
    """Return a logger with a component field."""
    return Logger(component)
