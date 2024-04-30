run:
	. .venv/bin/activate
	export $(grep -v '^#' .env | xargs); TOKENIZERS_PARALLELISM=true python3.12 ./mistal_api.py /Users/jcornevin/go/src/github.com/Work4Labs/uservice-scoring/pkg