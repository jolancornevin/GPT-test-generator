import asyncio
import sys
import os
import re

from openai import OpenAI
from transformers import AutoTokenizer


def call_chatgpt(message):
    from openai import OpenAI
    client = OpenAI(api_key=os.environ['CHATGPT_KEY'])

    completion = client.chat.completions.create(
        model="gpt-3.5-turbo",
        messages=[
            {"role": "system", "content": "You are a professional computer scientist. Based on the user input, generate code that follow the concepts in the examples."},
            {"role": "user", "content": message}
        ],
    )

    print(completion)
    return completion.choices[0].message.content


def hg_api_mistral_inference(system_instruct, message, target_f, max_tokens):
    def query(system_instruct, message, max_tokens):
        print(">>> querying")
        client = OpenAI(
            base_url=os.environ.get("OPENAI_API_URL"),
            api_key=os.environ.get("OPENAI_API_KEY")
        )

        response = client.chat.completions.create(
            model="tgi",
            messages=[
                {
                    "role": "system",
                    "content": system_instruct
                },
                {
                    "role": "user",
                    "content": message
                }
            ],
            max_tokens=max_tokens,
            stream=True,
            stop=["```\n"] # tell the model to stop once we have the code
        )

        content = ''

        for chunk in response:
            content += chunk.choices[0].delta.content

        try:
            # capture the code within the ``` ``` and return it so we ignore any explanation given by the model
            return re.findall(r'```go\n(.*)```', content, flags=re.MULTILINE|re.DOTALL)[0]
        except IndexError:
            return content

    text = query(system_instruct, message, max_tokens)
    target_f.write(text)
    return text


def mistral_token_count(message):
    tokenizer = AutoTokenizer.from_pretrained("deepseek-ai/deepseek-coder-6.7b-instruct")
    # Input text for which you want to count tokens
    input_text = message

    # Tokenize the input text
    tokens = tokenizer.encode_plus(input_text, return_tensors="pt")

    # Get the token count
    token_count = tokens['input_ids'].size(1)

    print(f'Token count for input text: {token_count}')
    return token_count


def generate_test(codeType, code_to_test, target):
    code_example_path = f"./pkg/{codeType}/code.go"
    test_example_path = f"./pkg/{codeType}/test.go"

    with open(code_example_path, encoding='UTF-8') as code_example_f:
        with open(test_example_path, encoding='UTF-8') as test_example_f:
            with open(code_to_test, encoding='UTF-8') as code_to_test_f:
                with open(target, 'w', encoding='UTF-8') as target_f:
                    print(">> starting generation for " + target)
                    system_instruct = f"""
                        You are a professional programmer and expert in the golang language.
                        I'm going to give you an example of code and associated tests.

                        example of code:
                        ```go
                        {code_example_f.read()}
                        ```

                        example of tests for the code:
                        ```go
                        {test_example_f.read()}
                        ```
                    """
                    message = f"""
                        Generate me test for this code.
                        It's very important to me that you copy the iteration over the flagTest array.

                        ```go
                        {code_to_test_f.read()}
                        ```
                    """
                    # call_chatgpt(system_instruct, message, target_f, 7800 - mistral_token_count(system_instruct + message))
                    hg_api_mistral_inference(system_instruct, message, target_f, 7800 - mistral_token_count(system_instruct + message))

    print(">> done")

async def main():
    print(">> starting")
    for codeType in ["services", "handlers", "dao"]: # , "handlers"
        directory = sys.argv[1] + "/" + codeType

        calls = []

        for file in os.listdir(directory):
            filename = os.fsdecode(file)
            test_filename = filename.split(".")[0]+ "_test.go"

            if (
                filename.endswith(".go")
                # don't test tests
                and not filename.endswith("_test.go")
                # don't test wire
                and not filename.endswith("_wire.go")
                and not filename.endswith("wire_gen.go")

                # other files
                and not filename.endswith("utils.go")
                and not filename.endswith("healthcheck.go")
                and not filename.endswith("healthcheck_handler.go")

                # don't override existing tests
                and not os.path.exists(os.path.join(directory, test_filename))
            ):
                calls.append(
                    (
                        codeType,
                        os.path.join(directory, filename),
                        os.path.join(directory, test_filename),
                    )
                )

        print(f">> running for {len(calls)}")
        await asyncio.gather(
            *[asyncio.to_thread(generate_test, *args) for args in calls]
        )

if __name__ == '__main__':
    asyncio.run(main())
