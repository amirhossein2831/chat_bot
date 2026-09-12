from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_prefix="OCR_", env_file=".env", extra="ignore")

    host: str = "0.0.0.0"
    port: int = 8100

    # "surya" needs torch and baked-in weights. "echo" exists so the rest of
    # the system can be exercised on a machine that has neither.
    recognizer: str = "surya"

    # The caller renders at 250 DPI, so a page is a megabyte or two of PNG.
    max_pages: int = 200
    max_page_bytes: int = 8 * 1024 * 1024

    languages: list[str] = ["fa", "en"]
