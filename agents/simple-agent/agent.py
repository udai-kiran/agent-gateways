import asyncio

from strands import Agent
from strands.models.openai import OpenAIModel

from models import AgentResponse

SIMPLE_SYSTEM_PROMPT = """
You are a simple agent that can answer questions and help with tasks.
"""


def create_agent(
    api_key: str, model_id: str = "gpt-5.4-nano-2026-03-17", history: list[dict] | None = None
) -> Agent:
    """Create a simple agent with optional conversation history."""
    model = OpenAIModel(model_id=model_id, client_args={"api_key": api_key})
    return Agent(
        system_prompt=SIMPLE_SYSTEM_PROMPT,
        model=model,
        messages=history or [],
    )


async def run(agent: Agent, text: str) -> AgentResponse:
    """Run the agent in a thread and return the response."""
    result = await asyncio.to_thread(agent, text)
    return AgentResponse(text=str(result))
