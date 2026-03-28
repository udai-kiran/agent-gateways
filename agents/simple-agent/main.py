from contextlib import asynccontextmanager

import redis.asyncio as redis
import structlog
from fastapi import FastAPI

import agent as agent_module
from history import RedisConversationHistory
from logging_config import setup_logging
from models import AgentResponse, IncomingMessage, Settings

setup_logging()
logger = structlog.get_logger()


settings: Settings
history: RedisConversationHistory
redis_pool: redis.Redis


@asynccontextmanager
async def lifespan(app: FastAPI):
    global settings, history, redis_pool
    settings = Settings()
    redis_pool = redis.from_url(settings.redis_url, decode_responses=True)
    history = RedisConversationHistory(redis_pool, ttl=settings.conversation_ttl)
    yield
    await redis_pool.aclose()


app = FastAPI(lifespan=lifespan)


@app.post("/process")
async def process(message: IncomingMessage) -> AgentResponse:
    try:
        messages = await history.load(message.conversation_id)
        a = agent_module.create_agent(
            api_key=settings.openai_api_key,
            model_id=settings.openai_model_id,
            history=messages,
        )
        agent_response = await agent_module.run(a, message.text)
        await history.save(message.conversation_id, a.messages)
    except Exception:
        logger.exception(
            "agent_error",
            conversation_id=message.conversation_id,
            sender_id=message.sender_id,
            platform=message.platform,
        )
        return AgentResponse(
            text="Sorry, I'm having trouble right now. Please try again."
        )

    logger.info(
        "message_processed",
        conversation_id=message.conversation_id,
        sender_id=message.sender_id,
        platform=message.platform,
    )
    return agent_response
