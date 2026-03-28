import json
from abc import ABC, abstractmethod

import redis.asyncio as redis


class ConversationHistory(ABC):
    """Interface for storing and retrieving conversation history."""

    @abstractmethod
    async def load(self, conversation_id: str) -> list[dict]: ...

    @abstractmethod
    async def save(self, conversation_id: str, messages: list[dict]) -> None: ...


class RedisConversationHistory(ConversationHistory):
    def __init__(self, r: redis.Redis, ttl: int = 3600) -> None:
        self._redis = r
        self._ttl = ttl

    def _key(self, conversation_id: str) -> str:
        return f"conversation:{conversation_id}"

    async def load(self, conversation_id: str) -> list[dict]:
        raw = await self._redis.get(self._key(conversation_id))
        if raw is None:
            return []
        return json.loads(raw)

    async def save(self, conversation_id: str, messages: list[dict]) -> None:
        await self._redis.set(
            self._key(conversation_id),
            json.dumps(messages, default=str),
            ex=self._ttl,
        )
