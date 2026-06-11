package main

var (
	goCode = `package main

import "fmt"

func main() {
	for i := range 10 {
		if i%2 == 0 {
			fmt.Println(i, "is even")
		}
	}
}
`
	jsCode = `const express = require('express');
const app = express();

app.get('/api/users/:id', async (req, res) => {
  try {
    const user = await db.findById(req.params.id);
    res.json({ user, timestamp: Date.now() });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

app.listen(3000);
`
	htmlCode = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Test</title>
    <style>
        body { font-family: sans-serif; color: #333; }
        .highlight { background: yellow; }
    </style>
</head>
<body>
    <div id="app">
        <h1>Hello World</h1>
    </div>
</body>
</html>
`
	tsCode = `interface User {
  id: number;
  name: string;
  email?: string;
}

function greet<T extends User>(user: T): string {
  return "Hello, " + user.name;
}

const users: User[] = [{ id: 1, name: "Alice" }];
`
	markdownCode = "# Heading\n\nA paragraph with **bold**, *italic*, and `code`.\n\n- Item one\n- Item two\n\n```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```\n"
)

// Input holds a named code snippet with its language identifier.
type Input struct {
	Name string
	Lang string
	Code string
}

var defaultInputs = []Input{
	{"Go", "go", goCode},
	{"JavaScript", "javascript", jsCode},
	{"HTML", "html", htmlCode},
	{"TypeScript", "typescript", tsCode},
	{"Markdown", "markdown", markdownCode},
}
