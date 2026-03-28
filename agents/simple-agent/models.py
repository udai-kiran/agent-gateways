from datetime import datetime

from pydantic import BaseModel, Field
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    model_config = {"env_file": ".env", "env_file_encoding": "utf-8"}

    openai_api_key: str
    openai_model_id: str = "gpt-4o"
    redis_url: str = "redis://localhost:6379"
    conversation_ttl: int = 3600


class IncomingMessage(BaseModel):
    platform: str = Field(pattern=r"^telegram$")
    sender_id: str = Field(min_length=1)
    conversation_id: str = Field(min_length=1)
    text: str = Field(min_length=1)
    timestamp: datetime


class AgentResponse(BaseModel):
    text: str
