

When parsing a tag, you'll see the following patterns:

An `<html>` tag:
- 0: startopentag: `<html`
- 1: endopentag: `<html>`

A `<head> and `</ head>` set of tags:
- 0: startopentag: <head
- 1: endopentag: <head>
- 2: closetag: </head>

A `<body style="font-size:10em;">` tag:
- 0: startopentag: <body
- 1: id: style
- 2: assign: =
- 3: string: "font-size:10em;"
- 4: endopentag: <body>

A `<meta id="foobar"/>` tag:
- 0: startopentag: <meta
- 1: id: header
- 2: assign: =
- 3: string: "foobar"
- 4: endselfclosingtag: <meta/>
