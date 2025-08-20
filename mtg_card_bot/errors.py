"""Simple error handling for MTG Card bot."""

from enum import Enum


class ErrorType(str, Enum):
    """Error type enumeration."""

    API = "api_error"
    NOT_FOUND = "not_found_error"
    RATE_LIMIT = "rate_limit_error"
    NETWORK = "network_error"
    VALIDATION = "validation_error"
    UNKNOWN = "unknown_error"
    DEPENDENCY = "dependency_error"
    PERMISSION = "permission_error"


class MTGError(Exception):
    """MTG Card bot specific error."""

    def __init__(self, error_type: ErrorType, message: str, cause: Exception = None):
        self.error_type = error_type
        self.message = message
        self.cause = cause
        super().__init__(message)


def create_error(
    error_type: ErrorType, message: str, cause: Exception = None
) -> MTGError:
    """Create an MTGError with the given type and message."""
    return MTGError(error_type, message, cause)
