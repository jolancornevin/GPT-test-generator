import sys
from openai import OpenAI

client = OpenAI(api_key='sk-Ac0deBG16zkaZUGeXXXXX')


def call_chatgpt(message):
    completion = client.chat.completions.create(
        model="gpt-3.5-turbo",
        messages=[
            {"role": "system", "content": "You are a professional computer scientist. Based on the user input, generate code that follow the concepts in the examples."},
            {"role": "user", "content": message}
        ],
    )

    print(completion)
    return completion.choices[0].message.content


def generate_test(code_example, test_example, code_to_test, target):
    with open(code_example) as code_example_f:
        with open(test_example) as test_example_f:
            with open(code_to_test) as code_to_test_f:
                with open(target, 'w') as target_f:
                    target_f.write(
                        call_chatgpt(
                            f"""here's an example of code:\n {code_example_f.read()}\n
                            here's an example of test for this code:\n {test_example_f.read()}\n
                            write me some tests looking like the previous example for this code:\n {code_to_test_f.read()}
                            """
                        )
                    )

def main():
    if len(sys.argv) != 5:
        print("you must pass code, test and code")
        return
    
    generate_test(sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4])

if __name__ == '__main__':
    # Execute when the module is not initialized from an import statement.
    main()
    