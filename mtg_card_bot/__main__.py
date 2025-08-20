"""Main entry point for the MTG Card Bot."""

import asyncio
import signal
import sys
from pathlib import Path

from . import config, logging
from .bot import MTGCardBot


async def async_main() -> None:
    """Main function to run the MTG Card Bot."""
    # Load environment from .env file if it exists
    env_file = Path(".env")
    if env_file.exists():
        config.load_env_file(env_file)

    # Load configuration
    try:
        cfg = config.load_config()
    except Exception as e:
        print(f"Failed to load configuration: {e}")
        sys.exit(1)

    # Validate configuration
    try:
        cfg.validate_config()
    except Exception as e:
        print(f"Invalid configuration: {e}")
        sys.exit(1)

    # Initialize logging
    logging.initialize_logger(cfg.log_level, cfg.json_logging)
    logger = logging.with_component("main")

    logger.info("Starting MTG Card Bot", version="2.0.0")

    # Create and start the bot
    bot = MTGCardBot(cfg)

    # Set up signal handlers for graceful shutdown
    shutdown_event = asyncio.Event()

    def signal_handler(signum: int, frame) -> None:
        logger.info("Received shutdown signal", signal=signum)
        shutdown_event.set()

    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    try:
        # Start bot in background
        bot_task = asyncio.create_task(bot.start())

        # Wait for shutdown signal or bot failure
        done, pending = await asyncio.wait(
            [bot_task, asyncio.create_task(shutdown_event.wait())],
            return_when=asyncio.FIRST_COMPLETED,
        )

        # Cancel pending tasks
        for task in pending:
            task.cancel()
            try:
                await task
            except asyncio.CancelledError:
                pass

        # Clean shutdown
        logger.info("Shutting down MTG Card bot...")
        await bot.close()

        # Check if bot task failed
        if bot_task in done and not bot_task.cancelled():
            try:
                bot_task.result()
            except Exception as e:
                logger.error("MTG Card bot failed", error=str(e))
                sys.exit(1)

    except Exception as e:
        logger.error("Failed to start MTG Card bot", error=str(e))
        await bot.close()
        sys.exit(1)


def main() -> None:
    """Entry point for the MTG Card Bot console script."""
    asyncio.run(async_main())


if __name__ == "__main__":
    main()
