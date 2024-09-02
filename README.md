# GPT test generator

A small project to generate tests in golang, following example in the /pkg folder.

https://huggingface.co/spaces/bigcode/bigcode-models-leaderboard to chose your model (or you can use chatgpt).

You need to export
- OPENAI_API_KEY
- OPENAI_API_URL

or CHATGPT_KEY


Then run `export $(grep -v '^#' .env | xargs); TOKENIZERS_PARALLELISM=true python3.12 ./mistal_api.py PATH_TO_PKG_FOLDER`