.PHONY: generate lint check

## Regenerate Go + TypeScript bindings
generate:
	buf generate

## Lint proto definitions
lint:
	buf lint

## Regenerate and fail if working tree is dirty
check:
	buf generate
	git diff --exit-code
