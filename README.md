# yze-go-jsontag

A [`yze`](https://github.com/gomatic/yze) analyzer (category `data`) that reports struct field `json`/`yaml` tag keys that are not `snake_case`, per the gomatic data-format standard that serialized keys are `snake_case`.

- **Rule:** `yze/jsontag`
- **Library:** exports `Analyzer` (a standard `go/analysis` analyzer) and `Registration` for the [`yze`](https://github.com/gomatic/yze) aggregator and [`stickler`](https://github.com/gomatic/stickler) runner.
- **Binary:** `cmd/yze-go-jsontag` runs it standalone (`text`/`-json`, and as a `go vet -vettool`).

Built on the [`go-yze`](https://github.com/gomatic/go-yze) framework.
