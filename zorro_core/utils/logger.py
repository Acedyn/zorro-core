import logging
import os


class ColoredFormatter(logging.Formatter):
    """Formatter that add colors according to the log level"""

    blue = "\u001b[34m"
    green = "\u001b[32m"
    yellow = "\u001b[33m"
    red = "\u001b[31m"
    reset = "\u001b[37m"
    format_string = (
        "[%(asctime)s] | %(levelname)s: "
        + reset
        + "%(message)s (%(filename)s:%(lineno)d)"
    )

    FORMATS = {
        logging.DEBUG: blue + format_string,
        logging.INFO: green + format_string,
        logging.WARNING: yellow + format_string,
        logging.ERROR: red + format_string,
        logging.CRITICAL: red + format_string,
    }

    def format(self, record: logging.LogRecord) -> str:
        log_fmt = self.FORMATS.get(record.levelno)
        colored_formatter = logging.Formatter(log_fmt)
        return colored_formatter.format(record)


logger = logging.getLogger("zorro")
stream_handler = logging.StreamHandler()
logger.handlers = [stream_handler]
stream_handler.setFormatter(ColoredFormatter())
logger.setLevel(getattr(logging, os.getenv("ZORRO_LOG_LEVEL", "DEBUG").upper()))
